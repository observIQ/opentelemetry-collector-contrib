// Copyright  The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apachesparkreceiver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver/internal/models"
)

const (
	URL = "http://sparkmock.com"
)

func TestNewApacheSparkClient(t *testing.T) {
	testCases := []struct {
		desc        string
		cfg         *Config
		host        component.Host
		settings    component.TelemetrySettings
		logger      *zap.Logger
		expectError error
	}{
		{
			desc: "Valid Configuration",
			cfg: &Config{
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: defaultEndpoint,
				},
			},
			host:        componenttest.NewNopHost(),
			settings:    componenttest.NewNopTelemetrySettings(),
			logger:      zap.NewNop(),
			expectError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ac, err := newApacheSparkClient(tc.cfg, tc.host, tc.settings)
			if tc.expectError != nil {
				require.Nil(t, ac)
				require.Contains(t, err.Error(), tc.expectError.Error())
			} else {
				require.NoError(t, err)

				actualClient, ok := ac.(*apacheSparkClient)
				require.True(t, ok)

				require.Equal(t, tc.logger, actualClient.logger)
				require.NotNil(t, actualClient.client)
			}
		})
	}
}

func TestGetClusterStats(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "Returns an error on Get() failure",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, clusterStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasSuffix(r.RequestURI, "metrics/json") {
						w.WriteHeader(http.StatusUnauthorized)
					} else {
						_, err := w.Write(data)
						require.NoError(t, err)
					}
				}))
				defer ts.Close()

				tc := createTestClient(t, URL)

				clusterStats, err := tc.GetClusterStats()
				require.NotNil(t, err)
				require.Nil(t, clusterStats)
			},
		},
		{
			desc: "Handles case where result cannot be unmarshalled",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, clusterStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					if strings.HasSuffix(r.RequestURI, "/metrics/json") {
						_, err = w.Write([]byte("[{}]"))
						require.NoError(t, err)
					} else {
						_, err = w.Write(data)
					}
					require.NoError(t, err)
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				clusterStats, err := tc.GetClusterStats()
				require.Nil(t, clusterStats)
				require.NotNil(t, err)
			},
		},
		{
			desc: "Success case",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, clusterStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					if strings.HasSuffix(r.RequestURI, "/metrics/json") {
						_, err = w.Write(data)
						require.NoError(t, err)
					}
					require.NoError(t, err)
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				var expected *models.ClusterProperties
				data = loadAPIResponseData(t, clusterStatsResponseFile)
				err := json.Unmarshal(data, &expected)
				require.NoError(t, err)

				clusterStats, err := tc.GetClusterStats()
				require.NoError(t, err)
				require.NotNil(t, clusterStats)
				require.Equal(t, expected, clusterStats)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

func TestGetApplications(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "Returns an error on Get() failure",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, appsStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasSuffix(r.RequestURI, "applications") {
						w.WriteHeader(http.StatusUnauthorized)
					} else {
						_, err := w.Write(data)
						require.NoError(t, err)
					}
				}))
				defer ts.Close()

				tc := createTestClient(t, URL)

				apps, err := tc.GetApplications()
				require.NotNil(t, err)
				require.Nil(t, apps)
			},
		},
		{
			desc: "Handles case where result cannot be unmarshalled",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, appsStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					if strings.HasSuffix(r.RequestURI, "applications") {
						_, err = w.Write([]byte(""))
						require.NoError(t, err)
					} else {
						_, err = w.Write(data)
					}
					require.NoError(t, err)
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				apps, err := tc.GetApplications()
				require.Nil(t, apps)
				require.NotNil(t, err)
			},
		},
		{
			desc: "Success case",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, appsStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					if strings.HasSuffix(r.RequestURI, "applications") {
						_, err = w.Write(data)
						require.NoError(t, err)
					}
					require.NoError(t, err)
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				var expected *models.Applications
				data = loadAPIResponseData(t, appsStatsResponseFile)
				err := json.Unmarshal(data, &expected)
				require.NoError(t, err)

				apps, err := tc.GetApplications()
				require.NoError(t, err)
				require.NotNil(t, apps)
				require.Equal(t, expected, apps)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

func TestGetStageStats(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "Returns an error on Get() failure",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, stagesStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasSuffix(r.RequestURI, "stages") {
						w.WriteHeader(http.StatusUnauthorized)
					} else {
						_, err := w.Write(data)
						require.NoError(t, err)
					}
				}))
				defer ts.Close()

				tc := createTestClient(t, URL)

				stageStats, err := tc.GetStageStats("some_app_id")
				require.NotNil(t, err)
				require.Nil(t, stageStats)
			},
		},
		{
			desc: "Handles case where result cannot be unmarshalled",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, stagesStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					if strings.HasSuffix(r.RequestURI, "stages") {
						_, err = w.Write([]byte(""))
						require.NoError(t, err)
					} else {
						_, err = w.Write(data)
					}
					require.NoError(t, err)
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				stageStats, err := tc.GetStageStats("some_app_id")
				require.Nil(t, stageStats)
				require.NotNil(t, err)
			},
		},
		{
			desc: "Success case",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, stagesStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					if strings.HasSuffix(r.RequestURI, "stages") {
						_, err = w.Write(data)
						require.NoError(t, err)
					}
					require.NoError(t, err)
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				var expected *models.Stages
				data = loadAPIResponseData(t, stagesStatsResponseFile)
				err := json.Unmarshal(data, &expected)
				require.NoError(t, err)

				stageStats, err := tc.GetStageStats("some_app_id")
				require.NoError(t, err)
				require.NotNil(t, stageStats)
				require.Equal(t, expected, stageStats)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

func TestGetExecutorStats(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "Returns an error on Get() failure",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, executorsStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasSuffix(r.RequestURI, "executors") {
						w.WriteHeader(http.StatusUnauthorized)
					} else {
						_, err := w.Write(data)
						require.NoError(t, err)
					}
				}))
				defer ts.Close()

				tc := createTestClient(t, URL)

				executorStats, err := tc.GetExecutorStats("some_app_id")
				require.NotNil(t, err)
				require.Nil(t, executorStats)
			},
		},
		{
			desc: "Handles case where result cannot be unmarshalled",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, executorsStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					if strings.HasSuffix(r.RequestURI, "executors") {
						_, err = w.Write([]byte(""))
						require.NoError(t, err)
					} else {
						_, err = w.Write(data)
					}
					require.NoError(t, err)
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				executorStats, err := tc.GetExecutorStats("some_app_id")
				require.Nil(t, executorStats)
				require.NotNil(t, err)
			},
		},
		{
			desc: "Success case",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, executorsStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					if strings.HasSuffix(r.RequestURI, "executors") {
						_, err = w.Write(data)
						require.NoError(t, err)
					}
					require.NoError(t, err)
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				var expected *models.Executors
				data = loadAPIResponseData(t, executorsStatsResponseFile)
				err := json.Unmarshal(data, &expected)
				require.NoError(t, err)

				executorStats, err := tc.GetExecutorStats("some_app_id")
				require.NoError(t, err)
				require.NotNil(t, executorStats)
				require.Equal(t, expected, executorStats)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

func TestGetJobStats(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "Returns an error on Get() failure",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, jobsStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasSuffix(r.RequestURI, "jobs") {
						w.WriteHeader(http.StatusUnauthorized)
					} else {
						_, err := w.Write(data)
						require.NoError(t, err)
					}
				}))
				defer ts.Close()

				tc := createTestClient(t, URL)

				jobStats, err := tc.GetJobStats("some_app_id")
				require.NotNil(t, err)
				require.Nil(t, jobStats)
			},
		},
		{
			desc: "Handles case where result cannot be unmarshalled",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, jobsStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					if strings.HasSuffix(r.RequestURI, "jobs") {
						_, err = w.Write([]byte(""))
						require.NoError(t, err)
					} else {
						_, err = w.Write(data)
					}
					require.NoError(t, err)
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				jobStats, err := tc.GetJobStats("some_app_id")
				require.Nil(t, jobStats)
				require.NotNil(t, err)
			},
		},
		{
			desc: "Success case",
			testFunc: func(t *testing.T) {
				data := loadAPIResponseData(t, jobsStatsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					if strings.HasSuffix(r.RequestURI, "jobs") {
						_, err = w.Write(data)
						require.NoError(t, err)
					}
					require.NoError(t, err)
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				var expected *models.Jobs
				data = loadAPIResponseData(t, jobsStatsResponseFile)
				err := json.Unmarshal(data, &expected)
				require.NoError(t, err)

				jobStats, err := tc.GetJobStats("some_app_id")
				require.NoError(t, err)
				require.NotNil(t, jobStats)
				require.Equal(t, expected, jobStats)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

func createTestClient(t *testing.T, baseEndpoint string) client {
	t.Helper()
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = baseEndpoint

	testClient, err := newApacheSparkClient(cfg, componenttest.NewNopHost(), componenttest.NewNopTelemetrySettings())
	require.NoError(t, err)
	return testClient
}

func loadAPIResponseData(t *testing.T, fileName string) []byte {
	t.Helper()
	fullPath := filepath.Join("testdata", "apiresponses", fileName)

	data, err := os.ReadFile(fullPath)
	require.NoError(t, err)

	return data
}
