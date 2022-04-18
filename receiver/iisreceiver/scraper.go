// Copyright The OpenTelemetry Authors
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

//go:build windows
// +build windows

package iisreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/iisreceiver"

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scrapererror"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/iisreceiver/internal/metadata"
)

type iisReceiver struct {
	params        component.ReceiverCreateSettings
	config        *Config
	consumer      consumer.Metrics
	watchers      []winperfcounters.PerfCounterWatcher
	metricBuilder *metadata.MetricsBuilder
}

// newIisReceiver returns an iisReceiver
func newIisReceiver(params component.ReceiverCreateSettings, cfg *Config, consumer consumer.Metrics) *iisReceiver {
	return &iisReceiver{params: params, config: cfg, consumer: consumer, metricBuilder: metadata.NewMetricsBuilder(cfg.Metrics)}
}

// start builds the paths to the watchers
func (rcvr *iisReceiver) start(ctx context.Context, host component.Host) error {
	rcvr.watchers = []winperfcounters.PerfCounterWatcher{}

	var errors scrapererror.ScrapeErrors
	for _, objCfg := range getScraperCfgs() {
		objWatchers, err := objCfg.BuildPaths()
		if err != nil {
			errors.AddPartial(1, fmt.Errorf("some performance counters could not be initialized; %w", err))
			continue
		}
		for _, objWatcher := range objWatchers {
			rcvr.watchers = append(rcvr.watchers, objWatcher)
		}
	}

	return errors.Combine()
}

// scrape pulls counter values from the watchers
func (rcvr *iisReceiver) scrape(ctx context.Context) (pdata.Metrics, error) {
	var errs error
	now := pdata.NewTimestampFromTime(time.Now())

	for _, watcher := range rcvr.watchers {
		counterValues, err := watcher.ScrapeData()
		if err != nil {
			rcvr.params.Logger.Warn("some performance counters could not be scraped; ", zap.Error(err))
			continue
		}
		var metricRep winperfcounters.MetricRep
		value := 0.0
		for _, counterValue := range counterValues {
			value += counterValue.Value
			metricRep = counterValue.MetricRep
		}
		rcvr.metricBuilder.RecordAny(now, value, metricRep.Name, metricRep.Attributes)
	}

	return rcvr.metricBuilder.Emit(), errs
}

// shutdown closes the watchers
func (rcvr iisReceiver) shutdown(ctx context.Context) error {
	var errs error
	for _, watcher := range rcvr.watchers {
		err := watcher.Close()
		if err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return errs
}
