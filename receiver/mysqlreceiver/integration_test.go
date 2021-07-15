// Copyright 2021, observIQ
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

package mysqlreceiver

import (
	"context"
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver/internal/metadata"
)

func mysqlContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.mysql",
		},
		ExposedPorts: []string{"3306:3306"},
		WaitingFor:   wait.ForListeningPort("3306"),
	}
	require.NoError(t, req.Validate())

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	code, err := container.Exec(context.Background(), []string{"/setup.sh"})
	require.NoError(t, err)
	require.Equal(t, 0, code)

	return container
}

type MysqlIntegrationSuite struct {
	suite.Suite
}

func TestMysqlIntegration(t *testing.T) {
	suite.Run(t, new(MysqlIntegrationSuite))
}

func (suite *MysqlIntegrationSuite) TestHappyPath() {
	t := suite.T()
	container := mysqlContainer(t)
	defer container.Terminate(context.Background())
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	sc := newMySQLScraper(zap.NewNop(), &Config{
		User:     "otel",
		Password: "otel",
		Endpoint: fmt.Sprintf("%s:3306", hostname),
	})
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)
	require.Equal(t, 1, rms.Len())

	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 14, ms.Len())
	require.Equal(t, 14, len(metadata.M.Names()))

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		switch m.Name() {
		case metadata.M.MysqlBufferPoolPages.Name():
			dps := m.DoubleGauge().DataPoints()
			require.Equal(t, 6, dps.Len())
			bufferPoolPagesMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.BufferPoolPagesState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				bufferPoolPagesMetrics[label] = true
			}
			require.Equal(t, 6, len(bufferPoolPagesMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_pages state:data":    true,
				"mysql.buffer_pool_pages state:dirty":   true,
				"mysql.buffer_pool_pages state:flushed": true,
				"mysql.buffer_pool_pages state:free":    true,
				"mysql.buffer_pool_pages state:misc":    true,
				"mysql.buffer_pool_pages state:total":   true,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperationsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				bufferPoolOperationsMetrics[label] = true
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_operations state:read_ahead":         true,
				"mysql.buffer_pool_operations state:read_ahead_evicted": true,
				"mysql.buffer_pool_operations state:read_ahead_rnd":     true,
				"mysql.buffer_pool_operations state:read_requests":      true,
				"mysql.buffer_pool_operations state:reads":              true,
				"mysql.buffer_pool_operations state:wait_free":          true,
				"mysql.buffer_pool_operations state:write_requests":     true,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.DoubleGauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSizeState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				bufferPoolSizeMetrics[label] = true
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_size state:data":  true,
				"mysql.buffer_pool_size state:dirty": true,
				"mysql.buffer_pool_size state:size":  true,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.CommandState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				commandsMetrics[label] = true
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.commands state:close":          true,
				"mysql.commands state:execute":        true,
				"mysql.commands state:fetch":          true,
				"mysql.commands state:prepare":        true,
				"mysql.commands state:reset":          true,
				"mysql.commands state:send_long_data": true,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.HandlerState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				handlersMetrics[label] = true
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]bool{
				"mysql.handlers state:commit":             true,
				"mysql.handlers state:delete":             true,
				"mysql.handlers state:discover":           true,
				"mysql.handlers state:lock":               true,
				"mysql.handlers state:mrr_init":           true,
				"mysql.handlers state:prepare":            true,
				"mysql.handlers state:read_first":         true,
				"mysql.handlers state:read_key":           true,
				"mysql.handlers state:read_last":          true,
				"mysql.handlers state:read_next":          true,
				"mysql.handlers state:read_prev":          true,
				"mysql.handlers state:read_rnd":           true,
				"mysql.handlers state:read_rnd_next":      true,
				"mysql.handlers state:rollback":           true,
				"mysql.handlers state:savepoint":          true,
				"mysql.handlers state:savepoint_rollback": true,
				"mysql.handlers state:update":             true,
				"mysql.handlers state:write":              true,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.DoubleWritesState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				doubleWritesMetrics[label] = true
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]bool{
				"mysql.double_writes state:writes":  true,
				"mysql.double_writes state:written": true,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.LogOperationsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				logOperationsMetrics[label] = true
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.log_operations state:requests": true,
				"mysql.log_operations state:waits":    true,
				"mysql.log_operations state:writes":   true,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.OperationsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				operationsMetrics[label] = true
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.operations state:fsyncs": true,
				"mysql.operations state:reads":  true,
				"mysql.operations state:writes": true,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.PageOperationsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				pageOperationsMetrics[label] = true
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.page_operations state:created": true,
				"mysql.page_operations state:read":    true,
				"mysql.page_operations state:written": true,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.RowLocksState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				rowLocksMetrics[label] = true
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]bool{
				"mysql.row_locks state:time":  true,
				"mysql.row_locks state:waits": true,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.RowOperationsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				rowOperationsMetrics[label] = true
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.row_operations state:deleted":  true,
				"mysql.row_operations state:inserted": true,
				"mysql.row_operations state:read":     true,
				"mysql.row_operations state:updated":  true,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.LocksState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				locksMetrics[label] = true
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]bool{
				"mysql.locks state:immediate": true,
				"mysql.locks state:waited":    true,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.SortsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				sortsMetrics[label] = true
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.sorts state:merge_passes": true,
				"mysql.sorts state:range":        true,
				"mysql.sorts state:rows":         true,
				"mysql.sorts state:scan":         true,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.DoubleGauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.ThreadsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				threadsMetrics[label] = true
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.threads state:cached":    true,
				"mysql.threads state:connected": true,
				"mysql.threads state:created":   true,
				"mysql.threads state:running":   true,
			}, threadsMetrics)
		}
	}
}

func (suite *MysqlIntegrationSuite) TestStartStop() {
	t := suite.T()
	container := mysqlContainer(t)
	defer container.Terminate(context.Background())
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	sc := newMySQLScraper(zap.NewNop(), &Config{
		User:     "otel",
		Password: "otel",
		Endpoint: fmt.Sprintf("%s:3306", hostname),
	})

	// require scraper to connection to be open
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	require.False(t, sc.client.Closed())

	// require scraper to connection to be closed
	err = sc.shutdown(context.Background())
	require.NoError(t, err)
	require.True(t, sc.client.Closed())

	// require scraper to produce no error when closed again
	err = sc.shutdown(context.Background())
	require.NoError(t, err)
	require.True(t, sc.client.Closed())
}
