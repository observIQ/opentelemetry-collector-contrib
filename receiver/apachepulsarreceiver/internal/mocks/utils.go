// Copyright The OpenTelemetry Authors
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

package mocks

import (
	utils "github.com/streamnative/pulsarctl/pkg/pulsar/utils"
	mock "github.com/stretchr/testify/mock"
)

type MockUtils struct {
	mock.Mock
}

func (_m *MockUtils) GetTopicName(name string) (*utils.TopicName, error) {
	ret := _m.Called()

	var r0 utils.TopicName
	if rf, ok := ret.Get(0).(func() utils.TopicName); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(utils.TopicName)
		}
	}

	return &r0, nil
}
