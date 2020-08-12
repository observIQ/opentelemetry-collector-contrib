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

package gcploggingexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	vkit "cloud.google.com/go/logging/apiv2"
	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/pdata"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
	sev "google.golang.org/genproto/googleapis/logging/type"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
)

type gcpLoggingExporter struct {
	done chan struct{}

	projectID   string
	credentials *google.Credentials
}

// Ensure this exporter adheres to required interface
var _ component.LogsExporter = (*gcpLoggingExporter)(nil)

func (e *gcpLoggingExporter) getClient(ctx context.Context) (*vkit.Client, error) {
	options := []option.ClientOption{option.WithCredentials(e.credentials)}
	client, err := vkit.NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	return client, nil
}

func (e *gcpLoggingExporter) Start(ctx context.Context, host component.Host) error {

	/*
		TODO this just tests that a client can be created.
		It should save the client on the exporter struct, but
		for some reason this does not persist to ConsumeLogs()
	*/
	_, err := e.getClient(ctx)
	return err
}

// Shutdown is invoked during service shutdown
func (e *gcpLoggingExporter) Shutdown(_ context.Context) error {
	close(e.done)
	return nil
}

func (e *gcpLoggingExporter) ConsumeLogs(ctx context.Context, ld pdata.Logs) error {
	numRecords := ld.LogRecordCount()
	resourceLogs := ld.ResourceLogs()
	for i := 0; i < numRecords; i++ {
		resourceLog := resourceLogs.At(i)
		resource := resourceLog.Resource()

		pbEntries := []*logpb.LogEntry{}

		instLogs := resourceLog.InstrumentationLibraryLogs()
		numInstLogs := instLogs.Len()
		for j := 0; j < numInstLogs; j++ {
			instLog := instLogs.At(j)
			logSlice := instLog.Logs()
			numLogs := logSlice.Len()
			for k := 0; k < numLogs; k++ {
				logRecord := logSlice.At(k)

				pbEntry, err := e.createProtobufEntry(logRecord)
				if err != nil {
					// TODO capture and combine errors
					continue
				}
				pbEntries = append(pbEntries, pbEntry)
			}
		}

		req := logpb.WriteLogEntriesRequest{
			Entries:  pbEntries,
			LogName:  e.toLogNamePath("default"), // TODO placeholder
			Resource: e.toResource(resource),     // TODO placeholder
		}

		clientCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		client, err := e.getClient(clientCtx)
		if err != nil {
			return err
		}

		_, err = client.WriteLogEntries(ctx, &req)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *gcpLoggingExporter) createProtobufEntry(lr pdata.LogRecord) (newEntry *logpb.LogEntry, err error) {

	ts, err := ptypes.TimestampProto(time.Unix(0, int64(lr.Timestamp())))
	// ts, err := ptypes.TimestampProto(time.Now()) // TODO For testing old logs only
	if err != nil {
		return nil, err
	}

	newEntry = &logpb.LogEntry{
		Timestamp: ts,
		Severity:  toSeverity(lr.SeverityNumber()),
	}

	newEntry.Labels = toLabelsMap(lr.Attributes())

	body := lr.Body()
	switch body.Type() {
	case pdata.AttributeValueBOOL:
		newEntry.Payload = &logpb.LogEntry_TextPayload{TextPayload: fmt.Sprintf("%t", body.BoolVal())}
	case pdata.AttributeValueINT:
		newEntry.Payload = &logpb.LogEntry_TextPayload{TextPayload: fmt.Sprintf("%d", body.IntVal())}
	case pdata.AttributeValueDOUBLE:
		newEntry.Payload = &logpb.LogEntry_TextPayload{TextPayload: fmt.Sprintf("%f", body.DoubleVal())}
	case pdata.AttributeValueSTRING:
		newEntry.Payload = &logpb.LogEntry_TextPayload{TextPayload: body.StringVal()}
	case pdata.AttributeValueMAP:
		s := toProtoStruct(body.MapVal())
		newEntry.Payload = &logpb.LogEntry_JsonPayload{JsonPayload: s}
	// case pdata.AttributeValueARRAY: TODO when added
	default: // including pdata.AttributeValueNULL
		newEntry.Payload = &logpb.LogEntry_TextPayload{TextPayload: ""}
	}

	return newEntry, nil
}

func (e *gcpLoggingExporter) toLogNamePath(logName string) string {
	return fmt.Sprintf("projects/%s/logs/%s", e.projectID, url.PathEscape(logName))
}

func (e *gcpLoggingExporter) toResource(r pdata.Resource) *mrpb.MonitoredResource {
	mr := mrpb.MonitoredResource{
		Type: "global", // TODO how best to support other types
		Labels: map[string]string{
			"project_id": e.projectID,
		},
	}

	r.Attributes().ForEach(func(k string, v pdata.AttributeValue) {
		mr.Labels[k] = attValToString(v)
	})

	return &mr
}

func toProtoStruct(attMap pdata.AttributeMap) *structpb.Struct {
	fields := map[string]*structpb.Value{}
	attMap.ForEach(func(k string, v pdata.AttributeValue) {
		switch v.Type() {
		case pdata.AttributeValueBOOL:
			fields[k] = &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: v.BoolVal()}}
		case pdata.AttributeValueINT:
			fields[k] = &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v.IntVal())}}
		case pdata.AttributeValueDOUBLE:
			fields[k] = &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: v.DoubleVal()}}
		case pdata.AttributeValueSTRING:
			fields[k] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: v.StringVal()}}
		case pdata.AttributeValueMAP:
			fields[k] = &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: toProtoStruct(v.MapVal())}}
		// case pdata.AttributeValueARRAY: TODO when added
		default: // including pdata.AttributeValueNULL
			fields[k] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: ""}}
		}
	})
	return &structpb.Struct{Fields: fields}
}

