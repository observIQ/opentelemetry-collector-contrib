package observiqreceiver

import (
	"context"
	"sync"
	"time"

	observiq "github.com/observiq/carbon/agent"
	observiqEntry "github.com/observiq/carbon/entry"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.uber.org/zap"
)

type observiqReceiver struct {
	startOnce sync.Once
	stopOnce  sync.Once
	done      chan struct{}

	config   *Config
	agent    *observiq.LogAgent
	consumer consumer.LogsConsumer
	logger   *zap.Logger
	logsChan chan observiqEntry.Entry
}

// Ensure this factory adheres to required interface
var _ component.LogsReceiver = (*observiqReceiver)(nil)

// Start tells the receiver to start
func (r *observiqReceiver) Start(ctx context.Context, host component.Host) error {
	logAgent := observiq.NewLogAgent(r.config.agentConfig, r.logger.Sugar(), "todo", "todo").
		WithBuildParameter("otc_output_chan", r.logsChan)

	err := componenterror.ErrAlreadyStarted
	r.startOnce.Do(func() {
		err = nil

		r.agent.Start()

		go func() {

			for {
				select {
				case al := <-r.logsChan:

					if err := r.consumer.ConsumeLogs(ctx, ld); err != nil {
						// TODO
					}
				case <-r.done:
					return
				}
			}
		}()
	})

	return err
}

// Shutdown is invoked during service shutdown
func (r *observiqReceiver) Shutdown(context.Context) error {
	r.stopOnce.Do(func() {
		close(r.done)
	})
	return nil
}
