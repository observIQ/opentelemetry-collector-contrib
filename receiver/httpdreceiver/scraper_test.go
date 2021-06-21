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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpdreceiver/internal/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	httpdMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.String() == "/server-status?auto" {
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte(`ServerUptimeSeconds: 410
Total Accesses: 14169
ReqPerSec: 719.771
BytesPerSec: 1129490
BusyWorkers: 13
IdleWorkers: 227
ConnsTotal: 110
Scoreboard: S_DD_L_GGG_____W__IIII_C________________W__________________________________.........................____WR______W____W________________________C______________________________________W_W____W______________R_________R________C_________WK_W________K_____W__C__________W___R______.............................................................................................................................
`))
			return
		}
		rw.WriteHeader(404)
	}))
	sc := newHttpdScraper(zap.NewNop(), &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: httpdMock.URL + "/server-status?auto",
		},
	})

	err := sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	assert.NotNil(t, sc.httpClient)

	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 7, ms.Len())

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		switch m.Name() {
		case metadata.M.HttpdUptime.Name():
			dps := m.IntSum().DataPoints()
			require.Equal(t, 1, m.IntSum().DataPoints().Len())
			require.True(t, m.IntSum().IsMonotonic())
			require.EqualValues(t, 410, dps.At(0).Value())
		case metadata.M.HttpdCurrentConnections.Name():
			dps := m.IntGauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 110, dps.At(0).Value())
		case metadata.M.HttpdWorkers.Name():
			dps := m.IntGauge().DataPoints()
			require.Equal(t, 2, m.IntGauge().DataPoints().Len())
			busyWorker := dps.At(0).Value()
			idleWorker := dps.At(1).Value()
			require.EqualValues(t, 13, busyWorker)
			require.EqualValues(t, 227, idleWorker)
		case metadata.M.HttpdRequests.Name():
			dps := m.DoubleGauge().DataPoints()
			require.Equal(t, 1, m.DoubleGauge().DataPoints().Len())
			require.EqualValues(t, 719.771, dps.At(0).Value())
		case metadata.M.HttpdBytes.Name():
			dps := m.DoubleGauge().DataPoints()
			require.Equal(t, 1, m.DoubleGauge().DataPoints().Len())
			require.EqualValues(t, 1129490, dps.At(0).Value())
		case metadata.M.HttpdTraffic.Name():
			dps := m.IntSum().DataPoints()
			require.Equal(t, 1, m.IntSum().DataPoints().Len())
			require.True(t, m.IntSum().IsMonotonic())
			require.EqualValues(t, 14169, dps.At(0).Value())
		case metadata.M.HttpdScoreboard.Name():
			dps := m.IntGauge().DataPoints()
			require.Equal(t, 11, dps.Len())
			scoreboardMetrics := map[string]int{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.ScoreboardState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				scoreboardMetrics[label] = int(dp.Value())
			}
			require.Equal(t, 11, len(scoreboardMetrics))
			require.Equal(t, map[string]int{
				"httpd.scoreboard state:open":         150,
				"httpd.scoreboard state:waiting":      217,
				"httpd.scoreboard state:starting":     1,
				"httpd.scoreboard state:reading":      4,
				"httpd.scoreboard state:sending":      12,
				"httpd.scoreboard state:keepalive":    2,
				"httpd.scoreboard state:dnslookup":    2,
				"httpd.scoreboard state:closing":      4,
				"httpd.scoreboard state:logging":      1,
				"httpd.scoreboard state:finishing":    3,
				"httpd.scoreboard state:idle_cleanup": 4,
			}, scoreboardMetrics)

		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
}

func TestScraperFailedStart(t *testing.T) {
	sc := newHttpdScraper(zap.NewNop(), &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "localhost:8080",
			TLSSetting: configtls.TLSClientSetting{
				TLSSetting: configtls.TLSSetting{
					CAFile: "/non/existent",
				},
			},
		},
	})
	err := sc.start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
}

func TestParseScoreboard(t *testing.T) {

	t.Run("test freq count", func(t *testing.T) {
		scoreboard := `S_DD_L_GGG_____W__IIII_C________________W__________________________________.........................____WR______W____W________________________C______________________________________W_W____W______________R_________R________C_________WK_W________K_____W__C__________W___R______.............................................................................................................................`
		results := parseScoreboard(scoreboard)

		require.EqualValues(t, int64(150), results["open"])
		require.EqualValues(t, int64(217), results["waiting"])
		require.EqualValues(t, int64(1), results["starting"])
		require.EqualValues(t, int64(4), results["reading"])
		require.EqualValues(t, int64(12), results["sending"])
		require.EqualValues(t, int64(2), results["keepalive"])
		require.EqualValues(t, int64(2), results["dnslookup"])
		require.EqualValues(t, int64(4), results["closing"])
		require.EqualValues(t, int64(1), results["logging"])
		require.EqualValues(t, int64(3), results["finishing"])
		require.EqualValues(t, int64(4), results["idle_cleanup"])
	})

	t.Run("test empty defaults", func(t *testing.T) {
		emptyString := ""
		results := parseScoreboard(emptyString)

		require.EqualValues(t, int64(0), results["open"])
		require.EqualValues(t, int64(0), results["waiting"])
		require.EqualValues(t, int64(0), results["starting"])
		require.EqualValues(t, int64(0), results["reading"])
		require.EqualValues(t, int64(0), results["sending"])
		require.EqualValues(t, int64(0), results["keepalive"])
		require.EqualValues(t, int64(0), results["dnslookup"])
		require.EqualValues(t, int64(0), results["closing"])
		require.EqualValues(t, int64(0), results["logging"])
		require.EqualValues(t, int64(0), results["finishing"])
		require.EqualValues(t, int64(0), results["idle_cleanup"])
	})
}

func TestParseInt(t *testing.T) {
	require.EqualValues(t, int64(10), parseInt("10"))
	require.EqualValues(t, int64(0), parseInt("0"))
}

func TestParseFloat(t *testing.T) {
	require.EqualValues(t, float64(10.5), parseFloat("10.5"))
	require.EqualValues(t, float64(0.0), parseFloat("0"))
}

func TestParseStats(t *testing.T) {
	t.Run("with empty value", func(t *testing.T) {
		emptyString := ""
		require.EqualValues(t, map[string]string{}, parseStats(emptyString))
	})
	t.Run("with multi colons", func(t *testing.T) {
		got := "CurrentTime: Thursday, 17-Jun-2021 14:06:32 UTC"
		want := map[string]string{
			"CurrentTime": "Thursday, 17-Jun-2021 14:06:32 UTC",
		}
		require.EqualValues(t, want, parseStats(got))
	})
	t.Run("with header/footer", func(t *testing.T) {
		got := `localhost
ReqPerSec: 719.771
IdleWorkers: 227
ConnsTotal: 110
		`
		want := map[string]string{
			"ReqPerSec":   "719.771",
			"IdleWorkers": "227",
			"ConnsTotal":  "110",
		}
		require.EqualValues(t, want, parseStats(got))
	})

}

func TestScraperError(t *testing.T) {
	t.Run("no client", func(t *testing.T) {
		sc := newHttpdScraper(zap.NewNop(), &Config{})
		sc.httpClient = nil

		_, err := sc.scrape(context.Background())
		require.Error(t, err)
		require.EqualValues(t, errors.New("failed to connect to http client"), err)
	})
}
