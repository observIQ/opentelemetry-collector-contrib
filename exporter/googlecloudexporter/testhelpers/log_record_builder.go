// Copyright  The OpenTelemetry Authors
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

package testhelpers

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

// LogRecordBuilder helps to build quick plog.LogRecords for testing
type LogRecordBuilder struct {
	logRecord *plog.LogRecord
}

// NewLogRecordBuilder creates a new LogRecordBuilder with a plog.LogRecord initialized with some attributes
func NewLogRecordBuilder() *LogRecordBuilder {
	logRecord := plog.NewLogRecord()
	setupOtelLogRecordAttributes(&logRecord)

	return &LogRecordBuilder{
		logRecord: &logRecord,
	}
}

// NewLogRecordBuilder creates a new LogRecordBuilder and initializes given plog.LogRecord with some attributes
func NewLogRecordBuilderWithLogRecord(logRecord plog.LogRecord) *LogRecordBuilder {
	setupOtelLogRecordAttributes(&logRecord)

	return &LogRecordBuilder{
		logRecord: &logRecord,
	}
}

// WithTimeStamp adds a timestamp to the log record and returns the builder
func (l *LogRecordBuilder) WithTimeStamp() *LogRecordBuilder {
	l.logRecord.SetTimestamp(pcommon.NewTimestampFromTime(time.UnixMilli(0)))
	return l
}

// WithTraceId adds a trace id to the log record and returns the builder
func (l *LogRecordBuilder) WithTraceID(bytes [16]byte) *LogRecordBuilder {
	l.logRecord.SetTraceID(pcommon.NewTraceID(bytes))
	return l
}

// WithSpanId adds a span id to the log record and returns the builder
func (l *LogRecordBuilder) WithSpanID(bytes [8]byte) *LogRecordBuilder {
	l.logRecord.SetSpanID(pcommon.NewSpanID(bytes))
	return l
}

// WithEmptyAttributes adds empty attrigutes to the log record and returns the builder
func (l *LogRecordBuilder) WithEmptyAttributes() *LogRecordBuilder {
	l.logRecord.Attributes().Clear()
	return l
}

// WithBodyString adds a body with the given string to the log record and returns the builder
func (l *LogRecordBuilder) WithBodyString(body string) *LogRecordBuilder {
	l.logRecord.Body().SetStringVal(body)
	return l
}

// WithBodyBytes adds a body with bytes using the given string to the log record and returns the builder
func (l *LogRecordBuilder) WithBodyBytes(body string) *LogRecordBuilder {
	l.logRecord.Body().SetMBytesVal([]byte(body))
	return l
}

// WithBodyMap adds a body with a map to the log record and returns the builder
func (l *LogRecordBuilder) WithBodyMap() *LogRecordBuilder {
	mapVal := pcommon.NewValueMap()
	mapVal.MapVal().InsertString("key1", "body1")
	mapVal.MapVal().InsertInt("key2", 10)
	mapVal.MapVal().InsertBool("key3", true)
	mapVal.CopyTo(l.logRecord.Body())

	return l
}

// Build builds the pdata.LogRecord and returns it
func (l *LogRecordBuilder) Build() *plog.LogRecord {
	return l.logRecord
}

// setupOtelLogRecordAttributes creates some initial Attributes for a log record
func setupOtelLogRecordAttributes(logRecord *plog.LogRecord) {
	// key1-1: val1
	// key1-2:
	//   key2-1: val2
	//   key2-2:
	//     key3-1: val3
	key2_2Val := pcommon.NewValueMap()
	key2_2Val.MapVal().InsertString("key3-1", "val3")
	key1_2Val := pcommon.NewValueMap()
	key1_2Val.MapVal().InsertString("key2-1", "val2")
	key1_2Val.MapVal().Insert("key2-2", key2_2Val)
	logRecord.Attributes().InsertString("key1-1", "val1")
	logRecord.Attributes().Insert("key1-2", key1_2Val)
}
