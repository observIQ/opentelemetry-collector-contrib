package elasticsearchreceiver

import (
	"errors"
	"fmt"
	"net/url"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	confighttp.HTTPClientSettings           `mapstructure:",squash"`

	Password string `mapstructure:"password"`
	Username string `mapstructure:"username"`
}

var (
	defaultEndpoint = "http://localhost:9200"
)

var (
	errEndpointBadScheme    = errors.New("endpoint scheme must be http or https")
	errUsernameNotSpecified = errors.New("password was specified, but not username")
	errPasswordNotSpecified = errors.New("username was specified, but not password")
	errEmptyEndpoint        = errors.New("endpoint must be specified")
)

var validSchemes = []string{
	"http",
	"https",
}

func (cfg *Config) Validate() error {
	var combinedErr error
	if err := invalidCredentials(cfg.Username, cfg.Password); err != nil {
		combinedErr = multierr.Append(combinedErr, err)
	}

	if cfg.Endpoint == "" {
		return multierr.Append(combinedErr, errEmptyEndpoint)
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return multierr.Append(
			combinedErr,
			fmt.Errorf("invalid endpoint '%s': %w", cfg.Endpoint, err),
		)
	}

	if !validScheme(u.Scheme) {
		return multierr.Append(combinedErr, errEndpointBadScheme)
	}

	return combinedErr
}

func validScheme(scheme string) bool {
	for _, s := range validSchemes {
		if s == scheme {
			return true
		}
	}

	return false
}

// invalidCredentials returns true if only one username or password is not empty.
func invalidCredentials(username, password string) error {
	if username == "" && password != "" {
		return errUsernameNotSpecified
	}

	if password == "" && username != "" {
		return errPasswordNotSpecified
	}
	return nil
}
