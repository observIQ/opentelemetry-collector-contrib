package apachepulsarreceiver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"
)

const (
	tenantsResponseFile = "get_tenants_response.json"
)

func TestGetTenants(t *testing.t) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "Returns a list of tenant names on success",
			testFunc: func(t *testing.T) {
				// Setup test server
				tenantData := loadAPIResponseData(t, tenantsResponseFile)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasSuffix(r.RequestURI, "stats") {
						_, err := w.Write(tenantData)
						require.NoError(t, err)
					} else {
						w.WriteHeader(http.StatusUnauthorized)
					}
				}))
				defer ts.Close()

				tc := createTestClient(t, ts.URL)

				var expected *models.Tenants
				err := json.Unmarshal(tenantData, &expected)
				require.NoError(t, err)

				tenants, err := tc.GetTenants(context.Background())
				require.NoError(t, err)
				require.Equal(t, expected, tenants)
			},
		},
	}
}

func createTestClient(t *testing.T, baseEndpoint string) client {
	t.Helper()
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = baseEndpoint

	testClient, err := newClient(cfg, componenttest.NewNopHost(), componenttest.NewNopTelemetrySettings(), zap.NewNop())
	require.NoError(t, err)
	return testClient
}

func loadAPIResponseData(t *testing, fileName string) []byte {
	t.Helper()
	fullPath := filepath.Join("testdata", "apiresponses", fileName)

	data, err := os.ReadFile(fullPath)
	require.NoError(t, err)

	return data
}
