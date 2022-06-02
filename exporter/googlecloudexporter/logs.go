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

// Package googlecloudexporter contains the wrapper for OpenTelemetry-GoogleCloud
// exporter to be used in opentelemetry-collector.
package googlecloudexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter"

import (
	"context"
	"fmt"
	"strings"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// logsExporter contains what is necessary to export logs to google
type logsExporter struct {
	Config         *Config
	requestBuilder RequestBuilder
	logClient      LogClient
	logger         *zap.Logger
}

// Shutdown cleans up anything related to the logsExporter
func (le *logsExporter) Shutdown(context.Context) error {
	return le.logClient.Close()
}

func setVersionInUserAgent(cfg *Config, version string) {
	cfg.UserAgent = strings.ReplaceAll(cfg.UserAgent, "{{version}}", version)
}

func generateClientOptions(cfg *Config) ([]option.ClientOption, error) {
	var copts []option.ClientOption
	// option.WithUserAgent is used by the Trace exporter, but not the Metric exporter (see comment below)
	if cfg.CollectorConfig.UserAgent != "" {
		copts = append(copts, option.WithUserAgent(cfg.CollectorConfig.UserAgent))
	}
	if cfg.LogConfig.Endpoint != "" {
		if cfg.LogConfig.UseInsecure {
			// option.WithGRPCConn option takes precedent over all other supplied options so the
			// following user agent will be used by both exporters if we reach this branch
			var dialOpts []grpc.DialOption
			if cfg.CollectorConfig.UserAgent != "" {
				dialOpts = append(dialOpts, grpc.WithUserAgent(cfg.CollectorConfig.UserAgent))
			}
			conn, err := grpc.Dial(cfg.LogConfig.Endpoint, append(dialOpts, grpc.WithStatsHandler(&ocgrpc.ClientHandler{}), grpc.WithTransportCredentials(insecure.NewCredentials()))...)
			if err != nil {
				return nil, fmt.Errorf("cannot configure grpc conn: %w", err)
			}
			copts = append(copts, option.WithGRPCConn(conn))
		} else {
			copts = append(copts, option.WithEndpoint(cfg.LogConfig.Endpoint))
		}
	}
	if cfg.LogConfig.GetClientOptions != nil {
		copts = append(copts, cfg.LogConfig.GetClientOptions()...)
	}
	return copts, nil
}

// newGoogleCloudLogsExporter creates a LogsExporter which is able to push logs to google cloud
func newGoogleCloudLogsExporter(cfg *Config, set component.ExporterCreateSettings) (component.LogsExporter, error) {
	setVersionInUserAgent(cfg, set.BuildInfo.Version)

	copts, err := generateClientOptions(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate client options: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	logClient, err := newLogClient(ctx, copts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create log client: %w", err)
	}

	entryBuilder := &GoogleEntryBuilder{
		MaxEntrySize: defaultMaxEntrySize,
		ProjectID:    cfg.CollectorConfig.ProjectID,
		NameFields:   cfg.LogConfig.NameFields,
	}
	requestBuilder := &GoogleRequestBuilder{
		MaxRequestSize: defaultMaxRequestSize,
		ProjectID:      cfg.CollectorConfig.ProjectID,
		EntryBuilder:   entryBuilder,
		SugaredLogger:  set.Logger.Sugar(),
	}

	lExp := &logsExporter{
		logger:         set.Logger,
		requestBuilder: requestBuilder,
		logClient:      logClient,
	}

	return exporterhelper.NewLogsExporter(
		cfg,
		set,
		lExp.pushLogs,
		exporterhelper.WithShutdown(lExp.Shutdown),
		exporterhelper.WithTimeout(cfg.TimeoutSettings),
		exporterhelper.WithQueue(cfg.QueueSettings),
		exporterhelper.WithRetry(cfg.RetrySettings))
}

// pushLogs converts all plog.Logs to google logging api format and creates them in google cloud
func (le *logsExporter) pushLogs(ctx context.Context, logs plog.Logs) error {
	len := logs.ResourceLogs().Len()
	if len == 0 {
		return nil
	}

	requests := le.requestBuilder.Build(&logs)

	for _, request := range requests {
		_, err := le.logClient.WriteLogEntries(ctx, request)
		if err != nil {
			return fmt.Errorf("failed to send all logs: %w", err)
		}
	}

	return nil
}
