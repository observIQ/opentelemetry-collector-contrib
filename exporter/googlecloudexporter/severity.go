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
	"go.opentelemetry.io/collector/pdata/plog"
	sev "google.golang.org/genproto/googleapis/logging/type"
)

var fastSev = map[plog.SeverityNumber]sev.LogSeverity{
	plog.SeverityNumberFATAL:     sev.LogSeverity_CRITICAL,
	plog.SeverityNumberERROR:     sev.LogSeverity_ERROR,
	plog.SeverityNumberWARN:      sev.LogSeverity_WARNING,
	plog.SeverityNumberINFO:      sev.LogSeverity_INFO,
	plog.SeverityNumberDEBUG:     sev.LogSeverity_DEBUG,
	plog.SeverityNumberTRACE:     sev.LogSeverity_DEBUG,
	plog.SeverityNumberUNDEFINED: sev.LogSeverity_DEFAULT,
}

// convertSeverity converts from otel's severity to google's severity levels
func convertSeverity(s plog.SeverityNumber) sev.LogSeverity {
	if logSev, ok := fastSev[s]; ok {
		return logSev
	}

	switch {
	case s >= plog.SeverityNumberFATAL:
		return sev.LogSeverity_CRITICAL
	case s >= plog.SeverityNumberERROR:
		return sev.LogSeverity_ERROR
	case s >= plog.SeverityNumberWARN:
		return sev.LogSeverity_WARNING
	case s >= plog.SeverityNumberINFO:
		return sev.LogSeverity_INFO
	case s > plog.SeverityNumberTRACE:
		return sev.LogSeverity_DEBUG
	default:
		return sev.LogSeverity_DEFAULT
	}
}
