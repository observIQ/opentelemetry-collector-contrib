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

package stubreceiver

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

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/stubreceiver/internal/metadata"
)

type StubIntegrationSuite struct {
	suite.Suite
}

func TestStubIntegration(t *testing.T) {
	suite.Run(t, new(StubIntegrationSuite))
}

func stubContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.stub",
		},
		ExposedPorts: []string{"8080:80"},
		WaitingFor:   wait.ForListeningPort("80"),
	}

	require.NoError(t, req.Validate())

	stub, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 6)
	return stub
}

func (suite *StubIntegrationSuite) TestStubScraperHappyPath() {
	t := suite.T()
	stub := stubContainer(t)
	defer stub.Terminate(context.Background())
	hostname, err := stub.Host(context.Background())
	require.NoError(t, err)

	cfg := &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 100 * time.Millisecond,
		},
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: fmt.Sprintf("http://%s:8080/server-status?auto", hostname),
		},
	}

	sc := newStubScraper(zap.NewNop(), cfg)
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

	require.EqualValues(t, 7, ms.Len())

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)

		switch m.Name() {
		case metadata.M.StubUptime.Name():
			require.Equal(t, 1, m.IntSum().DataPoints().Len())
		case metadata.M.StubCurrentConnections.Name():
			require.Equal(t, 1, m.IntGauge().DataPoints().Len())
		case metadata.M.StubWorkers.Name():
			require.Equal(t, 2, m.IntGauge().DataPoints().Len())
			dps := m.IntGauge().DataPoints()
			present := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get("state")
				switch state {
				case metadata.LabelWorkersState.Busy:
					present[state] = true
				case metadata.LabelWorkersState.Idle:
					present[state] = true
				default:
					require.Nil(t, state, fmt.Sprintf("connections with state %s not expected", state))
				}
			}
			require.Equal(t, 2, len(present))
		case metadata.M.StubRequests.Name():
			require.Equal(t, 1, m.DoubleGauge().DataPoints().Len())
		case metadata.M.StubBytes.Name():
			require.Equal(t, 1, m.DoubleGauge().DataPoints().Len())
		case metadata.M.StubTraffic.Name():
			require.Equal(t, 1, m.IntSum().DataPoints().Len())
			require.True(t, m.IntSum().IsMonotonic())
		case metadata.M.StubScoreboard.Name():
			dps := m.IntGauge().DataPoints()
			require.Equal(t, 11, dps.Len())
			present := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get("state")
				switch state {
				case metadata.LabelScoreboardState.Waiting:
					present[state] = true
				case metadata.LabelScoreboardState.Starting:
					present[state] = true
				case metadata.LabelScoreboardState.Reading:
					present[state] = true
				case metadata.LabelScoreboardState.Sending:
					present[state] = true
				case metadata.LabelScoreboardState.Keepalive:
					present[state] = true
				case metadata.LabelScoreboardState.Dnslookup:
					present[state] = true
				case metadata.LabelScoreboardState.Closing:
					present[state] = true
				case metadata.LabelScoreboardState.Logging:
					present[state] = true
				case metadata.LabelScoreboardState.Finishing:
					present[state] = true
				case metadata.LabelScoreboardState.IdleCleanup:
					present[state] = true
				case metadata.LabelScoreboardState.Open:
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
