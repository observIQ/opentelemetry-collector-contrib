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

package observiqreceiver

import (
	"testing"
	"time"

	obsentry "github.com/observiq/carbon/entry"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/pdata"
)

func TestConvertMetadata(t *testing.T) {

	now := time.Now()

	obsEntry := obsentry.New()
	obsEntry.Timestamp = now
	obsEntry.Severity = obsentry.Error
	obsEntry.AddLabel("one", "two")
	obsEntry.Record = true

	result := convert(obsEntry)

	resourceLogs := result.ResourceLogs()
	require.Equal(t, 1, resourceLogs.Len(), "expected 1 resource")

	libLogs := resourceLogs.At(0).InstrumentationLibraryLogs()
	require.Equal(t, 1, libLogs.Len(), "expected 1 library")

	logSlice := libLogs.At(0).Logs()
	require.Equal(t, 1, logSlice.Len(), "expected 1 log")

	log := logSlice.At(0)
	require.Equal(t, now.UnixNano(), int64(log.Timestamp()))

	require.Equal(t, pdata.SeverityNumberERROR, log.SeverityNumber())
	require.Equal(t, "Error", log.SeverityText())

	atts := log.Attributes()
	require.Equal(t, 1, atts.Len(), "expected 1 attribute")
	attVal, ok := atts.Get("one")
	require.True(t, ok, "expected label with key 'one'")
	require.Equal(t, "two", attVal.StringVal(), "expected label to have value 'two'")

	bod := log.Body()
	require.Equal(t, pdata.AttributeValueBOOL, int(bod.Type()))
	require.True(t, bod.BoolVal())
}

func TestConvertBody(t *testing.T) {

	recordToBody := func(record interface{}) pdata.AttributeValue {
		obsEntry := obsentry.New()
		obsEntry.Record = record
		return convert(obsEntry).ResourceLogs().At(0).InstrumentationLibraryLogs().At(0).Logs().At(0).Body()
	}

	require.True(t, recordToBody(true).BoolVal())
	require.False(t, recordToBody(false).BoolVal())

	require.Equal(t, "hello", recordToBody("hello").StringVal())
	require.Equal(t, "bytes", recordToBody([]byte("bytes")).StringVal())

	require.Equal(t, int64(1), recordToBody(1).IntVal())
	require.Equal(t, int64(1), recordToBody(int8(1)).IntVal())
	require.Equal(t, int64(1), recordToBody(int16(1)).IntVal())
	require.Equal(t, int64(1), recordToBody(int32(1)).IntVal())
	require.Equal(t, int64(1), recordToBody(int64(1)).IntVal())

	require.Equal(t, int64(1), recordToBody(uint(1)).IntVal())
	require.Equal(t, int64(1), recordToBody(uint8(1)).IntVal())
	require.Equal(t, int64(1), recordToBody(uint16(1)).IntVal())
	require.Equal(t, int64(1), recordToBody(uint32(1)).IntVal())
	require.Equal(t, int64(1), recordToBody(uint64(1)).IntVal())

	require.Equal(t, float64(1), recordToBody(float32(1)).DoubleVal())
	require.Equal(t, float64(1), recordToBody(float64(1)).DoubleVal())

	// TODO structured bodies
}
