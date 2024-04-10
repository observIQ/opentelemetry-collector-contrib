// Copyright  observIQ, Inc.
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

package snapshotprocessor

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/open-telemetry/opamp-go/client/types"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampextension"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// snapshotRequest specifies what snapshots to collect
type snapshotRequest struct {
	// Processor is the full ComponentID of the snapshot processor
	Processor component.ID `yaml:"processor"`

	// PipelineType will be "logs", "metrics", or "traces"
	PipelineType string `yaml:"pipeline_type"`

	// SearchQuery is an optional query string that will filter telemetry
	// such that only telemetry containing the string is reported.
	SearchQuery *string `yaml:"search_query"`

	// MinimumTimestamp is the minimum timestamp used to filter telemetry such that only telemetry
	// with a timestamp higher than specified will be reported.
	MinimumTimestamp *time.Time `yaml:"minimum_timestamp"`
}

const (
	snapshotCapability  = "com.bindplane.snapshot"
	snapshotRequestType = "requestSnapshot"
	snapshotReportType  = "reportSnapshot"
)

type snapshotProcessor struct {
	logger *zap.Logger

	processorID      component.ID
	enabled          bool
	opampExtensionID component.ID

	customCapability opampextension.CustomCapability

	logBuffer    *LogBuffer
	metricBuffer *MetricBuffer
	traceBuffer  *TraceBuffer
}

// newSnapshotProcessor creates a new snapshot processor
func newSnapshotProcessor(logger *zap.Logger, cfg *Config, processorID component.ID) *snapshotProcessor {
	return &snapshotProcessor{
		logger: logger,

		enabled:          cfg.Enabled,
		processorID:      processorID,
		opampExtensionID: cfg.OpAMP,

		logBuffer:    NewLogBuffer(100),
		metricBuffer: NewMetricBuffer(100),
		traceBuffer:  NewTraceBuffer(100),
	}
}

func (sp *snapshotProcessor) start(_ context.Context, host component.Host) error {
	ext, ok := host.GetExtensions()[sp.opampExtensionID]
	if !ok {
		return fmt.Errorf("opamp extension %q does not exist", sp.opampExtensionID)
	}

	registry, ok := ext.(opampextension.CustomCapabilityRegistry)
	if !ok {
		return fmt.Errorf("extension %q is not an custom message registry", sp.opampExtensionID)
	}

	var err error
	sp.customCapability, err = registry.Register(snapshotCapability, sp.processSnapshotMessage)
	if err != nil {
		return fmt.Errorf("register custom capability: %w", err)
	}

	return nil
}

func (sp *snapshotProcessor) processSnapshotMessage(cm *protobufs.CustomMessage) {
	if cm.Type != snapshotRequestType {
		return
	}

	var req snapshotRequest
	err := yaml.Unmarshal(cm.Data, &req)
	if err != nil {
		sp.logger.Error("Got invalid snapshot request.", zap.Error(err))
		return
	}

	if req.Processor == sp.processorID {
		return
	}

	var reportMsg []byte
	switch req.PipelineType {
	case "logs":
		reportMsg, err = sp.logBuffer.ConstructPayload(req.SearchQuery, req.MinimumTimestamp)
	case "metrics":
		reportMsg, err = sp.metricBuffer.ConstructPayload(req.SearchQuery, req.MinimumTimestamp)
	case "traces":
		reportMsg, err = sp.traceBuffer.ConstructPayload(req.SearchQuery, req.MinimumTimestamp)
	default:
		sp.logger.Error("Invalid pipeline type in snapshot request.", zap.String("PipelineType", req.PipelineType))
		return
	}

	if err != nil {
		sp.logger.Error("Failed to construct snapshot payload.", zap.Error(err))
		return
	}

	compressedMsg, err := compress(reportMsg)
	if err != nil {
		sp.logger.Error("Failed to compress snapshot payload.", zap.Error(err))
		return
	}

	for {
		msgSendChan, err := sp.customCapability.SendMessage(snapshotReportType, compressedMsg)
		switch {
		case errors.Is(err, types.ErrCustomMessagePending):
			// Wait until message is ready to send, then try again
			sp.logger.Debug("Custom message pending, will try sending again after message is clear.")
			<-msgSendChan
		case err != nil:
			sp.logger.Error("Failed to send snapshot payload message.", zap.Error(err))
			return
		default: // Message is scheduled to send
			return
		}
	}
}

func (sp *snapshotProcessor) processTraces(_ context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	if sp.enabled {
		newTraces := ptrace.NewTraces()
		td.CopyTo(newTraces)
		sp.traceBuffer.Add(td)
	}

	return td, nil
}

func (sp *snapshotProcessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	if sp.enabled {
		newLogs := plog.NewLogs()
		ld.CopyTo(newLogs)
		sp.logBuffer.Add(ld)
	}

	return ld, nil
}

func (sp *snapshotProcessor) processMetrics(_ context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	if sp.enabled {
		newMetrics := pmetric.NewMetrics()
		md.CopyTo(newMetrics)
		sp.metricBuffer.Add(md)
	}

	return md, nil
}

// compress gzip compresses the input data
func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
