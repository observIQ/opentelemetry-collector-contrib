package loggenreceiver

import (
	"time"

	"go.opentelemetry.io/collector/config"
)

type Config struct {
	config.ReceiverSettings `mapstructure:",squash"`

	LogsPerSec int `mapstructure:"logs_per_sec"`

	LogLine string `mapstructure:"log_line"`

	EmitInterval time.Duration `mapstructure:"emit_interval"`
}
