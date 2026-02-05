// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlesecopsexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlesecopsexporter"

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

func Test_createDefaultConfig(t *testing.T) {
	expectedCfg := &Config{
		TimeoutConfig:             exporterhelper.NewDefaultTimeoutConfig(),
		QueueBatchConfig:          configoptional.Some(exporterhelper.NewDefaultQueueConfig()),
		BackOffConfig:             configretry.NewDefaultBackOffConfig(),
		OverrideLogType:           true,
		Endpoint:                  "malachiteingestion-pa.googleapis.com",
		Compression:               "none",
		CollectAgentMetrics:       true,
		Protocol:                  protocolGRPC,
		BatchRequestSizeLimitGRPC: defaultBatchRequestSizeLimitGRPC,
		BatchRequestSizeLimitHTTP: defaultBatchRequestSizeLimitHTTP,
	}

	actual := createDefaultConfig()
	require.Equal(t, expectedCfg, actual)
}
