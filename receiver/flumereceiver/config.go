// Copyright  OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flumereceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/flumereceiver"

import (
	"fmt"
	"net/url"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/flumereceiver/internal/metadata"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
	MetricsBuilderConfig                    metadata.MetricsBuilderConfig `mapstructure:",squash"`
}

var (
	defaultProtocol = "http://"
	defaultHost     = "localhost"
	defaultPort     = "41414"
	defaultPath     = "metrics"
	defaultEndpoint = fmt.Sprintf("%s%s:%s/%s", defaultProtocol, defaultHost, defaultPort, defaultPath)
)

func (cfg *Config) Validate() error {
	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: '%s': %w", cfg.Endpoint, err)
	}

	if u.Hostname() == "" {
		return fmt.Errorf("missing hostname: '%s'", cfg.Endpoint)
	}

	if u.Port() == "" {
		return fmt.Errorf("missing port: '%s'", cfg.Endpoint)
	}

	if u.Path != "/metrics" {
		return fmt.Errorf("path must be 'metrics': '%s'", cfg.Endpoint)
	}

	return nil
}
