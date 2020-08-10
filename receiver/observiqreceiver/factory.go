// Copyright 2019, OpenTelemetry Authors
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

package observiqreceiver

import (
	"context"
	"fmt"

	observiq "github.com/observiq/carbon/agent"
	obsentry "github.com/observiq/carbon/entry"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	yaml "gopkg.in/yaml.v2"
)

const (
	typeStr = "observiq"
)

// Factory is the factory for the observiq receiver.
type Factory struct {
}

// Ensure this factory adheres to required interface
var _ component.LogsReceiverFactory = (*Factory)(nil)

// Type gets the type of the Receiver config created by this factory
func (f *Factory) Type() configmodels.Type {
	return configmodels.Type(typeStr)
}

// CustomUnmarshaler returns nil even though custom unmarshalling is necessary
func (f *Factory) CustomUnmarshaler() component.CustomUnmarshaler {
	/*
		TODO is the mapstructure requirement absolutely necessary?
		If not, the following custom unmarshal pattern would work
	*/

	// return func(componentViperSection *viper.Viper, intoCfg interface{}) error {

	// 	var cfgMap map[string]interface{}
	// 	componentViperSection.Unmarshal(&cfgMap)

	// 	cfgBytes, err := yaml.Marshal(cfgMap)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to remarshal config: %s", err)
	// 	}

	// 	cfg := &observiq.Config{}
	// 	if err := yaml.UnmarshalStrict(cfgBytes, cfg); err != nil {
	// 		return fmt.Errorf("failed to unmarshal config: %s", err)
	// 	}
	// 	intoCfg = cfg
	// 	return nil
	// }
	return nil
}

// CreateDefaultConfig creates the default configuration for the observiq receiver
func (f *Factory) CreateDefaultConfig() configmodels.Receiver {
	return &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: configmodels.Type(typeStr),
			NameVal: typeStr,
		},
		Pipeline: []interface{}{},
	}
}

// CreateLogsReceiver creates a logs receiver based on provided config
func (f *Factory) CreateLogsReceiver(
	ctx context.Context,
	params component.ReceiverCreateParams,
	cfg configmodels.Receiver,
	nextConsumer consumer.LogsConsumer,
) (component.LogsReceiver, error) {

	/*
		TODO reconcile mapstructure requirement with observiq's decisiont to opt out of mapstructure

		mapstructure has some custom unmarshaling limitations that were solvable with yaml unmarshal hooks.
		However, opentelemetry appears to have committed to requiring mapstructure.
		See: https://github.com/mitchellh/mapstructure/pull/183

		See additional comments in CustomUnmarshaler()
	*/
	rawCfg := cfg.(*Config)

	cfgBytes, err := yaml.Marshal(rawCfg.Pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to remarshal config: %s", err)
	}

	obsCfg := &observiq.Config{}
	if err := yaml.UnmarshalStrict(cfgBytes, &obsCfg.Pipeline); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	logsChan := make(chan *obsentry.Entry)
	// TODO allow configuration of plugins directory and offsets file
	logAgent := observiq.NewLogAgent(obsCfg, params.Logger.Sugar(), "plugins", "offsets.db").
		WithBuildParameter("otel_output_chan", logsChan)

	return &observiqReceiver{
		agent:    logAgent,
		config:   rawCfg, // TODO relax mapstructure requirement or migrate observiq to mapstructure
		logsChan: logsChan,
		consumer: nextConsumer,
		logger:   params.Logger,
		done:     make(chan struct{}),
	}, nil
}
