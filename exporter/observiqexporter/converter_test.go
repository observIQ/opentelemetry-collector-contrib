// Copyright  OpenTelemetry Authors
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

package observiqexporter

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/translator/conventions"
)

func resourceAndLogRecordsToLogs(r pdata.Resource, lrs []pdata.LogRecord) pdata.Logs {
	logs := pdata.NewLogs()
	resLogs := logs.ResourceLogs()
	resLogs.AppendEmpty()

	resLog := resLogs.At(0)
	resLogRes := resLog.Resource()

	r.Attributes().Range(func(k string, v pdata.AttributeValue) bool {
		resLogRes.Attributes().Insert(k, v)
		return true
	})

	resLog.InstrumentationLibraryLogs().Resize(len(lrs))
	for i, l := range lrs {
		ills := pdata.NewInstrumentationLibraryLogs()
		ills.Logs().Resize(1)
		l.CopyTo(ills.Logs().At(0))

		ills.CopyTo(resLog.InstrumentationLibraryLogs().At(i))
	}

	return logs
}

func TestLogdataToObservIQFormat(t *testing.T) {
	ts := time.Date(2021, 12, 11, 10, 9, 8, 1, time.UTC)
	stringTs := "2021-12-11T10:09:08.000Z"
	nanoTs := pdata.Timestamp(ts.UnixNano())
	timeNow = func() time.Time {
		return ts
	}

	testCases := []struct {
		name          string
		logRecordFn   func() pdata.LogRecord
		logResourceFn func() pdata.Resource
		agentName     string
		agentID       string
		output        observIQLogEntry
		expectErr     bool
	}{
		{
			"Happy path with string attributes",
			func() pdata.LogRecord {
				logRecord := pdata.NewLogRecord()
				logRecord.Body().SetStringVal("Message")
				logRecord.Attributes().InsertString(conventions.AttributeServiceName, "myapp")
				logRecord.Attributes().InsertString(conventions.AttributeHostName, "myhost")
				logRecord.Attributes().InsertString("custom", "custom")
				logRecord.SetTimestamp(nanoTs)
				return logRecord
			},
			pdata.NewResource,
			"agent",
			"agentID",
			observIQLogEntry{
				Timestamp: stringTs,
				Message:   "Message",
				Severity:  "default",
				Resource:  map[string]interface{}{},
				Data: map[string]interface{}{
					conventions.AttributeServiceName: "myapp",
					conventions.AttributeHostName:    "myhost",
					"custom":                         "custom",
				},
				Agent: &observIQAgentInfo{Name: "agent", ID: "agentID"},
			},
			false,
		},
		{
			"works with attributes of all types",
			func() pdata.LogRecord {
				logRecord := pdata.NewLogRecord()
				logRecord.Body().SetStringVal("Message")
				logRecord.Attributes().InsertString(conventions.AttributeServiceName, "myapp")
				logRecord.Attributes().InsertBool("bool", true)
				logRecord.Attributes().InsertDouble("double", 1.0)
				logRecord.Attributes().InsertInt("int", 3)
				logRecord.Attributes().InsertNull("null")

				mapVal := pdata.NewAttributeValueMap()
				mapVal.MapVal().Insert("mapKey", pdata.NewAttributeValueString("value"))
				logRecord.Attributes().Insert("map", mapVal)

				arrVal := pdata.NewAttributeValueArray()
				arrVal.ArrayVal().Resize(2)
				arrVal.ArrayVal().At(0).SetIntVal(1)
				arrVal.ArrayVal().At(1).SetIntVal(2)
				logRecord.Attributes().Insert("array", arrVal)

				logRecord.SetTimestamp(nanoTs)
				return logRecord
			},
			pdata.NewResource,
			"agent",
			"agentID",
			observIQLogEntry{
				Timestamp: stringTs,
				Message:   "Message",
				Severity:  "default",
				Resource:  map[string]interface{}{},
				Data: map[string]interface{}{
					conventions.AttributeServiceName: "myapp",
					"bool":                           true,
					"double":                         float64(1.0),
					"int":                            float64(3),
					"map": map[string]interface{}{
						"mapKey": "value",
					},
					"array": []interface{}{float64(1), float64(2)},
				},
				Agent: &observIQAgentInfo{Name: "agent", ID: "agentID"},
			},
			false,
		},
		{
			"Body is nil",
			func() pdata.LogRecord {
				logRecord := pdata.NewLogRecord()
				logRecord.SetTimestamp(nanoTs)
				return logRecord
			},
			pdata.NewResource,
			"agent",
			"agentID",
			observIQLogEntry{
				Timestamp: stringTs,
				Message:   "",
				Severity:  "default",
				Resource:  map[string]interface{}{},
				Data:      nil,
				Agent:     &observIQAgentInfo{Name: "agent", ID: "agentID"},
			},
			false,
		},
		{
			"Body is map",
			func() pdata.LogRecord {
				logRecord := pdata.NewLogRecord()

				mapVal := pdata.NewAttributeValueMap()
				mapVal.MapVal().Insert("mapKey", pdata.NewAttributeValueString("value"))
				mapVal.CopyTo(logRecord.Body())

				logRecord.SetTimestamp(nanoTs)
				return logRecord
			},
			pdata.NewResource,
			"agent",
			"agentID",
			observIQLogEntry{
				Timestamp: stringTs,
				Severity:  "default",
				Resource:  map[string]interface{}{},
				Data:      nil,
				Agent:     &observIQAgentInfo{Name: "agent", ID: "agentID"},
				Body: map[string]interface{}{
					"mapKey": "value",
				},
			},
			false,
		},
		{
			"No timestamp on record",
			func() pdata.LogRecord {
				logRecord := pdata.NewLogRecord()
				logRecord.Body().SetStringVal("Message")
				return logRecord
			},
			pdata.NewResource,
			"agent",
			"agentID",
			observIQLogEntry{
				Timestamp: stringTs,
				Message:   "Message",
				Severity:  "default",
				Resource:  map[string]interface{}{},
				Data:      nil,
				Agent:     &observIQAgentInfo{Name: "agent", ID: "agentID"},
			},
			false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			logs := resourceAndLogRecordsToLogs(testCase.logResourceFn(), []pdata.LogRecord{testCase.logRecordFn()})
			res, errs := logdataToObservIQFormat(logs, testCase.agentID, testCase.agentName)

			if testCase.expectErr {
				require.NotEmpty(t, errs)
			} else {
				require.Empty(t, errs)
			}

			require.Equal(t, 1, len(res.Logs))

			oiqLogEntry := observIQLogEntry{}
			err := json.Unmarshal(res.Logs[0].Entry, &oiqLogEntry)

			if testCase.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.output, oiqLogEntry)
		})
	}
}
