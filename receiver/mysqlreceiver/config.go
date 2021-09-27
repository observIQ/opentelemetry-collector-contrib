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

package mysqlreceiver

import (
	"errors"

	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	Username                                string `mapstructure:"username"`
	Password                                string `mapstructure:"password"`
	Database                                string `mapstructure:"database"`
	Endpoint                                string `mapstructure:"endpoint"`
}

// Errors for missing required config parameters.
const (
	ErrNoUsername = "invalid config: missing username"
	ErrNoPassword = "invalid config: missing password" // #nosec G101 - not hardcoded credentials
)

func (cfg *Config) Validate() error {
	var errs []error
	if cfg.Username == "" {
		errs = append(errs, errors.New(ErrNoUsername))
	}
	if cfg.Password == "" {
		errs = append(errs, errors.New(ErrNoPassword))
	}
	return multierr.Combine(errs...)
}
