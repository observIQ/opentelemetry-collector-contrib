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
	// This value has been validated against Google SecOps API documentation.
	defaultBatchRequestSizeLimitGRPC = 4000000

	// defaultBatchRequestSizeLimitHTTP is the maximum batch request size for HTTP (DataPlane API).
	// Set to 4,000,000 bytes (4 MB) to match Google SecOps backend limits.
	// Requests exceeding this limit are automatically split into multiple requests.
	// This value has been validated against Google SecOps API documentation.
	defaultBatchRequestSizeLimitHTTP = 4000000
)

// createDefaultConfig creates the default configuration for the exporter.
// The defaults have been validated against Google SecOps API requirements:
//   - Timeout: 5 seconds (OTel default) - adequate for cloud API latency
//   - Batch sizes: 4 MB for both gRPC and HTTP - matches SecOps backend limits
//   - Queue/Retry: OTel defaults - provide robust async sending and error handling
//
// See BATCH_TIMEOUT_VALIDATION.md for detailed validation methodology and results.
func createDefaultConfig() component.Config {
	return &Config{
		Protocol:                  protocolGRPC,
		TimeoutConfig:             exporterhelper.NewDefaultTimeoutConfig(),                    // 5 seconds - validated for Google SecOps API
		QueueBatchConfig:          configoptional.Some(exporterhelper.NewDefaultQueueConfig()), // Default queue: 1000 items
		BackOffConfig:             configretry.NewDefaultBackOffConfig(),                       // Default retry: 5s initial, 30s max
		OverrideLogType:           true,
		Compression:               noCompression,
		CollectAgentMetrics:       true,
		Endpoint:                  defaultEndpoint,
		BatchRequestSizeLimitGRPC: defaultBatchRequestSizeLimitGRPC, // 4 MB - validated against SecOps backend limits
		BatchRequestSizeLimitHTTP: defaultBatchRequestSizeLimitHTTP, // 4 MB - validated against SecOps backend limits
		LogErroredPayloads:        false,
		ValidateLogTypes:          false,
	}
}

// createLogsExporter creates a new log exporter based on this config.
func createLogsExporter(
	ctx context.Context,
	params exporter.Settings,
	cfg component.Config,
) (exp exporter.Logs, err error) {
	t, err := metadata.NewTelemetryBuilder(params.TelemetrySettings)
	if err != nil {
		return nil, fmt.Errorf("create telemetry builder: %w", err)
	}

	c := cfg.(*Config)
	if c.Protocol == protocolHTTPS {
		exp, err = newHTTPExporter(c, params, t)
	} else {
		exp, err = newGRPCExporter(c, params, t)
	}
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewLogs(
		ctx,
		params,
		c,
		exp.ConsumeLogs,
		exporterhelper.WithCapabilities(exp.Capabilities()),
		exporterhelper.WithTimeout(c.TimeoutConfig),
		exporterhelper.WithQueue(c.QueueBatchConfig),
		exporterhelper.WithRetry(c.BackOffConfig),
		exporterhelper.WithStart(exp.Start),
		exporterhelper.WithShutdown(exp.Shutdown),
	)
}
