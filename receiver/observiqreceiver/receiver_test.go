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
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/observiq/carbon/entry"
	"github.com/observiq/carbon/operator"
	"github.com/observiq/carbon/operator/builtin/input/file"
	"github.com/observiq/carbon/testutil"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap/zaptest"
)

func TestStart(t *testing.T) {
	factory := &Factory{}
	params := component.ReceiverCreateParams{
		Logger: zaptest.NewLogger(t),
	}
	mockConsumer := mockLogsConsumer{}
	receiver, _ := factory.CreateLogsReceiver(context.Background(), params, factory.CreateDefaultConfig(), &mockConsumer)

	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err, "receiver start failed")

	obsReceiver := receiver.(*observiqReceiver)
	obsReceiver.logsChan <- entry.New()
	receiver.Shutdown(context.Background())
	require.Equal(t, 1, mockConsumer.received, "one log entry expected")
}

func TestHandleStartError(t *testing.T) {
	factory := &Factory{}
	params := component.ReceiverCreateParams{
		Logger: zaptest.NewLogger(t),
	}
	mockConsumer := mockLogsConsumer{}

	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Pipeline = append(cfg.Pipeline, newUnstartableParams())

	receiver, err := factory.CreateLogsReceiver(context.Background(), params, cfg, &mockConsumer)
	require.NoError(t, err, "receiver should successfully build")

	err = receiver.Start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err, "receiver fails to start under rare circumstances")
}

func TestHandleConsumeError(t *testing.T) {
	factory := &Factory{}
	params := component.ReceiverCreateParams{
		Logger: zaptest.NewLogger(t),
	}
	mockConsumer := mockLogsRejecter{}
	receiver, _ := factory.CreateLogsReceiver(context.Background(), params, factory.CreateDefaultConfig(), &mockConsumer)

	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err, "receiver start failed")

	obsReceiver := receiver.(*observiqReceiver)
	obsReceiver.logsChan <- entry.New()
	receiver.Shutdown(context.Background())
	require.Equal(t, 1, mockConsumer.rejected, "one log entry expected")
}

func BenchmarkEndToEnd(b *testing.B) {

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		b.Errorf(err.Error())
		b.FailNow()
	}

	filePath := filepath.Join(tempDir, "bench.log")

	logsChan := make(chan *entry.Entry)
	defer close(logsChan)
	buildContext := testutil.NewBuildContext(b)
	buildContext.Parameters = map[string]interface{}{"logs_channel": logsChan}

	chanOutputCfg := NewReceiverOutputConfig("test-output")
	chanOperator, err := chanOutputCfg.Build(buildContext)
	require.NoError(b, err)

	fileInputCfg := file.NewInputConfig("test-input")
	fileInputCfg.OutputIDs = []string{"test-output"}
	fileInputCfg.Include = []string{filePath}
	fileInputCfg.StartAt = "beginning"

	fileOperator, err := fileInputCfg.Build(buildContext)
	require.NoError(b, err)

	err = fileOperator.SetOutputs([]operator.Operator{chanOperator})
	require.NoError(b, err)

	// Populate the file that will be consumed
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	require.NoError(b, err)
	for i := 0; i < b.N; i++ {
		file.WriteString("testlog\n")
	}

	// Run the actual benchmark
	b.ResetTimer()
	require.NoError(b, fileOperator.Start())
	for i := 0; i < b.N; i++ {
		<-logsChan
	}
}
