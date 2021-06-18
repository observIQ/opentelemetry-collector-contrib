// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
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
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpdreceiver/internal/metadata"
)

type HttpdIntegrationSuite struct {
	suite.Suite
}

func TestHttpdIntegration(t *testing.T) {
	suite.Run(t, new(HttpdIntegrationSuite))
}

func httpdContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.httpd",
		},
		ExposedPorts: []string{"8080:80"},
		WaitingFor:   wait.ForListeningPort("80"),
	}

	require.NoError(t, req.Validate())

	httpd, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 6)
	return httpd
}

func (suite *HttpdIntegrationSuite) TestHttpdScraperHappyPath() {
	t := suite.T()
	httpd := httpdContainer(t)
	defer httpd.Terminate(context.Background())
	hostname, err := httpd.Host(context.Background())
	require.NoError(t, err)

	cfg := &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 100 * time.Millisecond,
		},
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: fmt.Sprintf("http://%s:8080/server-status?auto", hostname),
		},
	}

	sc := newHttpdScraper(zap.NewNop(), cfg)
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.EqualValues(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.EqualValues(t, 5, ms.Len())

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)

		switch m.Name() {
		case metadata.M.HttpdCurrentConnections.Name():
			require.Equal(t, 1, m.IntGauge().DataPoints().Len())
		case metadata.M.HttpdIdleWorkers.Name():
			require.Equal(t, 1, m.IntGauge().DataPoints().Len())
		case metadata.M.HttpdRequests.Name():
			require.Equal(t, 1, m.DoubleGauge().DataPoints().Len())
		case metadata.M.HttpdTraffic.Name():
			require.Equal(t, 1, m.IntSum().DataPoints().Len())
			require.True(t, m.IntSum().IsMonotonic())
		case metadata.M.HttpdScoreboard.Name():
			dps := m.IntGauge().DataPoints()
			require.Equal(t, 11, dps.Len())
			present := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get("state")
				switch state {
				case metadata.LabelState.Waiting:
					present[state] = true
				case metadata.LabelState.Starting:
					present[state] = true
				case metadata.LabelState.Reading:
					present[state] = true
				case metadata.LabelState.Sending:
					present[state] = true
				case metadata.LabelState.Keepalive:
					present[state] = true
				case metadata.LabelState.Dnslookup:
					present[state] = true
				case metadata.LabelState.Closing:
					present[state] = true
				case metadata.LabelState.Logging:
					present[state] = true
				case metadata.LabelState.Finishing:
					present[state] = true
				case metadata.LabelState.IdleCleanup:
					present[state] = true
				case metadata.LabelState.Open:
					present[state] = true
				default:
					require.Nil(t, state, fmt.Sprintf("connections with state %s not expected", state))
				}
			}
			require.Equal(t, 11, len(present))
		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}

	}

}
