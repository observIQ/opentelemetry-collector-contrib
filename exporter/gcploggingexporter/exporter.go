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

	client *vkit.Client
}

// Ensure this exporter adheres to required interface
var _ component.LogsExporter = (*gcpLoggingExporter)(nil)

func (e *gcpLoggingExporter) Start(ctx context.Context, host component.Host) error {
	options := []option.ClientOption{option.WithCredentials(e.credentials)}
	client, err := vkit.NewClient(ctx, options...)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	e.client = client
	return nil
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

		// TODO convert and use in request
		// resource := resourceLog.Resource()

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
			Resource: e.globalResource(),         // TODO placeholder
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second) // TODO make configurable
		defer cancel()
		_, err := e.client.WriteLogEntries(ctx, &req)
		if err != nil {
			// TODO capture and combine errors
		}
	}

	return nil
}

func (e *gcpLoggingExporter) createProtobufEntry(lr pdata.LogRecord) (newEntry *logpb.LogEntry, err error) {

	ts, err := ptypes.TimestampProto(time.Unix(0, int64(lr.Timestamp())))
	if err != nil {
		return nil, err
	}

	newEntry = &logpb.LogEntry{
		Timestamp: ts,
		Severity:  toSeverity(lr.SeverityNumber()),
	}

	// TODO resource?

	newEntry.Labels = map[string]string{}
	attMap := lr.Attributes()
	attMap.ForEach(func(k string, v pdata.AttributeValue) {
		newEntry.Labels[k] = fmt.Sprintf("%v", v) // TODO do better
	})

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

func (e *gcpLoggingExporter) toLogNamePath(logName string) string {
	return fmt.Sprintf("projects/%s/logs/%s", e.projectID, url.PathEscape(logName))
}

func (e *gcpLoggingExporter) globalResource() *mrpb.MonitoredResource {
	return &mrpb.MonitoredResource{
		Type: "global",
		Labels: map[string]string{
			"project_id": e.projectID,
		},
	}
}

// TODO make this mapping less crude
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
