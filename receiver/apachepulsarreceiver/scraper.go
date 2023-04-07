package apachepulsarreceiver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scrapererror"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachepulsarreceiver/internal/metadata"
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

func (s *apachePulsarScraper) start(_ context.Context, host component.Host) (err error) {
	s.client, err = newClient(s.cfg, host, s.settings, s.logger)
	if err != nil {
		return err
	}
	return nil
}

func (s *apachePulsarScraper) scrape(_ context.Context) (pmetric.Metrics, error) {
	now := pcommon.NewTimestampFromTime(time.Now())
	var scrapeErrors scrapererror.ScrapeErrors

	if s.client == nil {
		return pmetric.NewMetrics(), errClientNotInit
	}

	tenants, err := s.client.GetTenants()
	if err != nil {
		scrapeErrors.AddPartial(1, err)
		s.logger.Warn("Failed to scrape tenants", zap.Error(err))
	}
	namespaces, err := s.client.GetNameSpaces(tenants)
	if err != nil {
		scrapeErrors.AddPartial(1, err)
		s.logger.Warn("Failed to scrape namespaces", zap.Error(err))
	}
	topics, err := s.client.GetTopics(namespaces)
	if err != nil {
		scrapeErrors.AddPartial(1, err)
		s.logger.Warn("Failed to scrape topic names", zap.Error(err))
	}
	topicStats, err := s.client.GetTopicStats(topics)
	if err != nil {
		scrapeErrors.AddPartial(1, err)
		s.logger.Warn("Failed to scrape topic stat metrics", zap.Error(err))
	}
	// scrape metrics for each topic
	for name, stats := range topicStats {
		// get metrics I want from stats
		msgRateIn := stats.MsgRateIn
		avgMsgSize := stats.AverageMsgSize

		s.mb.RecordPulsarTopicAvgmsgsizeDataPoint(now, int64(avgMsgSize), name.GetLocalName())
		s.mb.RecordPulsarTopicMsginrateDataPoint(now, int64(msgRateIn), name.GetLocalName())

		var totalUnackedMsgs int64 = 0

		for _, subStats := range stats.Subscriptions {
			totalUnackedMsgs += int64(subStats.UnAckedMessages)
		}

		fmt.Printf("New Topic Metrics for topic '%s': \nAvgMsgSize: %d\tMsgRateIn: %d\tUnackedMsgs: %d\n\n",
			name.GetLocalName(), int64(avgMsgSize), int64(msgRateIn), totalUnackedMsgs)
		s.mb.RecordPulsarTopicSubUnackedmsgsDataPoint(now, totalUnackedMsgs, name.GetLocalName())
	}

	return s.mb.Emit(), scrapeErrors.Combine()
}
