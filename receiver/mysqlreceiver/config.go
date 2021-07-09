package mysqlreceiver

import (
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	User                                    string `mapstructure:"user"`
	Password                                string `mapstructure:"password"`
	Endpoint                                string `mapstructure:"endpoint"`
}
