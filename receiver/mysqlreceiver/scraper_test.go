// Copyright 2021, observIQ
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

package mysqlreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver/internal/metadata"
)

func TestScraper(t *testing.T) {
	mysqlMock := fakeClient{}
	sc := newMySQLScraper(zap.NewNop(), &Config{
		User:     "otel",
		Password: "otel",
		Endpoint: "localhost:3306",
	})
	sc.client = &mysqlMock

	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 14, ms.Len())
	require.Equal(t, 14, len(metadata.M.Names()))

	metricsCount := map[string]int{}

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		metricsCount[m.Name()]++
	}
	require.Equal(t, map[string]int{
		"mysql.buffer_pool_operations": 7,
		"mysql.buffer_pool_pages":      6,
		"mysql.buffer_pool_size":       3,
		"mysql.commands":               6,
		"mysql.double_writes":          2,
		"mysql.handlers":               18,
		"mysql.locks":                  2,
		"mysql.log_operations":         3,
		"mysql.operations":             3,
		"mysql.page_operations":        3,
		"mysql.row_locks":              2,
		"mysql.row_operations":         4,
		"mysql.sorts":                  4,
		"mysql.threads":                4,
	}, metricsCount)
}

func TestScrapeErrorBadConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		user     string
		password string
		endpoint string
	}{
		{
			desc:     "no user",
			password: "otel",
			endpoint: "localhost:3306",
		},
		{
			desc:     "no password",
			user:     "otel",
			endpoint: "localhost:3306",
		},
		{
			desc:     "no endpoint",
			user:     "otel",
			password: "otel",
			endpoint: "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mysqlMock := fakeClient{}
			sc := newMySQLScraper(zap.NewNop(), &Config{
				User:     tC.user,
				Password: tC.password,
				Endpoint: tC.endpoint,
			})
			sc.client = &mysqlMock
			err := sc.start(context.Background(), componenttest.NewNopHost())
			require.NotNil(t, err)
		})
	}

	t.Run("good config", func(t *testing.T) {
		mysqlMock := fakeClient{}
		sc := newMySQLScraper(zap.NewNop(), &Config{
			User:     "otel",
			Password: "otel",
			Endpoint: "localhost:3306",
		})
		sc.client = &mysqlMock
		err := sc.start(context.Background(), componenttest.NewNopHost())
		require.Nil(t, err)
	})
}
