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
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver/internal/metadata"
)

func TestScraper(t *testing.T) {
	mysqlMock := fakeClient{}
	sc := newMySQLScraper(zap.NewNop(), &Config{
		User:     "otel",
		Password: "otel",
		Endpoint: "localhost:3306",
	})
	sc.client = &mysqlMock

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
			bufferPoolPagesMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.BufferPoolPagesState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				bufferPoolPagesMetrics[label] = dp.Value()
			}
			require.Equal(t, 6, len(bufferPoolPagesMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_pages state:data":    981,
				"mysql.buffer_pool_pages state:dirty":   0,
				"mysql.buffer_pool_pages state:flushed": 168,
				"mysql.buffer_pool_pages state:free":    7207,
				"mysql.buffer_pool_pages state:misc":    4,
				"mysql.buffer_pool_pages state:total":   8192,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperationsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				bufferPoolOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.buffer_pool_operations state:read_ahead":         0,
				"mysql.buffer_pool_operations state:read_ahead_evicted": 0,
				"mysql.buffer_pool_operations state:read_ahead_rnd":     0,
				"mysql.buffer_pool_operations state:read_requests":      14837,
				"mysql.buffer_pool_operations state:reads":              838,
				"mysql.buffer_pool_operations state:wait_free":          0,
				"mysql.buffer_pool_operations state:write_requests":     1669,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.DoubleGauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSizeState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				bufferPoolSizeMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_size state:data":  16072704,
				"mysql.buffer_pool_size state:dirty": 0,
				"mysql.buffer_pool_size state:size":  134217728,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.CommandState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				commandsMetrics[label] = dp.Value()
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.commands state:close":          0,
				"mysql.commands state:execute":        0,
				"mysql.commands state:fetch":          0,
				"mysql.commands state:prepare":        0,
				"mysql.commands state:reset":          0,
				"mysql.commands state:send_long_data": 0,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.HandlerState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				handlersMetrics[label] = dp.Value()
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]int64{
				"mysql.handlers state:commit":             564,
				"mysql.handlers state:delete":             0,
				"mysql.handlers state:discover":           0,
				"mysql.handlers state:lock":               7127,
				"mysql.handlers state:mrr_init":           0,
				"mysql.handlers state:prepare":            0,
				"mysql.handlers state:read_first":         52,
				"mysql.handlers state:read_key":           1680,
				"mysql.handlers state:read_last":          0,
				"mysql.handlers state:read_next":          3960,
				"mysql.handlers state:read_prev":          0,
				"mysql.handlers state:read_rnd":           0,
				"mysql.handlers state:read_rnd_next":      505063,
				"mysql.handlers state:rollback":           0,
				"mysql.handlers state:savepoint":          0,
				"mysql.handlers state:savepoint_rollback": 0,
				"mysql.handlers state:update":             315,
				"mysql.handlers state:write":              256734,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.DoubleWritesState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				doubleWritesMetrics[label] = dp.Value()
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]int64{
				"mysql.double_writes state:writes":  8,
				"mysql.double_writes state:written": 27,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.LogOperationsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				logOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.log_operations state:requests": 646,
				"mysql.log_operations state:waits":    0,
				"mysql.log_operations state:writes":   14,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.OperationsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				operationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.operations state:fsyncs": 46,
				"mysql.operations state:reads":  860,
				"mysql.operations state:writes": 215,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.PageOperationsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				pageOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.page_operations state:created": 144,
				"mysql.page_operations state:read":    837,
				"mysql.page_operations state:written": 168,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.RowLocksState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				rowLocksMetrics[label] = dp.Value()
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_locks state:time":  0,
				"mysql.row_locks state:waits": 0,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.RowOperationsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				rowOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_operations state:deleted":  0,
				"mysql.row_operations state:inserted": 0,
				"mysql.row_operations state:read":     0,
				"mysql.row_operations state:updated":  0,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.LocksState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				locksMetrics[label] = dp.Value()
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.locks state:immediate": 521,
				"mysql.locks state:waited":    0,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.SortsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				sortsMetrics[label] = dp.Value()
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.sorts state:merge_passes": 0,
				"mysql.sorts state:range":        0,
				"mysql.sorts state:rows":         0,
				"mysql.sorts state:scan":         0,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.DoubleGauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.ThreadsState)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				threadsMetrics[label] = dp.Value()
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]float64{
				"mysql.threads state:cached":    0,
				"mysql.threads state:connected": 1,
				"mysql.threads state:created":   1,
				"mysql.threads state:running":   2,
			}, threadsMetrics)
		}
	}
}

func TestScrapeErrorBadConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		user     string
		password string
		endpoint string
	}{
		{
			desc:     "no user",
			password: "otel",
			endpoint: "localhost:3306",
		},
		{
			desc:     "no password",
			user:     "otel",
			endpoint: "localhost:3306",
		},
		{
			desc:     "no endpoint",
			user:     "otel",
			password: "otel",
			endpoint: "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mysqlMock := fakeClient{}
			sc := newMySQLScraper(zap.NewNop(), &Config{
				User:     tC.user,
				Password: tC.password,
				Endpoint: tC.endpoint,
			})
			sc.client = &mysqlMock
			err := sc.start(context.Background(), componenttest.NewNopHost())
			require.NotNil(t, err)
		})
	}

	t.Run("good config", func(t *testing.T) {
		mysqlMock := fakeClient{}
		sc := newMySQLScraper(zap.NewNop(), &Config{
			User:     "otel",
			Password: "otel",
			Endpoint: "localhost:3306",
		})
		sc.client = &mysqlMock
		err := sc.start(context.Background(), componenttest.NewNopHost())
		require.Nil(t, err)
	})
}
