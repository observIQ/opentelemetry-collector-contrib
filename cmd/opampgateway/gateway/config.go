package gateway

import (
	"fmt"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// The upstream OpAMP server to connect to
	ServerEndpoint string      `yaml:"server_endpoint"`
	ServerHeaders  http.Header `yaml:"server_headers,omitempty"`

	GatewayListenAddress string `yaml:"gateway_endpoint"`

	ConnectionLimit int `yaml:"connection_limit"`
}

func (c *Config) Load(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return c, nil
}
