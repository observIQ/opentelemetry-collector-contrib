// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package windows // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/windows"

import (
	"fmt"

	"go.opentelemetry.io/collector/component"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
)

func init() {
	operator.Register(operatorType, func() operator.Builder { return NewConfig() })
}

// Build will build a windows event log operator.
func (c *Config) Build(set component.TelemetrySettings) (operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(set)
	if err != nil {
		return nil, err
	}

	if c.Channel == "" {
		return nil, fmt.Errorf("missing required `channel` field")
	}

	if c.MaxReads < 1 {
		return nil, fmt.Errorf("the `max_reads` field must be greater than zero")
	}

	if c.StartAt != "end" && c.StartAt != "beginning" {
		return nil, fmt.Errorf("the `start_at` field must be set to `beginning` or `end`")
	}

	for _, group := range c.RemoteGroups {
		if group.Credentials.Username == "" || group.Credentials.Password == "" {
			return nil, fmt.Errorf("each remote group must have non-empty `username` and `password`")
		}

		if len(group.Servers) == 0 {
			return nil, fmt.Errorf("each remote group must have at least one `server`")
		}

		for _, server := range group.Servers {
			if server == "" {
				return nil, fmt.Errorf("server names cannot be empty")
			}
		}
	}

	return &Input{
		InputOperator:    inputOperator,
		buffer:           NewBuffer(),
		channel:          c.Channel,
		maxReads:         c.MaxReads,
		startAt:          c.StartAt,
		pollInterval:     c.PollInterval,
		raw:              c.Raw,
		excludeProviders: c.ExcludeProviders,
		remoteGroups:     c.RemoteGroups,
	}, nil
}
