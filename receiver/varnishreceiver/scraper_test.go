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

package varnishreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/varnishreceiver"

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/scrapertest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/scrapertest/golden"
)

func TestScrape(t *testing.T) {

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	require.NotNil(t, cfg)

	t.Run("scrape success", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("GetStats").Return(getStats("mock_response.json"))

		scraper := newVarnishScraper(componenttest.NewNopTelemetrySettings(), cfg)
		scraper.client = mockClient

		// scrapedMetrics, _ := scraper.scrape(context.Background())
		// err := golden.WriteMetrics("./testdata/scraper/expected.json", scrapedMetrics)
		// require.NoError(t, err)

		actualMetrics, err := scraper.scrape(context.Background())
		require.NoError(t, err)

		expectedFile := filepath.Join("testdata", "scraper", "expected.json")
		expectedMetrics, err := golden.ReadMetrics(expectedFile)
		require.NoError(t, err)

		require.NoError(t, scrapertest.CompareMetrics(expectedMetrics, actualMetrics))
	})

	t.Run("scrape error", func(t *testing.T) {
		obs, logs := observer.New(zap.ErrorLevel)
		settings := componenttest.NewNopTelemetrySettings()
		settings.Logger = zap.New(obs)
		mockClient := new(MockClient)
		mockClient.On("GetStats").Return(getStats(""))
		scraper := newVarnishScraper(settings, cfg)
		scraper.client = mockClient

		_, err := scraper.scrape(context.Background())
		require.NotNil(t, err)
		require.Equal(t, 1, logs.Len())
		require.Equal(t, []observer.LoggedEntry{
			{
				Entry: zapcore.Entry{Level: zap.ErrorLevel, Message: "Failed to execute varnishstat"},
				Context: []zapcore.Field{
					zap.String("Working Directory:", cfg.WorkingDir),
					zap.String("Executable Directory:", cfg.ExecutableDir),
					zap.Error(errors.New("bad response")),
				},
			},
		}, logs.AllUntimed())
	})
}

func TestStart(t *testing.T) {
	t.Run("start success", func(t *testing.T) {
		f := NewFactory()
		cfg := f.CreateDefaultConfig().(*Config)

		scraper := newVarnishScraper(componenttest.NewNopTelemetrySettings(), cfg)
		err := scraper.start(context.Background(), componenttest.NewNopHost())
		require.NoError(t, err)
	})
}

// MockClient is an autogenerated mock type for the MockClient type
type MockClient struct {
	mock.Mock
}

// GetStats provides a mock function with given fields:
func (_m *MockClient) GetStats() (*Stats, error) {
	ret := _m.Called()

	var r0 *Stats
	if rf, ok := ret.Get(0).(func() *Stats); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Stats)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetVersion provides a mock function with given fields:
func (_m *MockClient) GetVersion() (version, error) {
	ret := _m.Called()

	var r0 version
	if rf, ok := ret.Get(0).(func() version); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(version)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func getStats(filename string) (*Stats, error) {
	if filename == "" {
		return nil, errors.New("bad response")
	}

	file, err := os.Open(path.Join("testdata", "scraper", filename))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return parseStats("", body)
}
