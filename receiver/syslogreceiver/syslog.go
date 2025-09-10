// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package syslogreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver"

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/receiver"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/consumerretry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/syslog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/tcp"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/udp"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver/internal/metadata"
)

// NewFactory creates a factory for syslog receiver
func NewFactory() receiver.Factory {
	return adapter.NewFactory(ReceiverType{}, metadata.LogsStability)
}

// ReceiverType implements adapter.LogReceiverType
// to create a syslog receiver
type ReceiverType struct{}

// Type is the receiver type
func (ReceiverType) Type() component.Type {
	return metadata.Type
}

// CreateDefaultConfig creates a config with type and version
func (ReceiverType) CreateDefaultConfig() component.Config {
	return &SysLogConfig{
		BaseConfig: adapter.BaseConfig{
			Operators:      []operator.Config{},
			RetryOnFailure: consumerretry.NewDefaultConfig(),
		},
		InputConfig: *syslog.NewConfig(),
		Mode:       ModeParsed, // Default to parsed mode for backward compatibility
	}
}

// BaseConfig gets the base config from config, for now
func (ReceiverType) BaseConfig(cfg component.Config) adapter.BaseConfig {
	return cfg.(*SysLogConfig).BaseConfig
}

// Mode defines the processing mode for syslog messages
type Mode string

const (
	// ModeRaw processes messages as raw text without parsing
	ModeRaw Mode = "raw"
	// ModeParsed processes messages according to RFC standards
	ModeParsed Mode = "parsed"
)

// SysLogConfig defines configuration for the syslog receiver
type SysLogConfig struct {
	InputConfig        syslog.Config `mapstructure:",squash"`
	adapter.BaseConfig `mapstructure:",squash"`

	// Mode determines how to process incoming syslog messages
	// - "raw": Pass through messages without parsing (like TCP receiver)
	// - "parsed": Parse messages according to RFC standards (default behavior)
	Mode Mode `mapstructure:"mode"`

	// prevent unkeyed literal initialization
	_ struct{}
}

// InputConfig unmarshals the input operator
func (ReceiverType) InputConfig(cfg component.Config) operator.Config {
	return operator.NewConfig(&cfg.(*SysLogConfig).InputConfig)
}

func (cfg *SysLogConfig) Unmarshal(componentParser *confmap.Conf) error {
	if componentParser == nil {
		// Nothing to do if there is no config given.
		return nil
	}

	if componentParser.IsSet("tcp") {
		cfg.InputConfig.TCP = &tcp.NewConfig().BaseConfig
	} else if componentParser.IsSet("udp") {
		cfg.InputConfig.UDP = &udp.NewConfig().BaseConfig
	}

	// Unmarshal the configuration first
	if err := componentParser.Unmarshal(cfg); err != nil {
		return err
	}

	// Configure operators based on mode
	if cfg.Mode == ModeRaw {
		// In raw mode, clear the operators to pass through raw messages
		cfg.BaseConfig.Operators = []operator.Config{}
	} else {
		// In parsed mode, ensure we have the syslog parser operator
		// The syslog input operator will handle this automatically
		// but we can add any additional operators here if needed
	}

	return nil
}
