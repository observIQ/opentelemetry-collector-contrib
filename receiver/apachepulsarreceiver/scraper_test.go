package apachepulsarreceiver

import (
	"context"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachepulsarreceiver/internal/mocks"
	"github.com/streamnative/pulsarctl/pkg/pulsar/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {

	cfg := newDefaultConfig().(*Config)
	require.NoError(t, component.ValidateConfig(cfg))

	scraper := newScraper(zap.NewNop(), cfg, receivertest.NewNopCreateSettings())
	scraper.client = newMockClient(t)

	actualMetrics, err := scraper.scrape(context.Background())
	require.NoError(t, err)
	// validate metrics are getting emitted
	expectedFile := filepath.Join("testdata", "scraper", "expected.yaml")
	// golden.WriteMetrics(t, expectedFile, actualMetrics)
	expectedMetrics, err := golden.ReadMetrics(expectedFile)
	require.NoError(t, err)

	require.NoError(t, pmetrictest.CompareMetrics(expectedMetrics, actualMetrics,
		pmetrictest.IgnoreMetricDataPointsOrder(), pmetrictest.IgnoreStartTimestamp(), pmetrictest.IgnoreTimestamp()))
}

func newMockClient(t *testing.T) client {

	mockReceiverClient := &mocks.MockReceiverClient{}

	dummyTopicName := utils.TopicName{}

	dummyTopicStats := map[*utils.TopicName]utils.TopicStats{&dummyTopicName: {
		MsgRateIn:           0.0,
		MsgRateOut:          0.0,
		MsgThroughputIn:     0.0,
		MsgThroughputOut:    0.0,
		AverageMsgSize:      0.0,
		StorageSize:         504,
		Publishers:          []utils.PublisherStats{},
		Subscriptions:       map[string]utils.SubscriptionStats{},
		Replication:         map[string]utils.ReplicatorStats{},
		DeDuplicationStatus: "Disabled",
	}}

	mockReceiverClient.On("GetTenants").Return([]string{"public", "pulsar"}, nil)
	mockReceiverClient.On("GetNameSpaces", mock.Anything).Return([]string{"Hello", "World"}, nil)
	mockReceiverClient.On("GetTopics", mock.Anything).Return([]string{"my-topic"}, nil)
	mockReceiverClient.On("GetTopicStats", mock.Anything).Return(dummyTopicStats, nil)

	return mockReceiverClient
}
