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

package flumereceiver

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		desc        string
		endpoint    string
		errExpected bool
		errText     string
	}{
		{
			desc:        "default_endpoint",
			endpoint:    "http://localhost:41414/metrics",
			errExpected: false,
		},
		{
			desc:        "custom_host",
			endpoint:    "http://123.123.123.123:41414/metrics",
			errExpected: false,
		},
		{
			desc:        "custom_port",
			endpoint:    "http://123.123.123.123:9090/metrics",
			errExpected: false,
		},
		{
			desc:        "custom_path",
			endpoint:    "http://localhost:8080/",
			errExpected: true,
			errText:     "path must be 'metrics': 'http://localhost:8080/'",
		},
		{
			desc:        "empty_path",
			endpoint:    "",
			errExpected: true,
			errText:     "missing hostname: ''",
		},
		{
			desc:        "missing_hostname",
			endpoint:    "http://:8080/metrics",
			errExpected: true,
			errText:     "missing hostname: 'http://:8080/metrics'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cfg := NewFactory().CreateDefaultConfig().(*Config)
			cfg.Endpoint = tc.endpoint
			err := component.ValidateConfig(cfg)
			if tc.errExpected {
				require.EqualError(t, err, tc.errText)
				return
			}
			require.NoError(t, err)
		})
	}
}
