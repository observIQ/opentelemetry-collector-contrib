package observiqreceiver

import (
	"github.com/observiq/carbon/agent"
	"go.opentelemetry.io/collector/config/configmodels"
)

// Config defines configuration for the observiq receiver
type Config struct {
	configmodels.ReceiverSettings `mapstructure:",squash"`
	agentConfig                   *agent.Config // TODO mapstructure?
}