func toLabelsMap(attMap pdata.AttributeMap) map[string]string {
	labelsMap := map[string]string{}
	attMap.ForEach(func(k string, v pdata.AttributeValue) {
		labelsMap[k] = attValToString(v)
	})
	return labelsMap
}

// TODO validate mapping
func toSeverity(s pdata.SeverityNumber) sev.LogSeverity {
	switch s {
	case pdata.SeverityNumberDEBUG:
		return sev.LogSeverity_DEBUG
	case pdata.SeverityNumberDEBUG2:
		return sev.LogSeverity_DEBUG
	case pdata.SeverityNumberDEBUG3:
		return sev.LogSeverity_DEBUG
	case pdata.SeverityNumberDEBUG4:
		return sev.LogSeverity_DEBUG
	case pdata.SeverityNumberINFO:
		return sev.LogSeverity_INFO
	case pdata.SeverityNumberINFO2:
		return sev.LogSeverity_INFO
	case pdata.SeverityNumberINFO3:
		return sev.LogSeverity_INFO
	case pdata.SeverityNumberINFO4:
		return sev.LogSeverity_INFO
	case pdata.SeverityNumberWARN:
		return sev.LogSeverity_WARNING
	case pdata.SeverityNumberWARN2:
		return sev.LogSeverity_WARNING
	case pdata.SeverityNumberWARN3:
		return sev.LogSeverity_WARNING
	case pdata.SeverityNumberWARN4:
		return sev.LogSeverity_WARNING
	case pdata.SeverityNumberERROR:
		return sev.LogSeverity_ERROR
	case pdata.SeverityNumberERROR2:
		return sev.LogSeverity_ERROR
	case pdata.SeverityNumberERROR3:
		return sev.LogSeverity_ERROR
	case pdata.SeverityNumberERROR4:
		return sev.LogSeverity_ERROR
	case pdata.SeverityNumberFATAL:
		return sev.LogSeverity_EMERGENCY
	case pdata.SeverityNumberFATAL2:
		return sev.LogSeverity_EMERGENCY
	case pdata.SeverityNumberFATAL3:
		return sev.LogSeverity_EMERGENCY
	case pdata.SeverityNumberFATAL4:
		return sev.LogSeverity_EMERGENCY
	default:
		return sev.LogSeverity_DEFAULT
	}
}

func attValToString(v pdata.AttributeValue) string {
	switch v.Type() {
	case pdata.AttributeValueBOOL:
		return fmt.Sprintf("%t", v.BoolVal())
	case pdata.AttributeValueINT:
		return fmt.Sprintf("%d", v.IntVal())
	case pdata.AttributeValueDOUBLE:
		return fmt.Sprintf("%f", v.DoubleVal())
	case pdata.AttributeValueSTRING:
		return v.StringVal()
	case pdata.AttributeValueMAP:
		attMap := v.MapVal()
		encoded, err := json.Marshal(attMap)
		if err != nil {
			return err.Error()
		}
		return string(encoded)
	// case pdata.AttributeValueARRAY: TODO when added
	default: // including pdata.AttributeValueNULL
		// don't include
		return "null"
	}
}
