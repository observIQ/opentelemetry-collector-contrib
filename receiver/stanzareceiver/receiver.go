// Copyright 2019, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stanzareceiver

import (
	"context"
	"fmt"
	"sync"
	"time"

	stanza "github.com/observiq/stanza/agent"
	"github.com/observiq/stanza/entry"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type stanzareceiver struct {
	sync.Mutex
	startOnce sync.Once
	stopOnce  sync.Once
	wg        sync.WaitGroup
	cancel    context.CancelFunc

	maxBatchSize int
	maxBatchTime time.Duration

	agent    *stanza.LogAgent
	emitter  *LogEmitter
	consumer consumer.LogsConsumer
	logger   *zap.Logger
}

// Ensure this receiver adheres to required interface
var _ component.LogsReceiver = (*stanzareceiver)(nil)

// Start tells the receiver to start
func (r *stanzareceiver) Start(ctx context.Context, host component.Host) error {
	r.Lock()
	defer r.Unlock()
	err := componenterror.ErrAlreadyStarted
	r.startOnce.Do(func() {
		err = nil
		rctx, cancel := context.WithCancel(ctx)
		r.cancel = cancel
		r.logger.Info("Starting stanza receiver")

		if obsErr := r.agent.Start(); obsErr != nil {
			err = fmt.Errorf("start stanza: %s", err)
			return
		}

		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			for {
				select {
				case <-rctx.Done():
					return
				case ent, ok := <-r.emitter.logChan:
					if !ok {
						break
					}
					if consumeErr := r.consumer.ConsumeLogs(ctx, convert(r.drainBatch(ctx, ent)...)); consumeErr != nil {
						r.logger.Error("ConsumeLogs() error", zap.String("error", consumeErr.Error()))
					}
				}
			}
		}()
	})

	return err
}

func (r *stanzareceiver) drainBatch(ctx context.Context, first *entry.Entry) []*entry.Entry {
	entries := []*entry.Entry{first}

	batchTimeout := time.After(r.maxBatchTime)
	for {
		select {
		case <-ctx.Done():
			return entries
		case <-batchTimeout:
			return entries
		case e, ok := <-r.emitter.logChan:
			if !ok {
				return entries
			}
			entries = append(entries, e)
			if len(entries) == r.maxBatchSize {
				return entries
			}
		default: // channel empty
			return entries
		}
	}
}

// Shutdown is invoked during service shutdown
func (r *stanzareceiver) Shutdown(context.Context) error {
	r.Lock()
	defer r.Unlock()

	err := componenterror.ErrAlreadyStopped
	r.stopOnce.Do(func() {
		r.logger.Info("Stopping stanza receiver")
		err = r.agent.Stop()
		r.cancel()
		r.wg.Wait()
	})
	return err
}
