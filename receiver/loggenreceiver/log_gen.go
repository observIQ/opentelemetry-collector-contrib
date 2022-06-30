package loggenreceiver

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type logGenReceiver struct {
	logsPerSec   int
	logLine      *string
	wg           sync.WaitGroup
	logger       *zap.Logger
	emitInterval time.Duration
	emitLogNum   int
	doneChan     chan struct{}

	logConsumer consumer.Logs
}

func newReceiver(
	set component.ReceiverCreateSettings,
	config *Config,
	consumer consumer.Logs,
) (component.LogsReceiver, error) {

	// Calculate how many logs to generate per millisecond
	logsPerMs := config.LogsPerSec / 1000

	return &logGenReceiver{
		logsPerSec:   config.LogsPerSec,
		logLine:      &config.LogLine,
		logger:       set.Logger,
		emitInterval: config.EmitInterval,
		emitLogNum:   logsPerMs * int(config.EmitInterval.Milliseconds()),
		doneChan:     make(chan struct{}),

		logConsumer: consumer,
	}, nil
}

func (l *logGenReceiver) Start(ctx context.Context, _ component.Host) error {
	l.wg.Add(1)

	go l.generator(ctx)

	return nil
}
func (l *logGenReceiver) Shutdown(ctx context.Context) error {
	close(l.doneChan)
	l.wg.Done()
	return nil
}

func (l *logGenReceiver) generator(ctx context.Context) {
	defer l.wg.Done()

	ticker := time.NewTicker(l.emitInterval)
	defer ticker.Stop()

	ld := plog.NewLogs()
	sl := ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Ensure we have generated enough logs for the emit Interval
			for ld.LogRecordCount() < l.emitLogNum {
				lr := sl.LogRecords().AppendEmpty()
				lr.Body().SetStringVal(*l.logLine)
				lr.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
			}

			l.logConsumer.ConsumeLogs(ctx, ld)

			ld = plog.NewLogs()
			sl = ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()
		default:
			// Generate log entries until we have enough for emit log num
			if ld.LogRecordCount() < l.emitLogNum {
				lr := sl.LogRecords().AppendEmpty()
				lr.Body().SetStringVal(*l.logLine)
				lr.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
			}
		}
	}
}
