// Copyright  The OpenTelemetry Authors
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

package expvarreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/expvarreceiver"

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

const (
	typeStr         = "expvar"
	defaultEndpoint = "http://localhost:8000/debug/vars"
	defaultTimeout  = 3 * time.Second
)

func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		typeStr,
		newDefaultConfig,
		component.WithMetricsReceiver(newMetricsReceiver))
}

func newMetricsReceiver(
	ctx context.Context,
	settings component.ReceiverCreateSettings,
	rCfg config.Receiver,
	metrics consumer.Metrics,
) (component.MetricsReceiver, error) {
	return nil, fmt.Errorf("not implemented")
}

func newDefaultConfig() config.Receiver {
	return &Config{
		ScraperControllerSettings: scraperhelper.NewDefaultScraperControllerSettings(typeStr),
		HTTP: &confighttp.HTTPClientSettings{
			Endpoint: defaultEndpoint,
			Timeout:  defaultTimeout,
		},
		MetricsConfig: []MetricConfig{},
	}
}
