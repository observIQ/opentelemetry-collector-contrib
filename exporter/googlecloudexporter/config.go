// Copyright 2019, OpenTelemetry Authors
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

package googlecloudexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter"

import (
	"fmt"

	"github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/collector"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"google.golang.org/api/option"
)

// Config defines configuration for Google Cloud exporter.
type Config struct {
	config.ExporterSettings `mapstructure:",squash"`
	CollectorConfig         `mapstructure:",squash"`

	// Timeout for all API calls. If not set, defaults to 12 seconds.
	exporterhelper.TimeoutSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
	exporterhelper.QueueSettings   `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings   `mapstructure:"retry_on_failure"`

	LogConfig LogConfig `mapstructure:"log"`
}

// CollectorConfig defines configuration for Google Cloud exporter (not using theirs because conflict with LogConfig).
type CollectorConfig struct {
	ProjectID    string                 `mapstructure:"project"`
	UserAgent    string                 `mapstructure:"user_agent"`
	TraceConfig  collector.TraceConfig  `mapstructure:"trace"`
	MetricConfig collector.MetricConfig `mapstructure:"metric"`
}

type LogConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	// Only has effect if Endpoint is not ""
	UseInsecure bool `mapstructure:"use_insecure"`
	// GetClientOptions returns additional options to be passed
	// to the underlying Google Cloud API client.
	// Must be set programmatically (no support via declarative config).
	// Optional.
	GetClientOptions func() []option.ClientOption
	// For determining which attributes to use for log name
	NameFields []string `mapstructure:"name_fields"`
}

func (cfg *Config) Validate() error {
	if err := cfg.ExporterSettings.Validate(); err != nil {
		return fmt.Errorf("exporter settings are invalid :%w", err)
	}
	googleConfig := convertToGoogleConfig(cfg.CollectorConfig)
	if err := collector.ValidateConfig(googleConfig); err != nil {
		return fmt.Errorf("googlecloud exporter settings are invalid :%w", err)
	}
	return nil
}

func convertToGoogleConfig(cfg CollectorConfig) collector.Config {
	googleConfig := collector.Config{
		ProjectID:    cfg.ProjectID,
		UserAgent:    cfg.UserAgent,
		TraceConfig:  cfg.TraceConfig,
		MetricConfig: cfg.MetricConfig,
		LogConfig:    collector.LogConfig{},
	}

	return googleConfig
}

func convertToCollectorConfig(cfg collector.Config) CollectorConfig {
	collectorConfig := CollectorConfig{
		ProjectID:    cfg.ProjectID,
		UserAgent:    cfg.UserAgent,
		TraceConfig:  cfg.TraceConfig,
		MetricConfig: cfg.MetricConfig,
	}

	return collectorConfig
}
