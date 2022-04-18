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

//go:build windows
// +build windows

package windowseventlogreceiver

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/service/servicetest"
	"golang.org/x/sys/windows/svc/eventlog"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza"
)

func TestDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	require.NotNil(t, cfg, "failed to create default config")
	require.NoError(t, configtest.CheckConfigStruct(cfg))
}

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfigAndValidate(filepath.Join("testdata", "config.yaml"), factories)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 1)

	assert.Equal(t, createTestConfig(), cfg.Receivers[config.NewComponentID("windowseventlog")])
}

func TestCreateWithInvalidInputConfig(t *testing.T) {
	t.Parallel()

	cfg := &WindowsLogConfig{
		BaseConfig: stanza.BaseConfig{},
		Input: stanza.InputConfig{
			"start_at": "end",
		},
	}
	cfg.Input["include"] = "not an array"

	_, err := NewFactory().CreateLogsReceiver(
		context.Background(),
		componenttest.NewNopReceiverCreateSettings(),
		cfg,
		new(consumertest.LogsSink),
	)
	require.Error(t, err, "receiver creation should fail if given invalid input config")
}

func TestReadWindowsEventLogger(t *testing.T) {
	ctx := context.Background()
	factory := NewFactory()
	createSettings := componenttest.NewNopReceiverCreateSettings()
	cfg := createTestConfig()
	sink := new(consumertest.LogsSink)

	receiver, err := factory.CreateLogsReceiver(ctx, createSettings, cfg, sink)
	require.NoError(t, err)

	err = receiver.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)
	defer receiver.Shutdown(ctx)

	src := "otel"
	err = eventlog.InstallAsEventCreate(src, eventlog.Info|eventlog.Warning|eventlog.Error)
	require.NoError(t, err)
	defer eventlog.Remove(src)

	logger, err := eventlog.Open(src)
	require.NoError(t, err)
	defer logger.Close()

	err = logger.Info(10, "Test log")
	require.NoError(t, err)

	logsReceived := func() bool {
		return sink.LogRecordCount() == 1
	}

	require.Eventually(t, logsReceived, 1*time.Minute, 200*time.Millisecond)
	results := sink.AllLogs()
	require.Len(t, results, 1)

	records := results[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords()
	require.Equal(t, records.Len(), 1)

	record := records.At(0)
	body := record.Body().MapVal().AsRaw()
	require.Equal(t, "Test log", body["message"])

	eventID := body["event_id"]
	require.NotNil(t, eventID)

	eventIDMap, ok := eventID.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, int64(10), eventIDMap["id"])
}

func createTestConfig() *WindowsLogConfig {
	return &WindowsLogConfig{
		BaseConfig: stanza.BaseConfig{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
			Operators:        stanza.OperatorConfigs{},
		},
		Input: stanza.InputConfig{
			"channel":  "application",
			"start_at": "end",
		},
	}
}
