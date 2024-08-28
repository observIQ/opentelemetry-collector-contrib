// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package transformprocessor

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processortest"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/plogtest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor/internal/common"
)

func TestFlattenDataDisabledByDefault(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	assert.False(t, oCfg.FlattenData)
	assert.NoError(t, oCfg.Validate())
}

func TestFlattenDataRequiresGate(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.FlattenData = true
	assert.Equal(t, errFlatLogsGateDisabled, oCfg.Validate())
}

func TestProcessLogsWithoutFlatten(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.LogStatements = []common.ContextStatements{
		{
			Context: "log",
			Statements: []string{
				`set(resource.attributes["host.name"], attributes["host.name"])`,
				`delete_key(attributes, "host.name")`,
			},
		},
	}
	sink := new(consumertest.LogsSink)
	p, err := factory.CreateLogsProcessor(context.Background(), processortest.NewNopSettings(), oCfg, sink)
	require.NoError(t, err)

	input, err := golden.ReadLogs(filepath.Join("testdata", "logs", "input.yaml"))
	require.NoError(t, err)
	expected, err := golden.ReadLogs(filepath.Join("testdata", "logs", "expected-without-flatten.yaml"))
	require.NoError(t, err)

	assert.NoError(t, p.ConsumeLogs(context.Background(), input))

	actual := sink.AllLogs()
	require.Len(t, actual, 1)

	assert.NoError(t, plogtest.CompareLogs(expected, actual[0]))
}

func TestProcessLogsWithFlatten(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.FlattenData = true
	oCfg.LogStatements = []common.ContextStatements{
		{
			Context: "log",
			Statements: []string{
				`set(resource.attributes["host.name"], attributes["host.name"])`,
				`delete_key(attributes, "host.name")`,
			},
		},
	}
	sink := new(consumertest.LogsSink)
	p, err := factory.CreateLogsProcessor(context.Background(), processortest.NewNopSettings(), oCfg, sink)
	require.NoError(t, err)

	input, err := golden.ReadLogs(filepath.Join("testdata", "logs", "input.yaml"))
	require.NoError(t, err)
	expected, err := golden.ReadLogs(filepath.Join("testdata", "logs", "expected-with-flatten.yaml"))
	require.NoError(t, err)

	assert.NoError(t, p.ConsumeLogs(context.Background(), input))

	actual := sink.AllLogs()
	require.Len(t, actual, 1)

	assert.NoError(t, plogtest.CompareLogs(expected, actual[0]))
}

func BenchmarkLogsWithoutFlatten(b *testing.B) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.LogStatements = []common.ContextStatements{
		{
			Context: "log",
			Statements: []string{
				`set(resource.attributes["host.name"], attributes["host.name"])`,
				`delete_key(attributes, "host.name")`,
			},
		},
	}
	sink := new(consumertest.LogsSink)
	p, err := factory.CreateLogsProcessor(context.Background(), processortest.NewNopSettings(), oCfg, sink)
	require.NoError(b, err)

	input, err := golden.ReadLogs(filepath.Join("testdata", "logs", "input.yaml"))
	require.NoError(b, err)

	for n := 0; n < b.N; n++ {
		assert.NoError(b, p.ConsumeLogs(context.Background(), input))
	}
}

func BenchmarkLogsWithFlatten(b *testing.B) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.FlattenData = true
	oCfg.LogStatements = []common.ContextStatements{
		{
			Context: "log",
			Statements: []string{
				`set(resource.attributes["host.name"], attributes["host.name"])`,
				`delete_key(attributes, "host.name")`,
			},
		},
	}
	sink := new(consumertest.LogsSink)
	p, err := factory.CreateLogsProcessor(context.Background(), processortest.NewNopSettings(), oCfg, sink)
	require.NoError(b, err)

	input, err := golden.ReadLogs(filepath.Join("testdata", "logs", "input.yaml"))
	require.NoError(b, err)

	for n := 0; n < b.N; n++ {
		assert.NoError(b, p.ConsumeLogs(context.Background(), input))
	}
}

func BenchmarkSetLogRecords(b *testing.B) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.LogStatements = []common.ContextStatements{
		{
			Context: "log",
			Statements: []string{
				`set(attributes["my_attribute"], "my_value")`,
			},
		},
	}
	sink := new(consumertest.LogsSink)

	set := processortest.NewNopSettings()
	spanRecorder := tracetest.NewSpanRecorder()
	set.TelemetrySettings.TracerProvider = trace.NewTracerProvider(trace.WithSpanProcessor(spanRecorder))

	p, err := factory.CreateLogsProcessor(context.Background(), set, oCfg, sink)
	require.NoError(b, err)

	logs := plog.NewLogs()
	scopeLogs := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()

	for i := range 200 {
		lr := scopeLogs.LogRecords().AppendEmpty()
		lr.Body().SetStr(fmt.Sprintf("Log body: %d", i))
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		assert.NoError(b, p.ConsumeLogs(context.Background(), logs))
	}
}
