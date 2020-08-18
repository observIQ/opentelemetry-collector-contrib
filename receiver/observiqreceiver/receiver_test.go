package observiqreceiver

import (
	"context"
	"testing"

	"github.com/observiq/carbon/entry"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"

	"go.uber.org/zap"
)

func TestStart(t *testing.T) {
	factory := &Factory{}
	params := component.ReceiverCreateParams{
		Logger: zap.NewNop(),
	}
	mockConsumer := mockLogsConsumer{}
	receiver, _ := factory.CreateLogsReceiver(context.Background(), params, factory.CreateDefaultConfig(), &mockConsumer)

	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err, "receiver start failed")

	obsReceiver := receiver.(*observiqReceiver)
	obsReceiver.logsChan <- entry.New()
	receiver.Shutdown(context.Background())
	require.Equal(t, 1, mockConsumer.received, "one log entry expected")
}

func TestHandleStartError(t *testing.T) {
	factory := &Factory{}
	params := component.ReceiverCreateParams{
		Logger: zap.NewNop(),
	}
	mockConsumer := mockLogsConsumer{}

	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Pipeline = append(cfg.Pipeline, newUnstartableParams())

	receiver, err := factory.CreateLogsReceiver(context.Background(), params, cfg, &mockConsumer)
	require.NoError(t, err, "receiver should successfully build")

	err = receiver.Start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err, "receiver fails to start under rare circumstances")
}

func TestHandleConsumeError(t *testing.T) {
	factory := &Factory{}
	params := component.ReceiverCreateParams{
		Logger: zap.NewNop(),
	}
	mockConsumer := mockLogsRejecter{}
	receiver, _ := factory.CreateLogsReceiver(context.Background(), params, factory.CreateDefaultConfig(), &mockConsumer)

	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err, "receiver start failed")

	obsReceiver := receiver.(*observiqReceiver)
	obsReceiver.logsChan <- entry.New()
	receiver.Shutdown(context.Background())
	require.Equal(t, 1, mockConsumer.rejected, "one log entry expected")
}
