// Code generated by mdatagen. DO NOT EDIT.

package metadata

import "go.opentelemetry.io/collector/confmap"

// MetricConfig provides common config for a particular metric.
type MetricConfig struct {
	Enabled bool `mapstructure:"enabled"`

	enabledSetByUser bool
}

func (ms *MetricConfig) Unmarshal(parser *confmap.Conf) error {
	if parser == nil {
		return nil
	}
	err := parser.Unmarshal(ms, confmap.WithErrorUnused())
	if err != nil {
		return err
	}
	ms.enabledSetByUser = parser.IsSet("enabled")
	return nil
}

// MetricsConfig provides config for hostmetricsreceiver/cpu metrics.
type MetricsConfig struct {
	SystemCPUFrequency     MetricConfig `mapstructure:"system.cpu.frequency"`
	SystemCPULogicalCount  MetricConfig `mapstructure:"system.cpu.logical.count"`
	SystemCPUPhysicalCount MetricConfig `mapstructure:"system.cpu.physical.count"`
	SystemCPUTime          MetricConfig `mapstructure:"system.cpu.time"`
	SystemCPUUtilization   MetricConfig `mapstructure:"system.cpu.utilization"`
}

func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		SystemCPUFrequency: MetricConfig{
			Enabled: false,
		},
		SystemCPULogicalCount: MetricConfig{
			Enabled: false,
		},
		SystemCPUPhysicalCount: MetricConfig{
			Enabled: false,
		},
		SystemCPUTime: MetricConfig{
			Enabled: true,
		},
		SystemCPUUtilization: MetricConfig{
			Enabled: false,
		},
	}
}

// MetricsBuilderConfig is a configuration for hostmetricsreceiver/cpu metrics builder.
type MetricsBuilderConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
}

func DefaultMetricsBuilderConfig() MetricsBuilderConfig {
	return MetricsBuilderConfig{
		Metrics: DefaultMetricsConfig(),
	}
}
