package apachepulsarreceiver

import (
	"context"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachepulsarreceiver/internal/mocks"
	utils "github.com/streamnative/pulsarctl/pkg/pulsar/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {

	cfg := newDefaultConfig().(*Config)
	require.NoError(t, component.ValidateConfig(cfg))

	scraper := newScraper(zap.NewNop(), cfg, receivertest.NewNopCreateSettings())
	scraper.client = newMockClient(t)

	err := scraper.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

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
	// mockTenant := &mocks.MockTenants{}
	// mockTenant.On("List").Return([]string{"public", "pulsar"}, nil)
	// mockClient := &mocks.MockClient{}
	// mockClient.On("Tenants").Return(mockTenant)

	// returnedNamespaces := []string{"Hello", "World"}
	// mockNamespace := &mocks.MockNamespaces{}
	// mockNamespace.On("GetNamespaces", mock.Anything).Return(returnedNamespaces, nil)
	// mockNamespace.On("GetTopics", mock.Anything).Return([]string{"my-topic"}, nil)
	// mockClient.On("Namespaces").Return(mockNamespace)

	// mockTopic := &mocks.MockTopics{}
	// mockTopic.On("GetStats", mock.Anything).Return(utils.TopicStats{
	// 	MsgRateIn:           0.0,
	// 	MsgRateOut:          0.0,
	// 	MsgThroughputIn:     0.0,
	// 	MsgThroughputOut:    0.0,
	// 	AverageMsgSize:      0.0,
	// 	StorageSize:         504,
	// 	Publishers:          []utils.PublisherStats{},
	// 	Subscriptions:       map[string]utils.SubscriptionStats{},
	// 	Replication:         map[string]utils.ReplicatorStats{},
	// 	DeDuplicationStatus: "Disabled",
	// }, nil)
	// mockClient.On("Topics").Return(mockTopic)

	mockReceiverClient := &mocks.MockReceiverClient{}

	mockReceiverClient.On("GetTenants").Return([]string{"public", "pulsar"}, nil)
	mockReceiverClient.On("GetNameSpaces", mock.Anything).Return([]string{"Hello", "World"}, nil)
	mockReceiverClient.On("GetTopics", mock.Anything).Return([]string{"my-topic"}, nil)
	mockReceiverClient.On("GetTopicStats", mock.Anything).Return(utils.TopicStats{
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
	}, nil)

	return mockReceiverClient
}
