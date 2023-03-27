// Copyright  OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flumereceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/flumereceiver"

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scrapererror"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/flumereceiver/internal/metadata"
)

type flumeMetrics struct {
	SinkMetrics struct {
		ConnectionCreatedCount string `json:"ConnectionCreatedCount"`
		ConnectionClosedCount  string `json:"ConnectionClosedCount"`
	} `json:"SINK.output"`

	SourceMetrics struct {
		EventAcceptedCount string `json:"EventAcceptedCount"`
	} `json:"SOURCE.seqgen"`

	ChannelMetrics struct {
		ChannelSize           string `json:"ChannelSize"`
		EventTakeSuccessCount string `json:"EventTakeSuccessCount"`
		EventPutSuccessCount  string `json:"EventPutSuccessCount"`
	} `json:"CHANNEL.memchan"`
}

type flumeScraper struct {
	settings   component.TelemetrySettings
	cfg        *Config
	httpClient *http.Client
	mb         *metadata.MetricsBuilder
	serverName string
	port       string
}

func newFlumeScraper(
	settings receiver.CreateSettings,
	cfg *Config,
	serverName string,
	port string,
) *flumeScraper {
	a := &flumeScraper{
		settings:   settings.TelemetrySettings,
		cfg:        cfg,
		mb:         metadata.NewMetricsBuilder(cfg.MetricsBuilderConfig, settings),
		serverName: serverName,
		port:       port,
	}

	return a
}

func (r *flumeScraper) start(_ context.Context, host component.Host) error {
	httpClient, err := r.cfg.ToClient(host, r.settings)
	if err != nil {
		return err
	}
	r.httpClient = httpClient
	return nil
}

func (r *flumeScraper) scrape(context.Context) (pmetric.Metrics, error) {
	if r.httpClient == nil {
		return pmetric.Metrics{}, errors.New("failed to connect to Apache HTTPd")
	}

	stats, err := r.GetStats()
	if err != nil {
		r.settings.Logger.Error("failed to fetch Apache Httpd stats", zap.Error(err))
		return pmetric.Metrics{}, err
	}

	errs := &scrapererror.ScrapeErrors{}
	now := pcommon.NewTimestampFromTime(time.Now())

	// go through flumeMetrics values
	addPartialIfError(errs, r.mb.RecordFlumeConnectionCreatedCountDataPoint(now, stats.SinkMetrics.ConnectionCreatedCount))
	addPartialIfError(errs, r.mb.RecordFlumeConnectionClosedCountDataPoint(now, stats.SinkMetrics.ConnectionClosedCount))
	addPartialIfError(errs, r.mb.RecordFlumeEventAcceptedCountDataPoint(now, stats.SourceMetrics.EventAcceptedCount))
	addPartialIfError(errs, r.mb.RecordFlumeChannelSizeDataPoint(now, stats.ChannelMetrics.ChannelSize))
	addPartialIfError(errs, r.mb.RecordFlumeEventTakeSuccessCountDataPoint(now, stats.ChannelMetrics.EventTakeSuccessCount))
	addPartialIfError(errs, r.mb.RecordFlumeEventPutSuccessCountDataPoint(now, stats.ChannelMetrics.EventPutSuccessCount))

	return r.mb.Emit(), errs.Combine()
}

func addPartialIfError(errs *scrapererror.ScrapeErrors, err error) {
	if err != nil {
		errs.AddPartial(1, err)
	}
}

// GetStats collects metric stats by making a get request at an endpoint.
func (r *flumeScraper) GetStats() (flumeMetrics, error) {
	resp, err := r.httpClient.Get(r.cfg.Endpoint)
	if err != nil {
		return flumeMetrics{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return flumeMetrics{}, err
	}

	var metrics flumeMetrics
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		return flumeMetrics{}, err
	}

	return metrics, nil
}
