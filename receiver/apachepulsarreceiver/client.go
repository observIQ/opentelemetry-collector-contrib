package apachepulsarreceiver

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type client interface {
	GetTenants()
	// GetNameSpaces(tenant string)
	// GetTopics()
	// GetTopicStats()
	// Connect() error
	// Close() error
}

type apachePulsarClient struct {
	client       *http.Client
	hostEndpoint string
	logger       *zap.Logger
}

func newClient(cfg *Config, host component.Host, settings component.TelemetrySettings, logger *zap.Logger) (client, error) {
	httpClient, err := cfg.ToClient(host, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP Client: %w", err)
	}

	return &apachePulsarClient{
		client:       httpClient,
		hostEndpoint: cfg.Endpoint,
		logger:       logger,
	}, nil
}

func (c *apachePulsarClient) GetTenants() {
	// tenants.List() returns a slice of strings and an error
	tenants, err := pulsarctl.tenants.List()
	if err != nil {
		return err, err.Error
	} else {
		return tenants
	}
}

// func (c *apachePulsarClient) GetNameSpaces(tenant string) {
// 	// GetNamespaces(tenant string) returns a list of all namespaces for a given tenant
// 	namespaces := pulsar.GetNameSpaces(tenants[0])

// 	// namespace.List(namespace string) returns a list of topics under a given namespace
// 	topics := pulsar.namespace.List(namespaces[0])

// 	// topic.GetStats(utils.TopicName) returns the stats for a topic
// 	stats := topics[0].GetStats(topics[0].TopicName)

// 	fmt.Println(stats)
// }

// func (c *apachePulsarClient) GetTopics() {

// }

// func (c *apachePulsarClient) GetTopicStats() {

// }

// func (c *apachePulsarClient) Connect() {
// }

// func (c *apachePulsarClient) Close() {

// }
