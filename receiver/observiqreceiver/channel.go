package observiqreceiver

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

import (
	"context"
	"fmt"

	"github.com/observiq/carbon/entry"
	"github.com/observiq/carbon/operator"
	"github.com/observiq/carbon/operator/helper"
)

func init() {
	operator.Register("channel_output", func() operator.Builder { return NewChannelOutputConfig("") })
}

// NewChannelOutputConfig creates new output config
func NewChannelOutputConfig(operatorID string) *ChannelOutputConfig {
	return &ChannelOutputConfig{
		OutputConfig: helper.NewOutputConfig(operatorID, "channel_output"),
	}
}

// ChannelOutputConfig is the configuration of an channel output operator
type ChannelOutputConfig struct {
	helper.OutputConfig `yaml:",inline"`
}

// Build will build an channel output operator
func (c ChannelOutputConfig) Build(context operator.BuildContext) (operator.Operator, error) {
	outputOperator, err := c.OutputConfig.Build(context)
	if err != nil {
		return nil, err
	}

	buildParam, ok := context.Parameters["otel_output_chan"]
	if !ok {
		return nil, fmt.Errorf("otel_output_chan not found in build context")
	}

	outChan, ok := buildParam.(chan *entry.Entry)
	if !ok {
		return nil, fmt.Errorf("otel_output_chan not of correct type")
	}

	channelOutput := &ChannelOutput{
		OutputOperator: outputOperator,
		outChan:        outChan,
	}

	return channelOutput, nil
}

// ChannelOutput is an operator that sends entries to channel
type ChannelOutput struct {
	helper.OutputOperator
	outChan chan *entry.Entry
}

// Process will write an entry to the output channel
func (c *ChannelOutput) Process(ctx context.Context, entry *entry.Entry) error {
	c.outChan <- entry
	return nil
}
