package observiqreceiver

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.uber.org/zap"
)

func TestDefaultConfig(t *testing.T) {
	factory := &Factory{}
	cfg := factory.CreateDefaultConfig()
	require.Equal(t, factory.Type(), cfg.Type())
	require.NotNil(t, cfg, "failed to create default config")
	require.NoError(t, configcheck.ValidateConfig(cfg))
}

func TestCreateReceiver(t *testing.T) {
	factory := &Factory{}
	params := component.ReceiverCreateParams{
		Logger: zap.NewNop(),
	}
	receiver, err := factory.CreateLogsReceiver(context.Background(), params, factory.CreateDefaultConfig(), &mockLogsConsumer{})
	require.NoError(t, err, "receiver creation failed")
	require.NotNil(t, receiver, "receiver creation failed")

	badCfg := factory.CreateDefaultConfig().(*Config)
	badCfg.OffsetsFile = os.Args[0] // current executable cannot be opened
	receiver, err = factory.CreateLogsReceiver(context.Background(), params, badCfg, &mockLogsConsumer{})
	require.Error(t, err, "receiver creation should fail if offsets file is invalid")
	require.Nil(t, receiver, "receiver creation should have failed due to invalid offsets file")
}
