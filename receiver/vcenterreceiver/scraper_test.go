// Copyright  The OpenTelemetry Authors
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

package vcenterreceiver // import github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver

import (
	"context"
	"embed"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/scrapertest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/scrapertest/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver/internal/metadata"
	mock "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver/internal/mockserver"
)

//go:embed internal/mockserver/responses/*
var resources embed.FS

func TestScrape(t *testing.T) {
	ctx := context.Background()
	mockServer := mock.MockServer(t, resources)

	cfg := &Config{
		Metrics:  metadata.DefaultMetricsSettings(),
		Endpoint: mockServer.URL,
		Username: mock.MockUsername,
		Password: mock.MockPassword,
	}
	scraper := newVmwareVcenterScraper(zap.NewNop(), cfg)

	metrics, err := scraper.scrape(ctx)
	require.NoError(t, err)
	require.NotEqual(t, metrics.MetricCount(), 0)

	goldenPath := filepath.Join("testdata", "metrics", "expected.json")
	expectedMetrics, err := golden.ReadMetrics(goldenPath)
	require.NoError(t, err)

	err = scrapertest.CompareMetrics(expectedMetrics, metrics)
	require.NoError(t, err)
	require.NoError(t, scraper.Shutdown(ctx))
}

func TestScrape_NoClient(t *testing.T) {
	ctx := context.Background()
	scraper := &vcenterMetricScraper{
		client: nil,
		config: &Config{
			Endpoint: "http://vcsa.localnet",
		},
		mb:     metadata.NewMetricsBuilder(metadata.DefaultMetricsSettings()),
		logger: zap.NewNop(),
	}
	metrics, err := scraper.scrape(ctx)
	require.ErrorContains(t, err, "unable to connect to vSphere SDK")
	require.Equal(t, metrics.MetricCount(), 0)
	require.NoError(t, scraper.Shutdown(ctx))
}

func TestStartFailures_Metrics(t *testing.T) {
	cases := []struct {
		desc string
		cfg  Config
		err  error
	}{
		{
			desc: "bad client connect",
			cfg: Config{
				Endpoint: "http://no-host",
			},
		},
		{
			desc: "unparsable endpoint",
			cfg: Config{
				Endpoint: "<protocol>://some-host",
			},
		},
	}

	ctx := context.Background()
	for _, tc := range cases {
		scraper := newVmwareVcenterScraper(zap.NewNop(), &tc.cfg)
		err := scraper.Start(ctx, nil)
		if tc.err != nil {
			require.ErrorContains(t, err, tc.err.Error())
		} else {
			require.NoError(t, err)
		}
	}
}
