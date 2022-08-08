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

package filterprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"

import (
	"errors"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/multierr"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filtermetric"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterset"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterset/regexp"
)

// Config defines configuration for Resource processor.
type Config struct {
	config.ProcessorSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct

	Metrics MetricFilters `mapstructure:"metrics"`

	Logs LogFilters `mapstructure:"logs"`

	Spans SpanFilters `mapstructure:"spans"`
}

// MetricFilters filters by Metric properties.
type MetricFilters struct {
	// Include match properties describe metrics that should be included in the Collector Service pipeline,
	// all other metrics should be dropped from further processing.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Include *filtermetric.MatchProperties `mapstructure:"include"`

	// Exclude match properties describe metrics that should be excluded from the Collector Service pipeline,
	// all other metrics should be included.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Exclude *filtermetric.MatchProperties `mapstructure:"exclude"`

	// RegexpConfig specifies options for the Regexp match type
	RegexpConfig *regexp.Config `mapstructure:"regexp"`
}

// SpanFilters filters by Span attributes and various other fields, Regexp config is per matcher
type SpanFilters struct {
	// Include match properties describe spans that should be included in the Collector Service pipeline,
	// all other spans should be dropped from further processing.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Include *filterconfig.MatchProperties `mapstructure:"include"`

	// Exclude match properties describe spans that should be excluded from the Collector Service pipeline,
	// all other spans should be included.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Exclude *filterconfig.MatchProperties `mapstructure:"exclude"`
}

// LogFilters filters by Log properties.
type LogFilters struct {
	// Include match properties describe logs that should be included in the Collector Service pipeline,
	// all other logs should be dropped from further processing.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Include *LogMatchProperties `mapstructure:"include"`
	// Exclude match properties describe logs that should be excluded from the Collector Service pipeline,
	// all other logs should be included.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Exclude *LogMatchProperties `mapstructure:"exclude"`
}

// LogMatchType specifies the strategy for matching against `plog.Log`s.
type LogMatchType string

// These are the MatchTypes that users can specify for filtering
// `plog.Log`s.
const (
	Strict = LogMatchType(filterset.Strict)
	Regexp = LogMatchType(filterset.Regexp)
)

var severityToNumber = map[string]plog.SeverityNumber{
	"TRACE":  plog.SeverityNumberTRACE,
	"TRACE2": plog.SeverityNumberTRACE2,
	"TRACE3": plog.SeverityNumberTRACE3,
	"TRACE4": plog.SeverityNumberTRACE4,
	"DEBUG":  plog.SeverityNumberDEBUG,
	"DEBUG2": plog.SeverityNumberDEBUG2,
	"DEBUG3": plog.SeverityNumberDEBUG3,
	"DEBUG4": plog.SeverityNumberDEBUG4,
	"INFO":   plog.SeverityNumberINFO,
	"INFO2":  plog.SeverityNumberINFO2,
	"INFO3":  plog.SeverityNumberINFO3,
	"INFO4":  plog.SeverityNumberINFO4,
	"WARN":   plog.SeverityNumberWARN,
	"WARN2":  plog.SeverityNumberWARN2,
	"WARN3":  plog.SeverityNumberWARN3,
	"WARN4":  plog.SeverityNumberWARN4,
	"ERROR":  plog.SeverityNumberERROR,
	"ERROR2": plog.SeverityNumberERROR2,
	"ERROR3": plog.SeverityNumberERROR3,
	"ERROR4": plog.SeverityNumberERROR4,
	"FATAL":  plog.SeverityNumberFATAL,
	"FATAL2": plog.SeverityNumberFATAL2,
	"FATAL3": plog.SeverityNumberFATAL3,
	"FATAL4": plog.SeverityNumberFATAL4,
}

var errInvalidSeverity = errors.New("not a valid severity")

// logSeverity is a type that represents a SeverityNumber as a string
type logSeverity string

// validate checks that the logSeverity is valid
func (l logSeverity) validate() error {
	if l == "" {
		// No severity specified, which means to ignore this field.
		return nil
	}

	capsSeverity := strings.ToUpper(string(l))
	if _, ok := severityToNumber[capsSeverity]; !ok {
		return fmt.Errorf("'%s' is not a valid severity: %w", string(l), errInvalidSeverity)
	}
	return nil
}

// severityNumber returns the severity number that the logSeverity represents
func (l logSeverity) severityNumber() plog.SeverityNumber {
	capsSeverity := strings.ToUpper(string(l))
	return severityToNumber[capsSeverity]
}

// LogMatchProperties specifies the set of properties in a log to match against and the
// type of string pattern matching to use.
type LogMatchProperties struct {
	// LogMatchType specifies the type of matching desired
	LogMatchType LogMatchType `mapstructure:"match_type"`

	// ResourceAttributes defines a list of possible resource attributes to match logs against.
	// A match occurs if any resource attribute matches all expressions in this given list.
	ResourceAttributes []filterconfig.Attribute `mapstructure:"resource_attributes"`

	// RecordAttributes defines a list of possible record attributes to match logs against.
	// A match occurs if any record attribute matches at least one expression in this given list.
	RecordAttributes []filterconfig.Attribute `mapstructure:"record_attributes"`

	// SeverityTexts is a list of strings that the LogRecord's severity text field must match
	// against.
	SeverityTexts []string `mapstructure:"severity_texts"`

	// MinSeverity is the minimum severity needed for the log record to match.
	// This corresponds to the short names specified here:
	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#displaying-severity
	// this field is case-insensitive ("INFO" == "info")
	MinSeverity logSeverity `mapstructure:"min_severity"`

	// MatchUndefinedSeverity lets logs records with "unknown" severity match.
	// This is only applied if MinSeverity is set.
	// If MinSeverity is not set, this field is ignored, as fields are not matched based on severity.
	MatchUndefinedSeverity bool `mapstructure:"match_undefined_severity"`

	// LogBodies is a list of strings that the LogRecord's body field must match
	// against.
	LogBodies []string `mapstructure:"bodies"`
}

// validate checks that the LogMatchProperties is valid
func (lmp LogMatchProperties) validate() error {
	return lmp.MinSeverity.validate()
}

// isEmpty returns true if the properties is "empty" (meaning, there are no filters specified)
// if this is the case, the filter should be ignored.
func (lmp LogMatchProperties) isEmpty() bool {
	return len(lmp.ResourceAttributes) == 0 && len(lmp.RecordAttributes) == 0 &&
		len(lmp.SeverityTexts) == 0 && len(lmp.LogBodies) == 0 && lmp.MinSeverity == ""
}

// matchProperties converts the LogMatchProperties to a corresponding filterconfig.MatchProperties
func (lmp LogMatchProperties) matchProperties() *filterconfig.MatchProperties {
	return &filterconfig.MatchProperties{
		Config: filterset.Config{
			MatchType: filterset.MatchType(lmp.LogMatchType),
		},
		Resources:                 lmp.ResourceAttributes,
		Attributes:                lmp.RecordAttributes,
		LogSeverityTexts:          lmp.SeverityTexts,
		LogBodies:                 lmp.LogBodies,
		LogMinSeverity:            lmp.MinSeverity.severityNumber(),
		LogMatchUndefinedSeverity: lmp.MatchUndefinedSeverity,
	}
}

var _ config.Processor = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	var err error

	if cfg.Logs.Include != nil {
		err = multierr.Append(err, cfg.Logs.Include.validate())
	}

	if cfg.Logs.Exclude != nil {
		err = multierr.Append(err, cfg.Logs.Exclude.validate())
	}

	return err
}
