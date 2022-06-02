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

package googlecloudexporter

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
	sev "google.golang.org/genproto/googleapis/logging/type"
)

func TestConvertSeverity(t *testing.T) {
	testCases := []struct {
		name         string
		origSeverity plog.SeverityNumber
		newSeverity  sev.LogSeverity
	}{
		{
			name:         "original severity above FATAL",
			origSeverity: plog.SeverityNumberFATAL4,
			newSeverity:  sev.LogSeverity_CRITICAL,
		},
		{
			name:         "original severity equals FATAL",
			origSeverity: plog.SeverityNumberFATAL,
			newSeverity:  sev.LogSeverity_CRITICAL,
		},
		{
			name:         "original severity above ERROR",
			origSeverity: plog.SeverityNumberERROR4,
			newSeverity:  sev.LogSeverity_ERROR,
		},
		{
			name:         "original severity equals ERROR",
			origSeverity: plog.SeverityNumberERROR,
			newSeverity:  sev.LogSeverity_ERROR,
		},
		{
			name:         "original severity above WARN",
			origSeverity: plog.SeverityNumberWARN4,
			newSeverity:  sev.LogSeverity_WARNING,
		},
		{
			name:         "original severity equals WARN",
			origSeverity: plog.SeverityNumberWARN4,
			newSeverity:  sev.LogSeverity_WARNING,
		},
		{
			name:         "original severity above INFO",
			origSeverity: plog.SeverityNumberINFO4,
			newSeverity:  sev.LogSeverity_INFO,
		},
		{
			name:         "original severity equals INFO",
			origSeverity: plog.SeverityNumberINFO,
			newSeverity:  sev.LogSeverity_INFO,
		},
		{
			name:         "original severity above DEBUG",
			origSeverity: plog.SeverityNumberDEBUG4,
			newSeverity:  sev.LogSeverity_DEBUG,
		},
		{
			name:         "original severity equals DEBUG",
			origSeverity: plog.SeverityNumberDEBUG,
			newSeverity:  sev.LogSeverity_DEBUG,
		},
		{
			name:         "original severity above TRACE",
			origSeverity: plog.SeverityNumberTRACE4,
			newSeverity:  sev.LogSeverity_DEBUG,
		},
		{
			name:         "original severity equals TRACE",
			origSeverity: plog.SeverityNumberTRACE,
			newSeverity:  sev.LogSeverity_DEBUG,
		},
		{
			name:         "original severity equals UNDEFINED",
			origSeverity: plog.SeverityNumberUNDEFINED,
			newSeverity:  sev.LogSeverity_DEFAULT,
		},
		{
			name:         "original severity not recognized",
			origSeverity: -10,
			newSeverity:  sev.LogSeverity_DEFAULT,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertSeverity(tc.origSeverity)
			require.Equal(t, tc.newSeverity, result)
		})
	}
}
