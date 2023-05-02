// Copyright  The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apachesparkreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver"

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver/internal/metadata"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
)

const (
	defaultCollectionInterval = 15 * time.Second
	defaultEndpoint           = "http://localhost:4040"
)

var (
	errInvalidEndpoint = errors.New(`endpoint [%s] must be in the form of <scheme>://<hostname>:<port>`)
)

// Config defines the configuration for the various elements of the receiver agent.
type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	metadata.MetricsBuilderConfig           `mapstructure:",squash"`
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
	WhitelistedApplicationIds               []string `mapstructure:"whitelisted_application_ids"`
}

// Validate validates missing and invalid configuration fields.
func (cfg *Config) Validate() error {
	var err error

	_, parseErr := url.Parse(cfg.Endpoint)
	if parseErr != nil {
		err = multierr.Append(err, fmt.Errorf(errInvalidEndpoint.Error(), cfg.Endpoint))
	}
	return err
}
