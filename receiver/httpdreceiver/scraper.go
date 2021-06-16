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
	"fmt"
	"io/ioutil"
	"log"
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
	// 	r.client, err = client.NewHttpdClient(r.httpClient, r.cfg.HTTPClientSettings.Endpoint) // use this endpoint for now
	// 	if err != nil {
	// 		r.client = nil
	// 		return pdata.ResourceMetricsSlice{}, err
	// 	}
	// }

	resp, err := r.httpClient.Get(r.cfg.HTTPClientSettings.Endpoint)
	if err != nil {
		r.logger.Error(err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error(err.Error())
	}

	parsedMetrics := parseResponse(string(body))
	processedMetrics := processMetrics(parsedMetrics)

	r.logger.Error(processedMetrics.String())

	metrics := simple.Metrics{
		Metrics:                    pdata.NewMetrics(),
		Timestamp:                  time.Now(),
		MetricFactoriesByName:      metadata.M.FactoriesByName(),
		InstrumentationLibraryName: "otelcol/httpd",
	}

	metrics.AddSumDataPoint(metadata.M.HttpdTraffic.Name(), processedMetrics.TotalAccess)

	// metrics.AddSumDataPoint(metadata.M.HttpdRequests.Name(), stats.Requests)
	// metrics.AddSumDataPoint(metadata.M.HttpdConnectionsAccepted.Name(), stats.Connections.Accepted)
	// metrics.AddSumDataPoint(metadata.M.HttpdConnectionsHandled.Name(), stats.Connections.Handled)

	metrics.AddGaugeDataPoint(metadata.M.HttpdCurrentConnections.Name(), int64(processedMetrics.ConnsTotal))
	metrics.AddGaugeDataPoint(metadata.M.HttpdIdleWorkers.Name(), int64(processedMetrics.IdleWorkers))
	metrics.AddDGaugeDataPoint(metadata.M.HttpdRequests.Name(), float64(processedMetrics.ReqPerSec))

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Open}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.Open])

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Waiting}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.Waiting])

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Starting}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.Starting])

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Reading}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.Reading])

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Sending}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.Sending])

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Keepalive}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.Keepalive])

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Dnslookup}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.Dnslookup])

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Closing}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.Closing])

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Logging}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.Logging])

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Finishing}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.Finishing])

	metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.IdleCleanup}).AddGaugeDataPoint(metadata.M.HttpdScoreboard.Name(), processedMetrics.Scoreboard.Freq[metadata.LabelState.IdleCleanup])

	// metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Active}).AddGaugeDataPoint(metadata.M.HttpdConnectionsCurrent.Name(), stats.Connections.Active)
	// metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Reading}).AddGaugeDataPoint(metadata.M.HttpdConnectionsCurrent.Name(), stats.Connections.Reading)
	// metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Writing}).AddGaugeDataPoint(metadata.M.HttpdConnectionsCurrent.Name(), stats.Connections.Writing)
	// metrics.WithLabels(map[string]string{metadata.L.State: metadata.LabelState.Waiting}).AddGaugeDataPoint(metadata.M.HttpdConnectionsCurrent.Name(), stats.Connections.Waiting)

	return metrics.Metrics.ResourceMetrics(), nil
}

// parseResponse converts a response body key:values into a map.
func parseResponse(resp string) map[string]string {
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

type Metrics struct {
	ConnsTotal  float64
	IdleWorkers float64
	ReqPerSec   int64
	TotalAccess int64
	Scoreboard  Scoreboard
}

func (m *Metrics) String() string {
	return fmt.Sprintf("\n\tConnsTotal: %#v\n\tIdleWorkers: %#v\n\tReqPerSec: %#v\n\tTotalAccess: %#v\n\tScoreboard: %#v\n", m.ConnsTotal, m.IdleWorkers, m.ReqPerSec, m.TotalAccess, m.Scoreboard)
}

// parseFloat converts string to float64.
func parseFloat(value string) float64 {
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}
	log.Printf("error expected a parsable float but got %v", value)
	return 0
}

// parseInt converts string to int64.
func parseInt(value string) int64 {
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return int64(f)
	}
	log.Printf("error expected a parsable Int but got %v", value)
	return 0
}

// Scoreboard stores a description of the mappings and the freq of each instance found.
type Scoreboard struct {
	Desc string
	Freq map[string]int64
}

// NewScoreboard returns a new instance of a scoreboard.
func NewScoreboard() *Scoreboard {
	return &Scoreboard{
		// Desc: scoreboardDesc(),
		Freq: make(map[string]int64),
	}
}

// scoreboardDesc explains the meaning behind the freq.
func scoreboardDesc() string {
	desc := `Scoreboard meaning:
	"_" Waiting for Connection
	"S" Starting up
	"R" Reading Request
	"W" Sending Reply
	"K" Keepalive (read)
	"D" DNS Lookup
	"C" Closing connection
	"L" Logging
	"G" Gracefully finishing
	"I" Idle cleanup of worker
	"." Open slot with no current process`
	return desc
}

// parseScoreboard quantifies the symbolic mapping of the scoreboard.
func parseScoreboard(values string) *Scoreboard {
	scoreboard := NewScoreboard()
	for _, char := range values {
		switch string(char) {
		case "_":
			scoreboard.Freq["waiting"] += 1
		case "S":
			scoreboard.Freq["starting"] += 1
		case "R":
			scoreboard.Freq["reading"] += 1
		case "W":
			scoreboard.Freq["sending"] += 1
		case "K":
			scoreboard.Freq["keepalive"] += 1
		case "D":
			scoreboard.Freq["dnslookup"] += 1
		case "C":
			scoreboard.Freq["closing"] += 1
		case "L":
			scoreboard.Freq["logging"] += 1
		case "G":
			scoreboard.Freq["finishing"] += 1
		case "I":
			scoreboard.Freq["idle_cleanup"] += 1
		case ".":
			scoreboard.Freq["open"] += 1
		default:
			continue
		}
	}
	return scoreboard
}

// processMetrics filters out desired google metrics.
func processMetrics(metricsMap map[string]string) *Metrics {
	metrics := Metrics{}
	for k, v := range metricsMap {
		switch k {
		case "ConnsTotal":
			metrics.ConnsTotal = parseFloat(v)
		case "IdleWorkers":
			metrics.IdleWorkers = parseFloat(v)
		case "ReqPerSec":
			metrics.ReqPerSec = parseInt(v)
		case "Total Accesses":
			metrics.TotalAccess = parseInt(v)
		case "Scoreboard":
			metrics.Scoreboard = *parseScoreboard(v)
		default:
			continue
		}
	}
	return &metrics
}
