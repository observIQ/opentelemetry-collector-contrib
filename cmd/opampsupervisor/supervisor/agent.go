package supervisor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/opampsupervisor/supervisor/commander"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/opampsupervisor/supervisor/config"
	"go.uber.org/zap"
)

const (
	agentConfigFilePath = "effective.yaml"
)

type agent interface {
	// Start starts the agent
	Start() error
	// Stop stops the agent
	Stop() error
	// AgentDescription returns an AgentDescription that describes the current agent
	AgentDescription() *protobufs.AgentDescription
	// Reconfigure reconfigures the agent with the new remote config
	Reconfigure(*protobufs.AgentRemoteConfig) error
	// Restart restarts the agent process
	Restart() error
}

type opentelemetryCollectorAgent struct {
	logger *zap.Logger
	cfg    config.Agent

	remoteConfig *protobufs.AgentRemoteConfig

	configChan            chan []byte
	currentRenderedConfig []byte

	healthCheckPort int

	wg       *sync.WaitGroup
	doneChan chan struct{}
}

func newOpentelemetryCollectorAgent(
	logger *zap.Logger,
	agentCfg config.Agent,
	initialConfig *protobufs.AgentRemoteConfig,
) (opentelemetryCollectorAgent, error) {

	renderedConfig, err := renderConfig(initialConfig)
	if err != nil {
		return opentelemetryCollectorAgent{}, fmt.Errorf("failed to render initial config: %w", err)
	}

	return opentelemetryCollectorAgent{
		logger:                logger,
		cfg:                   agentCfg,
		remoteConfig:          initialConfig,
		currentRenderedConfig: renderedConfig,
		wg:                    &sync.WaitGroup{},
	}
}

func (a opentelemetryCollectorAgent) Start() error {
	// TODO: determine Health check port
	// TODO: Write out initial config
	agentCommand, err := commander.NewCommander(
		a.logger,
		a.cfg,
		"--config", agentConfigFilePath,
	)
	if err != nil {
		return fmt.Errorf("failed to create agent command: %w", err)
	}

	err = agentCommand.Start(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.agentLoop(agentCommand)
	}()

	return nil
}

func (a opentelemetryCollectorAgent) Stop() {
	close(a.doneChan)
	a.wg.Wait()
}
func (opentelemetryCollectorAgent) Reconfigure(*protobufs.AgentRemoteConfig) error { return nil }
func (opentelemetryCollectorAgent) Restart()                                       {}

func (a opentelemetryCollectorAgent) agentLoop(agentCommand *commander.Commander) {
	healthCheckBackoff := backoff.NewExponentialBackOff()
	healthCheckBackoff.MaxInterval = 60 * time.Second
	healthCheckBackoff.MaxElapsedTime = 0 // Never stop
	healthCheckTicker := backoff.NewTicker(healthCheckBackoff)
	defer healthCheckTicker.Stop()

	restartTimer := time.NewTimer(0)
	restartTimer.Stop()

	for {
		select {
		case _ = <-a.configChan:
			a.logger.Debug("Restarting agent due to new config")
			// TODO Replace config, restart agent

		case <-agentCommand.Exited():
			a.logger.Debug("Agent process exited unexpectedly. Will restart in a bit...", zap.Int("pid", agentCommand.Pid()), zap.Int("exit_code", agentCommand.ExitCode()))

		case <-restartTimer.C:
			a.logger.Debug("Agent starting after start backoff")

		case <-healthCheckTicker.C:
			a.logger.Debug("Performing agent health check")

		case <-a.doneChan:
			a.logger.Debug("Shutting down agent...")

			err := agentCommand.Stop(context.Background())
			if err != nil {
				a.logger.Error("Could not stop agent process", zap.Error(err))
			}
			return
		}
	}
}

func renderConfig(*protobufs.AgentRemoteConfig) ([]byte, error) {
	return nil, nil
}
