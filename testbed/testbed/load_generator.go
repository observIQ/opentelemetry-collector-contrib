// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package testbed // import "github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed"

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/collector/consumer/consumererror"
	"golang.org/x/text/message"
)

var printer = message.NewPrinter(message.MatchLanguage("en"))

// LoadGenerator is a simple load generator.
type LoadGenerator struct {
	sender DataSender

	dataProvider DataProvider

	// Number of data items (spans or metric data points) sent.
	dataItemsSent atomic.Uint64

	// Number of permanent errors received
	permanentErrors    atomic.Uint64
	nonPermanentErrors atomic.Uint64

	stopOnce   sync.Once
	stopWait   sync.WaitGroup
	stopSignal chan struct{}

	options LoadOptions

	// Record information about previous errors to avoid flood of error messages.
	prevErr error
}

// LoadOptions defines the options to use for generating the load.
type LoadOptions struct {
	// DataItemsPerSecond specifies how many spans, metric data points, or log
	// records to generate each second.
	DataItemsPerSecond int

	// ItemsPerBatch specifies how many spans, metric data points, or log
	// records per batch to generate. Should be greater than zero. The number
	// of batches generated per second will be DataItemsPerSecond/ItemsPerBatch.
	ItemsPerBatch int

	// Attributes to add to each generated data item. Can be empty.
	Attributes map[string]string

	// Parallel specifies how many goroutines to send from.
	Parallel int
}

// NewLoadGenerator creates a load generator that sends data using specified sender.
func NewLoadGenerator(dataProvider DataProvider, sender DataSender) (*LoadGenerator, error) {
	if sender == nil {
		return nil, fmt.Errorf("cannot create load generator without DataSender")
	}

	lg := &LoadGenerator{
		stopSignal:   make(chan struct{}),
		sender:       sender,
		dataProvider: dataProvider,
	}

	return lg, nil
}

// Start the load.
func (lg *LoadGenerator) Start(options LoadOptions) {
	lg.options = options

	if lg.options.ItemsPerBatch == 0 {
		// 10 items per batch by default.
		lg.options.ItemsPerBatch = 10
	}

	log.Printf("Starting load generator at %d items/sec.", lg.options.DataItemsPerSecond)

	// Indicate that generation is in progress.
	lg.stopWait.Add(1)

	// Begin generation
	go lg.generate()
}

// Stop the load.
func (lg *LoadGenerator) Stop() {
	lg.stopOnce.Do(func() {
		// Signal generate() to stop.
		close(lg.stopSignal)

		// Wait for it to stop.
		lg.stopWait.Wait()

		// Print stats.
		log.Printf("Stopped generator. %s", lg.GetStats())
	})
}

// GetStats returns the stats as a printable string.
func (lg *LoadGenerator) GetStats() string {
	return fmt.Sprintf("Sent:%10d items", lg.DataItemsSent())
}

func (lg *LoadGenerator) DataItemsSent() uint64 {
	return lg.dataItemsSent.Load()
}

func (lg *LoadGenerator) PermanentErrors() uint64 {
	return lg.permanentErrors.Load()
}

func (lg *LoadGenerator) NonPermanentErrors() uint64 {
	return lg.nonPermanentErrors.Load()
}

// IncDataItemsSent is used when a test bypasses the LoadGenerator and sends data
// directly via TestCases's Sender. This is necessary so that the total number of sent
// items in the end is correct, because the reports are printed from LoadGenerator's
// fields. This is not the best way, a better approach would be to refactor the
// reports to use their own counter and load generator and other sending sources
// to contribute to this counter. This could be done as a future improvement.
func (lg *LoadGenerator) IncDataItemsSent() {
	lg.dataItemsSent.Add(1)
}

func (lg *LoadGenerator) generate() {
	// Indicate that generation is done at the end
	defer lg.stopWait.Done()

	if lg.options.DataItemsPerSecond == 0 {
		return
	}

	lg.dataProvider.SetLoadGeneratorCounters(&lg.dataItemsSent)

	err := lg.sender.Start()
	if err != nil {
		log.Printf("Cannot start sender: %v", err)
		return
	}

	numWorkers := 1

	if lg.options.Parallel > 0 {
		numWorkers = lg.options.Parallel
	}

	var workers sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		workers.Add(1)

		go func() {
			defer workers.Done()
			t := time.NewTicker(time.Second / time.Duration(lg.options.DataItemsPerSecond/lg.options.ItemsPerBatch/numWorkers))
			defer t.Stop()
			for {
				select {
				case <-t.C:
					switch lg.sender.(type) {
					case TraceDataSender:
						lg.generateTrace()
					case MetricDataSender:
						lg.generateMetrics()
					case LogDataSender:
						lg.generateLog()
					default:
						log.Printf("Invalid type of LoadGenerator sender")
					}
				case <-lg.stopSignal:
					return
				}
			}
		}()
	}

	workers.Wait()

	// Send all pending generated data.
	lg.sender.Flush()
}

func (lg *LoadGenerator) generateTrace() {
	traceSender := lg.sender.(TraceDataSender)

	traceData, done := lg.dataProvider.GenerateTraces()
	if done {
		return
	}

	for {
		err := traceSender.ConsumeTraces(context.Background(), traceData)
		if err == nil {
			lg.prevErr = nil
			break
		}

		if !consumererror.IsPermanent(err) {
			lg.nonPermanentErrors.Add(uint64(traceData.SpanCount()))
			continue
		}

		lg.permanentErrors.Add(uint64(traceData.SpanCount()))

		// update prevErr to err if it's different than last observed error
		if lg.prevErr == nil || lg.prevErr.Error() != err.Error() {
			lg.prevErr = err
			log.Printf("Cannot send traces: %v", err)
		}
		break
	}
}

func (lg *LoadGenerator) generateMetrics() {
	metricSender := lg.sender.(MetricDataSender)

	metricData, done := lg.dataProvider.GenerateMetrics()
	if done {
		return
	}
	for {
		err := metricSender.ConsumeMetrics(context.Background(), metricData)
		if err == nil {
			lg.prevErr = nil
			break
		}

		if !consumererror.IsPermanent(err) {
			lg.nonPermanentErrors.Add(uint64(metricData.DataPointCount()))
			continue
		}

		lg.permanentErrors.Add(uint64(metricData.DataPointCount()))

		// update prevErr to err if it's different than last observed error
		if lg.prevErr == nil || lg.prevErr.Error() != err.Error() {
			lg.prevErr = err
			log.Printf("Cannot send metrics: %v", err)
		}
		break
	}
}

func (lg *LoadGenerator) generateLog() {
	logSender := lg.sender.(LogDataSender)

	logData, done := lg.dataProvider.GenerateLogs()
	if done {
		return
	}
	for {
		err := logSender.ConsumeLogs(context.Background(), logData)
		if err == nil {
			lg.prevErr = nil
			break
		}

		if !consumererror.IsPermanent(err) {
			lg.nonPermanentErrors.Add(uint64(logData.LogRecordCount()))
			continue
		}

		lg.permanentErrors.Add(uint64(logData.LogRecordCount()))

		// update prevErr to err if it's different than last observed error
		if lg.prevErr == nil || lg.prevErr.Error() != err.Error() {
			lg.prevErr = err
			log.Printf("Cannot send logs: %v", err)
		}
		break
	}
}
