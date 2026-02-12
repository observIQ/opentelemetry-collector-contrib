// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlesecopsexporter

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlesecopsexporter/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/plog"
)

// NewFactory creates a new SecOps exporter factory.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		metadata.Type,
		createDefaultConfig,
		exporter.WithLogs(createLogsExporter, metadata.LogsStability))
}

const (
	// defaultEndpoint is the default endpoint for the Legacy Ingestion API (gRPC protocol)
	defaultEndpoint = "malachiteingestion-pa.googleapis.com"

	// defaultBatchRequestSizeLimitGRPC is the maximum batch request size for gRPC (Legacy Ingestion API).
	// Set to 4,000,000 bytes (4 MB) to match Google SecOps backend limits.
	// Requests exceeding this limit are automatically split into multiple requests.
	defaultBatchRequestSizeLimitGRPC = 4000000

	// defaultBatchRequestSizeLimitHTTP is the maximum batch request size for HTTP (DataPlane API).
	// Set to 4,000,000 bytes (4 MB) to match Google SecOps backend limits.
	// Requests exceeding this limit are automatically split into multiple requests.
	defaultBatchRequestSizeLimitHTTP = 4000000
)

// createDefaultConfig creates the default configuration for the exporter.
func createDefaultConfig() component.Config {
	return &Config{
		Protocol:                  protocolGRPC,
		TimeoutConfig:             exporterhelper.NewDefaultTimeoutConfig(),
		QueueBatchConfig:          configoptional.Some(exporterhelper.NewDefaultQueueConfig()),
		BackOffConfig:             configretry.NewDefaultBackOffConfig(),
		OverrideLogType:           true,
		Compression:               noCompression,
		CollectAgentMetrics:       true,
		Endpoint:                  defaultEndpoint,
		BatchRequestSizeLimitGRPC: defaultBatchRequestSizeLimitGRPC,
		BatchRequestSizeLimitHTTP: defaultBatchRequestSizeLimitHTTP,
		LogErroredPayloads:        false,
		ValidateLogTypes:          false,
	}
}

// createLogsExporter creates a new log exporter based on this config.
func createLogsExporter(
	ctx context.Context,
	params exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {
	c := cfg.(*Config)
	return exporterhelper.NewLogs(
		ctx,
		params,
		c,
		func(_ context.Context, _ plog.Logs) error {
			return fmt.Errorf("not yet implemented")
		},
		exporterhelper.WithTimeout(c.TimeoutConfig),
		exporterhelper.WithQueue(c.QueueBatchConfig),
		exporterhelper.WithRetry(c.BackOffConfig),
	)
}
