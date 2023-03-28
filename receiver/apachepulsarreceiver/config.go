package apachepulsarreceiver

import (
	"time"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachepulsarreceiver/internal/metadata"
)

const (
	defaultCollectionInterval = 60 * time.Second
	defaultEndpoint           = "http://localhost:8080/admin/v2"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"` // JSON to Go struct "rules" (Go tag, used for marshaling/unmarshaling)
	metadata.MetricsBuilderConfig           `mapstructure:",squash"`
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
}
