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

package stubreceiver

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/consumer/simple"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/stubreceiver/internal/metadata"
)

type stubScraper struct {
	httpClient *http.Client

	logger *zap.Logger
	cfg    *Config
}

func newStubScraper(
	logger *zap.Logger,
	cfg *Config,
) *stubScraper {
	return &stubScraper{
		logger: logger,
		cfg:    cfg,
	}
}

func (r *stubScraper) start(_ context.Context, host component.Host) error {
	httpClient, err := r.cfg.ToClient(host.GetExtensions())
	if err != nil {
		return err
	}
	r.httpClient = httpClient

	return nil
}

func (r *stubScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {

	if r.httpClient == nil {
		return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to http client")
	}

	metrics := simple.Metrics{
		Metrics:                    pdata.NewMetrics(),
		Timestamp:                  time.Now(),
		MetricFactoriesByName:      metadata.M.FactoriesByName(),
		InstrumentationLibraryName: "otelcol/stub",
	}

	stats, err := r.GetStats()
	if err != nil {
		r.logger.Error("Failed to fetch stub stats", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	for metricKey, metricValue := range parseStats(stats) {
		// TODO make metrics
		// switch metricKey {
		// case "ServerUptimeSeconds":
		// 	metrics.AddSumDataPoint(metadata.M.StubUptime.Name(), parseInt(metricValue))
		// case "ConnsTotal":
		// 	metrics.AddGaugeDataPoint(metadata.M.StubCurrentConnections.Name(), parseInt(metricValue))
		// case "BusyWorkers":
		// 	metrics.WithLabels(map[string]string{metadata.L.WorkersState: "busy"}).AddGaugeDataPoint(metadata.M.StubWorkers.Name(), parseInt(metricValue))
		// case "IdleWorkers":
		// 	metrics.WithLabels(map[string]string{metadata.L.WorkersState: "idle"}).AddGaugeDataPoint(metadata.M.StubWorkers.Name(), parseInt(metricValue))
		// case "ReqPerSec":
		// 	metrics.AddDGaugeDataPoint(metadata.M.StubRequests.Name(), parseFloat(metricValue))
		// case "BytesPerSec":
		// 	metrics.AddDGaugeDataPoint(metadata.M.StubBytes.Name(), parseFloat(metricValue))
		// case "Total Accesses":
		// 	metrics.AddSumDataPoint(metadata.M.StubTraffic.Name(), parseInt(metricValue))
		// case "Scoreboard":
		// 	scoreboard := parseScoreboard(metricValue)
		// 	for identifier, score := range scoreboard {
		// 		metrics.WithLabels(map[string]string{metadata.L.ScoreboardState: identifier}).AddGaugeDataPoint(metadata.M.StubScoreboard.Name(), score)
		// 	}
		// }
	}

	return metrics.Metrics.ResourceMetrics(), nil
}

// GetStats collects metric stats by making a get request at an endpoint.
func (r *stubScraper) GetStats() (string, error) {
	// TODO collect metrics
	// resp, err := r.httpClient.Get(r.cfg.HTTPClientSettings.Endpoint)
	// if err != nil {
	// 	return "", err
	// }

	// defer resp.Body.Close()

	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return "", err
	// }
	// return string(body), nil
}

// parseStats converts a response body key:values into a map.
func parseStats(resp string) map[string]string {
	metrics := make(map[string]string)

	fields := strings.Split(resp, "\n")
	for _, field := range fields {
		index := strings.Index(field, ": ")
		if index == -1 {
			continue
		}
		metrics[field[:index]] = field[index+2:]
	}
	return metrics
}
