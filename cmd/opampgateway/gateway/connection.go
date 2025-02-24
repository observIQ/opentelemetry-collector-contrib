package gateway

import (
	"context"

	clientTypes "github.com/open-telemetry/opamp-go/client/types"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opamp-go/server/types"
)

type Connection struct {
	ID      clientTypes.InstanceUid
	Gateway *Gateway
}

func NewConnection(gateway *Gateway) *Connection {
	return &Connection{Gateway: gateway}
}

func (c *Connection) OnConnected(ctx context.Context, conn types.Connection) {
	c.Gateway.handleClientConnected(c.ID, conn)
}

func (c *Connection) OnMessage(ctx context.Context, conn types.Connection, msg *protobufs.AgentToServer) *protobufs.ServerToAgent {
	c.ID = clientTypes.InstanceUid(msg.InstanceUid)
	return c.Gateway.handleClientMessage(conn, msg)
}

func (c *Connection) OnConnectionClose(_ types.Connection) {
	c.Gateway.handleClientConnectionClose(c.ID)
}

func (c *Connection) toConnectionCallbacks() types.ConnectionCallbacks {
	return types.ConnectionCallbacks{
		OnMessage:         c.OnMessage,
		OnConnectionClose: c.OnConnectionClose,
		OnConnected:       c.OnConnected,
	}
}
