package apachepulsarreceiver

import (
	"errors"
	"fmt"

	pulsarctl "github.com/streamnative/pulsarctl/pkg/pulsar"
	"github.com/streamnative/pulsarctl/pkg/pulsar/common"
	utils "github.com/streamnative/pulsarctl/pkg/pulsar/utils"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

var (
	errEmptyParam         = errors.New(`cannot perform this operation on an empty array`)
	errTopicNameNotFound  = errors.New(`could not find topic name for a topic`)
	errTopicStatsNotFound = errors.New(`error occurred while fetching stats for a topic`)
)

type client interface {
	GetTenants() ([]string, error)
	GetNameSpaces(tenants []string) ([]string, error)
	GetTopics(namespaces []string) ([]string, error)
	GetTopicStats(topics []string) ([]utils.TopicStats, error)
}

type apachePulsarClient struct {
	client       pulsarctl.Client
	hostEndpoint string
	logger       *zap.Logger
}

func newClient(cfg *Config, host component.Host, settings component.TelemetrySettings, logger *zap.Logger) (client, error) {
	client, err := pulsarctl.New(&common.Config{
		WebServiceURL: cfg.Endpoint,
	})
	if err != nil {
		return nil, err
	}

	return &apachePulsarClient{
		client:       client,
		hostEndpoint: cfg.Endpoint,
		logger:       logger,
	}, nil
}

func (c *apachePulsarClient) GetTenants() ([]string, error) {
	// tenants.List() returns a slice of strings and an error
	tenants := c.client.Tenants()
	return tenants.List()
}

func (c *apachePulsarClient) GetNameSpaces(tenants []string) ([]string, error) {
	namespaceInterface := c.client.Namespaces()
	var namespaces = []string{}
	for i := range tenants {
		// GetNamespaces(tenant string) returns a list of all namespaces for a given tenant
		tenantNamespaces, err := namespaceInterface.GetNamespaces(tenants[i])
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, tenantNamespaces...)
	}
	return namespaces, nil
}

func (c *apachePulsarClient) GetTopics(namespaces []string) ([]string, error) {
	namespaceInterface := c.client.Namespaces()
	var topics = []string{}
	// namespace.GetTopics(namespace string) returns a list of topics under a given namespace
	for i := range namespaces {
		namespaceTopics, err := namespaceInterface.GetTopics(namespaces[i])
		if err != nil {
			return nil, err
		}
		topics = append(topics, namespaceTopics...)
	}
	return topics, nil
}

func (c *apachePulsarClient) GetTopicStats(topics []string) ([]utils.TopicStats, error) {
	topicInterface := c.client.Topics()
	var statsList = []utils.TopicStats{}
	for i := range topics {
		name, err := utils.GetTopicName(topics[i])
		if err != nil {
			return nil, errTopicNameNotFound
		}
		// topic.GetStats(utils.TopicName) returns the stats for a topic
		stats, err := topicInterface.GetStats(*name)
		if err != nil {
			return nil, errTopicStatsNotFound
		}
		fmt.Println(stats)

		statsList = append(statsList, stats)
	}
	return statsList, nil
}
