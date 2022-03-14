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

package varnishreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/varnishreceiver"

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"
)

func TestNewVarnishClient(t *testing.T) {
	client := newVarnishClient(
		createDefaultConfig().(*Config),
		componenttest.NewNopHost(),
		componenttest.NewNopTelemetrySettings())
	require.NotNil(t, client)
}

func TestBuildCommand(t *testing.T) {
	testCases := []struct {
		desc    string
		config  Config
		command string
		argList []string
	}{
		{
			desc:    "no working or exec dir",
			config:  Config{},
			command: "varnishstat",
			argList: []string{"-j"},
		},
		{
			desc: "with working and exec dir",
			config: Config{
				WorkingDir: "/working/dir",
				ExecDir:    "/exec/dir/varnishstat",
			},
			command: "/exec/dir/varnishstat",
			argList: []string{"-j", "-n", "/working/dir"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			client := newVarnishClient(
				&tC.config,
				componenttest.NewNopHost(),
				componenttest.NewNopTelemetrySettings())
			command, argList := client.BuildCommand()
			require.EqualValues(t, tC.command, command)
			require.EqualValues(t, tC.argList, argList)
		})
	}
}

func getBytes(filename string) ([]byte, error) {
	if filename == "" {
		return nil, errors.New("bad response")
	}

	file, err := os.Open(path.Join("testdata", "scraper", filename))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func TestGetStats(t *testing.T) {

	mockExecuter := new(MockExecuter)
	// mockExecuter.On("Execute", "varnishstat", "-j").Return([]byte(""), nil)
	mockExecuter.On("Execute", "varnishstat", []string{"-j"}).Return(getBytes("mock_response6_0.json"))

	myclient := varnishClient{
		exec:   mockExecuter,
		cfg:    createDefaultConfig().(*Config),
		logger: zap.NewNop(),
	}

	stats, err := myclient.GetStats()
	require.NoError(t, err)
	require.NotNil(t, stats)

	mockExecuter6_5 := new(MockExecuter)
	// mockExecuter6_5.On("Execute", "varnishstat", "-j").Return([]byte(""), nil)
	mockExecuter6_5.On("Execute", "varnishstat", []string{"-j"}).Return(getBytes("mock_response6_5.json"))

	myclient6_5 := varnishClient{
		exec:   mockExecuter6_5,
		cfg:    createDefaultConfig().(*Config),
		logger: zap.NewNop(),
	}

	stats6_5, err := myclient6_5.GetStats()
	require.NoError(t, err)
	require.NotNil(t, stats)

	require.EqualValues(t, stats, stats6_5)

}

type MockExecuter struct {
	mock.Mock
}

// Execute provides a mock function with given fields: command, args
func (_m *MockExecuter) Execute(command string, args []string) ([]byte, error) {
	ret := _m.Called(command, args)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, []string) []byte); ok {
		r0 = rf(command, args)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, []string) error); ok {
		r1 = rf(command, args)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
