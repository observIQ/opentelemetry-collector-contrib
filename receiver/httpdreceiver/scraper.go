// Copyright 2020, OpenTelemetry Authors
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

package httpdreceiver

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/consumer/simple"
	"go.uber.org/zap"
)

type httpdScraper struct {
	httpClient *http.Client
	// client     *client.HttpdClient

	logger *zap.Logger
	cfg    *Config
}

func newHttpdScraper(
	logger *zap.Logger,
	cfg *Config,
) *httpdScraper {
	return &httpdScraper{
		logger: logger,
		cfg:    cfg,
	}
}

func (r *httpdScraper) start(_ context.Context, host component.Host) error {
	httpClient, err := r.cfg.ToClient(host.GetExtensions())
	if err != nil {
		return err
	}
	r.httpClient = httpClient

	return nil
}

func (r *httpdScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {
	// Init client in scrape method in case there are transient errors in the
	// constructor.
	// TODO: use client to call server status auto
	// if r.client == nil {
	// 	var err error
	// 	r.client, err = client.NewHttpdClient(r.httpClient, r.cfg.HTTPClientSettings.Endpoint)
	// 	if err != nil {
	// 		r.client = nil
	// 		return pdata.ResourceMetricsSlice{}, err
	// 	}
	// }

	metrics := simple.Metrics{
		Metrics:   pdata.NewMetrics(),
		Timestamp: time.Now(),
		// MetricFactoriesByName:      metadata.M.FactoriesByName(),
		InstrumentationLibraryName: "otelcol/httpd",
	}

	r.logger.Error("hello world")

	// stats, err := r.client.GetStubStats()
	// if err != nil {
	// 	r.logger.Error("Failed to fetch nginx stats", zap.Error(err))
	// 	return pdata.ResourceMetricsSlice{}, err
	// }

	// metrics.AddSumDataPoint(metadata.M.HttpdRequests.Name(), stats.Requests)
	// metrics.AddSumDataPoint(metadata.M.HttpdConnectionsAccepted.Name(), stats.Connections.Accepted)
	// metrics.AddSumDataPoint(metadata.M.HttpdConnectionsHandled.Name(), stats.Connections.Handled)

	// metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Active}).AddGaugeDataPoint(metadata.M.HttpdConnectionsCurrent.Name(), stats.Connections.Active)
	// metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Reading}).AddGaugeDataPoint(metadata.M.HttpdConnectionsCurrent.Name(), stats.Connections.Reading)
	// metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Writing}).AddGaugeDataPoint(metadata.M.HttpdConnectionsCurrent.Name(), stats.Connections.Writing)
	// metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Waiting}).AddGaugeDataPoint(metadata.M.HttpdConnectionsCurrent.Name(), stats.Connections.Waiting)

	return metrics.Metrics.ResourceMetrics(), nil
}
