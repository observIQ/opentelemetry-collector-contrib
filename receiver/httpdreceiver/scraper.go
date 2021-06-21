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

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpdreceiver/internal/metadata"
)

type httpdScraper struct {
	httpClient *http.Client

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

	if r.httpClient == nil {
		return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to http client")
	}

	metrics := simple.Metrics{
		Metrics:                    pdata.NewMetrics(),
		Timestamp:                  time.Now(),
		MetricFactoriesByName:      metadata.M.FactoriesByName(),
		InstrumentationLibraryName: "otelcol/httpd",
	}

	stats, err := r.GetStats()
	if err != nil {
		r.logger.Error("Failed to fetch httpd stats", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	for metricKey, metricValue := range parseStats(stats) {
		switch metricKey {
		case "ServerUptimeSeconds":
			metrics.AddSumDataPoint(metadata.M.HttpdUptime.Name(), parseInt(metricValue))
		case "ConnsTotal":
			metrics.AddGaugeDataPoint(metadata.M.HttpdCurrentConnections.Name(), parseInt(metricValue))
		case "BusyWorkers":
			metrics.WithLabels(map[string]string{metadata.L.WorkersState: "busy"}).AddGaugeDataPoint(metadata.M.HttpdWorkers.Name(), parseInt(metricValue))
		case "IdleWorkers":
			metrics.WithLabels(map[string]string{metadata.L.WorkersState: "idle"}).AddGaugeDataPoint(metadata.M.HttpdWorkers.Name(), parseInt(metricValue))
		case "ReqPerSec":
			metrics.AddDGaugeDataPoint(metadata.M.HttpdRequests.Name(), parseFloat(metricValue))
		case "BytesPerSec":
			metrics.AddDGaugeDataPoint(metadata.M.HttpdBytes.Name(), parseFloat(metricValue))
		case "Total Accesses":
			metrics.AddSumDataPoint(metadata.M.HttpdTraffic.Name(), parseInt(metricValue))
		case "Scoreboard":
			scoreboard := parseScoreboard(metricValue)
			for identifier, score := range scoreboard {
				metrics.WithLabels(map[string]string{metadata.L.ScoreboardState: identifier}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), score)
			}
		}
	}

	return metrics.Metrics.ResourceMetrics(), nil
}

// GetStats collects metric stats by making a get request at an endpoint.
func (r *httpdScraper) GetStats() (string, error) {
	resp, err := r.httpClient.Get(r.cfg.HTTPClientSettings.Endpoint)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
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

// parseFloat converts string to float64.
func parseFloat(value string) float64 {
	f, _ := strconv.ParseFloat(value, 64)
	return f
}

// parseInt converts string to int64.
func parseInt(value string) int64 {
	i, _ := strconv.ParseInt(value, 10, 64)
	return i
}

// parseScoreboard quantifies the symbolic mapping of the scoreboard.
func parseScoreboard(values string) map[string]int64 {
	scoreboard := map[string]int64{
		"waiting":      0,
		"starting":     0,
		"reading":      0,
		"sending":      0,
		"keepalive":    0,
		"dnslookup":    0,
		"closing":      0,
		"logging":      0,
		"finishing":    0,
		"idle_cleanup": 0,
		"open":         0,
	}

	for _, char := range values {
		switch string(char) {
		case "_":
			scoreboard["waiting"] += 1
		case "S":
			scoreboard["starting"] += 1
		case "R":
			scoreboard["reading"] += 1
		case "W":
			scoreboard["sending"] += 1
		case "K":
			scoreboard["keepalive"] += 1
		case "D":
			scoreboard["dnslookup"] += 1
		case "C":
			scoreboard["closing"] += 1
		case "L":
			scoreboard["logging"] += 1
		case "G":
			scoreboard["finishing"] += 1
		case "I":
			scoreboard["idle_cleanup"] += 1
		case ".":
			scoreboard["open"] += 1
		}
	}
	return scoreboard
}
