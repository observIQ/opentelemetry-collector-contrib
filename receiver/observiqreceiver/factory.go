package observiqreceiver

import (
	"context"

	"github.com/observiq/carbon/agent"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
)

const (
	typeStr = "observiq"
)

// Factory is the factory for the observiq receiver.
type Factory struct {
}

// Ensure this factory adheres to required interface
var _ component.LogsReceiverFactory = (*Factory)(nil)

// Type gets the type of the Receiver config created by this factory
func (f *Factory) Type() configmodels.Type {
	return configmodels.Type(typeStr)
}

// CustomUnmarshaler returns nil because we don't need custom unmarshaling for this config
func (f *Factory) CustomUnmarshaler() component.CustomUnmarshaler {
	return nil
}

// CreateDefaultConfig creates the default configuration for the observiq receiver
func (f *Factory) CreateDefaultConfig() configmodels.Receiver {
	return &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: configmodels.Type(typeStr),
			NameVal: typeStr,
		},
		agentConfig: agent.Config{},
		// TODO additional config values
	}
}

// CreateLogsReceiver creates a logs receiver based on provided config
func (f *Factory) CreateLogsReceiver(
	ctx context.Context,
	params component.ReceiverCreateParams,
	cfg configmodels.Receiver,
	nextConsumer consumer.LogsConsumer,
) (component.LogsReceiver, error) {

	c := cfg.(*Config)
	return &observiqReceiver{
		config:   c,
		consumer: nextConsumer,
		logger:   params.Logger,
	}, nil
}
