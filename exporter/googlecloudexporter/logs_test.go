// Copyright 2019 OpenTelemetry Authors
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

package googlecloudexporter

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/plog"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/genproto/googleapis/logging/v2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type testLogServer struct {
	reqCh chan *logging.WriteLogEntriesRequest
}

func (ts *testLogServer) DeleteLog(context.Context, *logging.DeleteLogRequest) (*emptypb.Empty, error) {
	return nil, nil
}

func (ts *testLogServer) WriteLogEntries(ctx context.Context, req *logging.WriteLogEntriesRequest) (*logging.WriteLogEntriesResponse, error) {
	go func() { ts.reqCh <- req }()
	return &logging.WriteLogEntriesResponse{}, nil
}

func (ts *testLogServer) ListLogEntries(ctx context.Context, req *logging.ListLogEntriesRequest) (*logging.ListLogEntriesResponse, error) {
	return &logging.ListLogEntriesResponse{}, nil
}

func (ts *testLogServer) ListMonitoredResourceDescriptors(ctx context.Context, req *logging.ListMonitoredResourceDescriptorsRequest) (*logging.ListMonitoredResourceDescriptorsResponse, error) {
	return &logging.ListMonitoredResourceDescriptorsResponse{}, nil
}

func (ts *testLogServer) ListLogs(ctx context.Context, req *logging.ListLogsRequest) (*logging.ListLogsResponse, error) {
	return &logging.ListLogsResponse{}, nil
}

func (ts *testLogServer) TailLogEntries(req logging.LoggingServiceV2_TailLogEntriesServer) error {
	return nil
}

func TestGoogleCloudLogExport(t *testing.T) {
	type testCase struct {
		name        string
		cfg         *Config
		expectedErr string
	}

	testCases := []testCase{
		{
			name: "Standard",
			cfg: &Config{
				CollectorConfig: CollectorConfig{
					ProjectID: "idk",
				},
				ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
				LogConfig: LogConfig{
					Endpoint:    "127.0.0.1:8080",
					UseInsecure: true,
				},
			},
		},
		{
			name: "Standard_WithoutSendingQueue",
			cfg: &Config{
				CollectorConfig: CollectorConfig{
					ProjectID: "idk",
				},
				ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
				LogConfig: LogConfig{
					Endpoint:    "127.0.0.1:8080",
					UseInsecure: true,
				},
				QueueSettings: exporterhelper.QueueSettings{
					Enabled: false,
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			srv := grpc.NewServer()
			reqCh := make(chan *logging.WriteLogEntriesRequest)
			logging.RegisterLoggingServiceV2Server(srv, &testLogServer{reqCh: reqCh})

			lis, err := net.Listen("tcp", "localhost:8080")
			require.NoError(t, err)
			defer lis.Close()

			go srv.Serve(lis)

			le, err := newGoogleCloudLogsExporter(test.cfg, componenttest.NewNopExporterCreateSettings())
			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
				return
			}
			require.NoError(t, err)
			defer func() { require.NoError(t, le.Shutdown(context.Background())) }()

			logs := plog.NewLogs()
			resourceLogsSlice := plog.NewResourceLogsSlice()
			resourceLogs := resourceLogsSlice.AppendEmpty()
			scopeLogsSlice := plog.NewScopeLogsSlice()
			scopeLogs := scopeLogsSlice.AppendEmpty()
			logRecordSlice := plog.NewLogRecordSlice()

			logRecordOne := logRecordSlice.AppendEmpty()
			populateLogRecord(logRecordOne, "request 1")
			logRecordTwo := logRecordSlice.AppendEmpty()
			populateLogRecord(logRecordTwo, "request 2")
			logRecordThree := logRecordSlice.AppendEmpty()
			populateLogRecord(logRecordThree, "request 3")
			logRecordFour := logRecordSlice.AppendEmpty()
			populateLogRecord(logRecordFour, "request 4")

			logRecordSlice.CopyTo(scopeLogs.LogRecords())
			scopeLogsSlice.CopyTo(resourceLogs.ScopeLogs())
			resourceLogsSlice.CopyTo(logs.ResourceLogs())

			err = le.ConsumeLogs(context.Background(), logs)
			assert.NoError(t, err)

			r := <-reqCh
			assert.Len(t, r.Entries, 4)
			assert.Equal(t, "", r.Entries[0].LogName)
			assert.Equal(t, "projects/idk/logs/default", r.LogName)
			assert.Equal(t, "global", r.Resource.Type)
			assert.Equal(t, "idk", r.Resource.GetLabels()["project_id"])
			assert.Equal(t, (*monitoredres.MonitoredResource)(nil), r.Entries[0].Resource)
			assert.Equal(t, "request 1", r.Entries[0].GetTextPayload())
		})
	}
}
