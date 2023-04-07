// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apachepulsarreceiver

import (
	"context"
	"errors"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachepulsarreceiver/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

const (
	typeStr   = "apachepulsar"
	stability = component.StabilityLevelDevelopment
)

var errConfigNotPulsar = errors.New("config was not a Pulsar receiver config")

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		newDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, stability))
}

func newDefaultConfig() component.Config {
	return &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 10 * time.Second,
		},
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: defaultEndpoint,
		},
		MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
	}
}

func createMetricsReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	config component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	pulsarConfig, ok := config.(*Config)
	if !ok {
		return nil, errConfigNotPulsar
	}

	pulsarScraper := newScraper(params.Logger, pulsarConfig, params)
	scraper, err := scraperhelper.NewScraper(typeStr, pulsarScraper.scrape,
		scraperhelper.WithStart(pulsarScraper.start))
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewScraperControllerReceiver(&pulsarConfig.ScraperControllerSettings, params,
		consumer, scraperhelper.AddScraper(scraper))

}
