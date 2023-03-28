package apachepulsarreceiver

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

// custom errors
var (
	errClientNotInit    = errors.New("client not initialized")
	errScrapedNoMetrics = errors.New("failed to scrape any metrics")
)

// apachePulsarScraper handles the scraping of Apache Pulsar metrics
type apachePulsarScraper struct {
	client   client
	logger   *zap.Logger
	cfg      *Config
	settings component.TelemetrySettings
	mb       *metadata.MetricsBuilder
}

func newScraper(logger *zap.Logger, cfg *Config, settings receiver.CreateSettings) *apachePulsarScraper {
	return &apachePulsarScraper{
		logger:   logger,
		cfg:      cfg,
		settings: settings.TelemetrySettings,
		mb:       metadata.NewMetricsBuilder(cfg.MetricsBuilderConfig, settings),
	}
}

func (p *apachePulsarScraper) start(_ context.Context, host component.Host) (err error) {
	p.client, err = newClient(p.cfg, host, p.settings, p.logger)
	return
}

func (p *apachePulsarScraper) scrape(_ context.Context) (pmetric.Metrics, error) {
	md := pmetric.NewMetrics()

	return md, nil
}
