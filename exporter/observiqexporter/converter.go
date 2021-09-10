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
	"encoding/hex"
	"encoding/json"
	"hash/fnv"

	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/model/pdata"
)

type observIQLogBatch struct {
	Logs []*observIQLog `json:"logs"`
}

type observIQLog struct {
	ID    string         `json:"id"`
	Size  int            `json:"size"`
	Entry preEncodedJSON `json:"entry"`
}

type observIQAgentInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type observIQLogInfo struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type observIQLogEntrySource struct {
	ID         string `json:"id"`
	SourceType string `json:"type,omitempty"`
	Name       string `json:"name,omitempty"`
	Version    string `json:"version,omitempty"`
}

type observIQLogEntry struct {
	Timestamp  string                  `json:"@timestamp"`
	Severity   string                  `json:"severity,omitempty"`
	EntryType  string                  `json:"type,omitempty"`
	LogInfo    *observIQLogInfo        `json:"log,omitempty"`
	SourceInfo *observIQLogEntrySource `json:"source,omitempty"`
	Message    string                  `json:"message,omitempty"`
	Resource   interface{}             `json:"resource,omitempty"`
	Agent      *observIQAgentInfo      `json:"agent,omitempty"`
	Data       map[string]interface{}  `json:"data,omitempty"`
	Labels     map[string]string       `json:"labels,omitempty"`
}

// Hash related variables, re-used to avoid multiple allocations
var fnvHash = fnv.New128a()
var fnvHashOut = make([]byte, 0, 16)

// Convert pdata.Logs to observIQLogBatch
func logdataToObservIQFormat(ld pdata.Logs, agentID string, agentName string, buildVersion string) (*observIQLogBatch, []error) {
	var rls = ld.ResourceLogs()
	var sliceOut = make([]*observIQLog, 0, ld.LogRecordCount())
	var errorsOut = make([]error, 0)

	for i := 0; i < rls.Len(); i++ {
		rl := rls.At(i)
		res := rl.Resource()
		resMap := attributeMapToBaseType(res.Attributes())
		ills := rl.InstrumentationLibraryLogs()
		for j := 0; j < ills.Len(); j++ {
			ill := ills.At(j)
			logs := ill.Logs()
			for k := 0; k < logs.Len(); k++ {
				oiqLogEntry := resourceAndInstrumentationLogToEntry(resMap, logs.At(k), agentID, agentName, buildVersion)

				jsonOIQLogEntry, err := json.Marshal(oiqLogEntry)

				if err != nil {
					//Skip this log, keep record of error
					errorsOut = append(errorsOut, consumererror.Permanent(err))
					continue
				}

				//fnv sum of the message is ID
				fnvHash.Reset()
				_, err = fnvHash.Write(jsonOIQLogEntry)
				if err != nil {
					errorsOut = append(errorsOut, consumererror.Permanent(err))
					continue
				}

				fnvHashOut = fnvHashOut[:0]
				fnvHashOut = fnvHash.Sum(fnvHashOut)

				fnvHashAsHex := hex.EncodeToString(fnvHashOut)

				sliceOut = append(sliceOut, &observIQLog{
					ID:    fnvHashAsHex,
					Size:  len(jsonOIQLogEntry),
					Entry: preEncodedJSON(jsonOIQLogEntry),
				})
			}
		}
	}

	return &observIQLogBatch{Logs: sliceOut}, errorsOut
}

// Output timestamp format, an ISO8601 compliant timestamp with millisecond precision
const timestampFieldOutputLayout = "2006-01-02T15:04:05.000Z07:00"

func resourceAndInstrumentationLogToEntry(resMap interface{}, log pdata.LogRecord, agentID string, agentName string, buildVersion string) *observIQLogEntry {
	ent := &observIQLogEntry{
		Timestamp: timestampFromRecord(log),
		Severity:  severityFromRecord(log),
		Labels:    flattenStringMap(attributeMapToBaseType(log.Attributes())),
		Resource:  resMap,
		Message:   messageFromRecord(log),
		Data:      dataFromLogRecord(log),
		Agent:     &observIQAgentInfo{Name: agentName, ID: agentID, Version: buildVersion},
	}
	doExtraTransforms(ent)
	return ent
}

func timestampFromRecord(log pdata.LogRecord) string {
	if log.Timestamp() == 0 {
		return timeNow().UTC().Format(timestampFieldOutputLayout)
	}
	return log.Timestamp().AsTime().UTC().Format(timestampFieldOutputLayout)
}

func messageFromRecord(log pdata.LogRecord) string {
	if log.Body().Type() == pdata.AttributeValueTypeString {
		return log.Body().StringVal()
	}

	return ""
}

func dataFromLogRecord(log pdata.LogRecord) map[string]interface{} {
	if log.Body().Type() == pdata.AttributeValueTypeMap {
		bodyMap := log.Body().MapVal()
		return attributeMapToBaseType(bodyMap)
	}
	return nil
}

//Mappings from opentelemetry severity number to observIQ severity string
var severityNumberToObservIQName = map[int32]string{
	0:  "default",
	1:  "trace",
	2:  "trace",
	3:  "trace",
	4:  "trace",
	5:  "debug",
	6:  "debug",
	7:  "debug",
	8:  "debug",
	9:  "info",
	10: "notice",
	11: "notice",
	12: "notice",
	13: "warning",
	14: "warning",
	15: "warning",
	16: "warning",
	17: "error",
	18: "critical",
	19: "alert",
	20: "alert",
	21: "emergency",
	22: "emergency",
	23: "catastrophe",
	24: "catastrophe",
}

/*
	Get severity from the a log record.
	We prefer the severity number, and map it to a string
	representing the opentelemetry defined severity.
	If there is no severity number, we use "default"
*/
func severityFromRecord(log pdata.LogRecord) string {
	var sevAsInt32 = int32(log.SeverityNumber())
	if sevAsInt32 < int32(len(severityNumberToObservIQName)) && sevAsInt32 >= 0 {
		return severityNumberToObservIQName[sevAsInt32]
	}
	return "default"
}
