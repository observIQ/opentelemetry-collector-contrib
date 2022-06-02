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
	mock "github.com/stretchr/testify/mock"
	"go.opentelemetry.io/collector/pdata/plog"
	logging "google.golang.org/genproto/googleapis/logging/v2"
)

// EntryBuilder is an autogenerated mock type for the EntryBuilder type
type MockEntryBuilder struct {
	mock.Mock
}

// Build provides a mock function with given fields: oLogRecord
func (_m *MockEntryBuilder) Build(oLogRecord *plog.LogRecord) (*logging.LogEntry, error) {
	ret := _m.Called(oLogRecord)

	var r0 *logging.LogEntry
	if rf, ok := ret.Get(0).(func(*plog.LogRecord) *logging.LogEntry); ok {
		r0 = rf(oLogRecord)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*logging.LogEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*plog.LogRecord) error); ok {
		r1 = rf(oLogRecord)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
