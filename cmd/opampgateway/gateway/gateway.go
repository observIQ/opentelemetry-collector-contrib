package gateway

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/open-telemetry/opamp-go/client"
	clientTypes "github.com/open-telemetry/opamp-go/client/types"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opamp-go/server"
	serverTypes "github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"
)

const (
	headerAgentID = "Agent-Id"
)

type Gateway struct {
	logger *zap.Logger
	config *Config

	client client.OpAMPClient
	server server.OpAMPServer

	connectionsMutex sync.Mutex
	connections      map[clientTypes.InstanceUid]*Agent
	count            int

	awaitConnections sync.WaitGroup
}

type Agent struct {
	AgentDescription     *protobufs.AgentDescription
	Capabilities         uint64
	EffectiveConfig      *protobufs.EffectiveConfig
	InstanceUid          clientTypes.InstanceUid
	FirstMessageReceived bool
	Connection           serverTypes.Connection
}

func NewGateway(logger *zap.Logger, cfg *Config) *Gateway {
	return &Gateway{
		logger:           logger,
		config:           cfg,
		connectionsMutex: sync.Mutex{},
		connections:      make(map[clientTypes.InstanceUid]*Agent),
		awaitConnections: sync.WaitGroup{},
		count:            0,
	}
}

func (g *Gateway) Start(ctx context.Context) error {
	// Start the gateway server
	if err := g.startGatewayServer(); err != nil {
		return fmt.Errorf("failed to start gateway server: %w", err)
	}

	cycles := 0
	for {
		if g.connectionsReady() {
			break
		}
		time.Sleep(3 * time.Second)
		cycles++
		if cycles > 10 {
			return fmt.Errorf("timed out waiting for connections to be ready")
		}
	}

	// Start the upstream client
	if err := g.startGatewayClient(); err != nil {
		return fmt.Errorf("failed to start gateway client: %w", err)
	}

	return nil
}

func (g *Gateway) Stop(ctx context.Context) error {
	return nil
}

func (g *Gateway) connectionsReady() bool {
	defer g.connectionsMutex.Unlock()
	g.connectionsMutex.Lock()

	ready := true
	for _, a := range g.connections {
		if !a.FirstMessageReceived {
			ready = false
		}
	}
	return ready
}

// Server Implementation --------------------------------------------------------

func (g *Gateway) startGatewayServer() error {
	g.server = server.New(newLoggerFromZap(g.logger, "opamp-server"))

	serverSettings := server.StartSettings{
		Settings: server.Settings{
			Callbacks: g.toCallbacks(),
		},
		ListenEndpoint: g.config.GatewayListenAddress,
	}

	g.awaitConnections.Add(g.config.ConnectionLimit)

	if err := g.server.Start(serverSettings); err != nil {
		return fmt.Errorf("failed to start gateway server: %w", err)
	}

	g.logger.Info("Gateway server started", zap.String("listen_address", g.config.GatewayListenAddress))
	g.logger.Info("Waiting for connections", zap.Int("connection_limit", g.config.ConnectionLimit))

	g.awaitConnections.Wait()

	g.logger.Info("All connections started")

	return nil
}

func (g *Gateway) toCallbacks() serverTypes.Callbacks {
	return serverTypes.Callbacks{
		OnConnecting: g.OnConnecting,
	}
}

func (g *Gateway) OnConnecting(request *http.Request) serverTypes.ConnectionResponse {
	shouldConnect, rejectStatusCode := g.handleClientConnecting(request)
	return serverTypes.ConnectionResponse{
		Accept:              shouldConnect,
		HTTPStatusCode:      rejectStatusCode,
		ConnectionCallbacks: NewConnection(g).toConnectionCallbacks(),
	}
}

// handleClientConnect handles a new client connection to the gateway.
// It checks if the connection limit has been reached and if the client ID is present.
// If the connection limit has been reached or the client ID is not present, it returns a 429 error.
// Otherwise, it adds the client ID to the list of connections and returns a 200 OK.
func (g *Gateway) handleClientConnecting(_ *http.Request) (bool, int) {
	defer g.connectionsMutex.Unlock()
	g.connectionsMutex.Lock()

	if g.count >= g.config.ConnectionLimit {
		return false, http.StatusTooManyRequests
	}

	return true, http.StatusOK
}

// handleClientConnected handles a client connection event.
// It adds the client ID to the list of connections.
func (g *Gateway) handleClientConnected(id clientTypes.InstanceUid, conn serverTypes.Connection) {}

// handleClientConnectionClose handles a client connection close event.
// It removes the client ID from the list of connections.
func (g *Gateway) handleClientConnectionClose(id clientTypes.InstanceUid) {
	defer g.connectionsMutex.Unlock()
	g.connectionsMutex.Lock()
	delete(g.connections, id)
}

