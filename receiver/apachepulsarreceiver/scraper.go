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
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachepulsarreceiver/internal/models"
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

	collectedMetrics := false

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
	} else {
		collectedMetrics = true
		// scrape metrics for each topic
		for name, stats := range topicStats {
			// get metrics I want from stats
			msgRateIn := stats.MsgRateIn // can't get count, only rate
			avgMsgSize := stats.AverageMsgSize

			var totalUnackedMsgs int64 = 0

			for _, subStats := range stats.Subscriptions {
				totalUnackedMsgs += int64(subStats.UnAckedMessages)
			}
			metrics := models.TopicMetrics{TopicName: *name, MsgRateIn: msgRateIn, AvgMsgSize: avgMsgSize, UnackedMessages: totalUnackedMsgs}
			s.collectTopicMetrics(&metrics, now)
		}
	}

	if !collectedMetrics {
		return pmetric.NewMetrics(), errScrapedNoMetrics
	}
	return s.mb.Emit(), scrapeErrors.Combine()
}

func (s *apachePulsarScraper) collectTopicMetrics(topicMetrics *models.TopicMetrics, now pcommon.Timestamp) {
	fmt.Println("Average message size being collected: ", topicMetrics.AvgMsgSize)
	s.mb.RecordTopicAvgmsgsizeDataPoint(now, int64(topicMetrics.AvgMsgSize), string(topicMetrics.TopicName.GetLocalName()))
	s.mb.RecordTopicMsginrateDataPoint(now, int64(topicMetrics.MsgRateIn), string(topicMetrics.TopicName.GetLocalName()))
	s.mb.RecordTopicSubUnackedmsgsDataPoint(now, topicMetrics.UnackedMessages, string(topicMetrics.TopicName.GetLocalName()))

	s.mb.Emit()
}