func (g *Gateway) handleClientMessage(conn serverTypes.Connection, msg *protobufs.AgentToServer) *protobufs.ServerToAgent {
	defer g.connectionsMutex.Unlock()
	g.connectionsMutex.Lock()

	a, ok := g.connections[clientTypes.InstanceUid(msg.InstanceUid)]
	if !ok {
		a = &Agent{
			FirstMessageReceived: false,
			Connection:           conn,
			InstanceUid:          clientTypes.InstanceUid(msg.InstanceUid),
			AgentDescription:     msg.AgentDescription,
			Capabilities:         msg.Capabilities,
		}
	}

	g.logger.Info("Received message from client", zap.String("agent_id", string(msg.InstanceUid)))
	if msg.AgentDescription != nil {
		g.logger.Info("Received agent description from client", zap.String("agent_id", string(msg.InstanceUid)))

		a.AgentDescription = msg.AgentDescription
		if a.FirstMessageReceived {
			g.client.SetAgentDescription(a.InstanceUid, msg.AgentDescription)
		}
	}
	if msg.EffectiveConfig != nil {
		g.logger.Info("Received effective config from client", zap.String("agent_id", string(msg.InstanceUid)))

		a.EffectiveConfig = msg.EffectiveConfig
		if a.FirstMessageReceived {
			g.client.UpdateEffectiveConfig(a.InstanceUid, context.Background())
		}
	}
	if msg.CustomMessage != nil {
		g.logger.Info("Received custom message from client", zap.String("agent_id", string(msg.InstanceUid)))
		if a.FirstMessageReceived {
			g.client.SendCustomMessage(a.InstanceUid, msg.CustomMessage)
		}
	}
	if msg.Health != nil {
		g.logger.Info("Received health status from client", zap.String("agent_id", string(msg.InstanceUid)))
		if a.FirstMessageReceived {
			g.client.SetHealth(a.InstanceUid, msg.Health)
		}
	}
	a.FirstMessageReceived = true

	return nil
}

// Client Implementation --------------------------------------------------------

func (g *Gateway) startGatewayClient() error {
	g.client = client.NewWebSocket(newLoggerFromZap(g.logger, "opamp-client"))

	agents := g.prepareAgents()

	settings := clientTypes.StartSettings{
		OpAMPServerURL: g.config.ServerEndpoint,
		Header:         g.config.ServerHeaders,
		Agents:         agents,
		Callbacks: clientTypes.Callbacks{
			OnConnect: func(_ context.Context) {
				g.logger.Info("Connected to the server")
			},
			OnConnectFailed: func(_ context.Context, err error) {
				g.logger.Error("Failed to connect to the server", zap.Error(err))
			},
			OnError: func(_ context.Context, err *protobufs.ServerErrorResponse) {
				g.logger.Error("Server returned an error response", zap.String("message", err.ErrorMessage))
			},
			OnMessage: func(_ context.Context, msg *clientTypes.MessageData) {
				g.handleServerMessage(msg)
			},
		},
	}

	if err := g.client.PrepareStart(context.Background(), settings); err != nil {
		return fmt.Errorf("failed to prepare gateway client: %w", err)
	}
	if err := g.client.Start(context.Background(), settings); err != nil {
		return fmt.Errorf("failed to start gateway client: %w", err)
	}

	g.logger.Info("Gateway client started")

	g.sendDescriptions()

	return nil
}

func (g *Gateway) prepareAgents() []*clientTypes.Agent {
	agents := make([]*clientTypes.Agent, 0, len(g.connections))
	for _, a := range g.connections {
		agents = append(agents, &clientTypes.Agent{
			InstanceUid:      a.InstanceUid,
			AgentDescription: a.AgentDescription,
			Capabilities:     protobufs.AgentCapabilities(a.Capabilities),
		})
	}
	return agents
}

func (g *Gateway) sendDescriptions() {
	for _, a := range g.connections {
		g.client.SetAgentDescription(a.InstanceUid, a.AgentDescription)
	}
}

func (g *Gateway) handleServerMessage(msg *clientTypes.MessageData) {
	g.logger.Info("Received message from the server")
	defer g.connectionsMutex.Unlock()
	g.connectionsMutex.Lock()

	id := msg.InstanceUid
	conn := g.connections[clientTypes.InstanceUid(id)].Connection

	conn.Send(context.Background(), &protobufs.ServerToAgent{
		InstanceUid:        id[:],
		CustomMessage:      msg.CustomMessage,
		CustomCapabilities: msg.CustomCapabilities,
		RemoteConfig:       msg.RemoteConfig,
		PackagesAvailable:  msg.PackagesAvailable,
	})
}

// Helpers ----------------------------------------------------------------------

func parseIDHeader(headers http.Header) (string, error) {
	id := headers.Get(headerAgentID)
	if id == "" {
		return "", fmt.Errorf("no %s header found", headerAgentID)
	}

	return id, nil
}
