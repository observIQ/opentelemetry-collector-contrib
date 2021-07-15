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
	"errors"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver/internal/metadata"
)

type mySQLScraper struct {
	client client

	logger *zap.Logger
	config *Config
}

func newMySQLScraper(
	logger *zap.Logger,
	config *Config,
) *mySQLScraper {
	return &mySQLScraper{
		logger: logger,
		config: config,
	}
}

// start starts the scraper by initializing the db client connection.
func (m *mySQLScraper) start(_ context.Context, host component.Host) error {
	if m.config.User == "" || m.config.Password == "" || m.config.Endpoint == "" {
		return errors.New("missing database configuration parameters")
	}
	client, err := newMySQLClient(mySQLConfig{
		user:     m.config.User,
		pass:     m.config.Password,
		endpoint: m.config.Endpoint,
	})
	if err != nil {
		return err
	}
	m.client = client

	return nil
}

// shutdown closes open connections.
func (m *mySQLScraper) shutdown(context.Context) error {
	if !m.client.Closed() {
		m.logger.Info("gracefully shutdown")
		return m.client.Close()
	}
	return nil
}

// initMetric initializes a metric with a metadata label.
func initMetric(ms pdata.MetricSlice, mi metadata.MetricIntf) pdata.Metric {
	m := ms.AppendEmpty()
	mi.Init(m)
	return m
}

// addToDoubleLabeledMetric adds and labels a double gauge datapoint to a metricslice.
func addToDoubleLabeledMetric(metric pdata.DoubleDataPointSlice, now pdata.Timestamp, labels pdata.StringMap, value float64) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(now)
	dataPoint.SetValue(value)
	labels.CopyTo(dataPoint.LabelsMap())
}

// addToIntLabeledMetric adds and labels a int sum datapoint to metricslice.
func addToIntLabeledMetric(metric pdata.IntDataPointSlice, now pdata.Timestamp, labels pdata.StringMap, value int64) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(now)
	dataPoint.SetValue(value)
	labels.CopyTo(dataPoint.LabelsMap())
}

// scrape scrapes the mysql db metric stats, transforms them and labels them into a metric slices.
func (m *mySQLScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {

	if m.client == nil {
		return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to http client")
	}

	// metric initialization
	rms := pdata.NewResourceMetricsSlice()
	ilm := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/mysql")
	now := pdata.TimestampFromTime(time.Now())

	bufferPoolPages := initMetric(ilm.Metrics(), metadata.M.MysqlBufferPoolPages).DoubleGauge().DataPoints()
	bufferPoolOperations := initMetric(ilm.Metrics(), metadata.M.MysqlBufferPoolOperations).IntSum().DataPoints()
	bufferPoolSize := initMetric(ilm.Metrics(), metadata.M.MysqlBufferPoolSize).DoubleGauge().DataPoints()
	commands := initMetric(ilm.Metrics(), metadata.M.MysqlCommands).IntSum().DataPoints()
	handlers := initMetric(ilm.Metrics(), metadata.M.MysqlHandlers).IntSum().DataPoints()
	doubleWrites := initMetric(ilm.Metrics(), metadata.M.MysqlDoubleWrites).IntSum().DataPoints()
	logOperations := initMetric(ilm.Metrics(), metadata.M.MysqlLogOperations).IntSum().DataPoints()
	operations := initMetric(ilm.Metrics(), metadata.M.MysqlOperations).IntSum().DataPoints()
	pageOperations := initMetric(ilm.Metrics(), metadata.M.MysqlPageOperations).IntSum().DataPoints()
	rowLocks := initMetric(ilm.Metrics(), metadata.M.MysqlRowLocks).IntSum().DataPoints()
	rowOperations := initMetric(ilm.Metrics(), metadata.M.MysqlRowOperations).IntSum().DataPoints()
	locks := initMetric(ilm.Metrics(), metadata.M.MysqlLocks).IntSum().DataPoints()
	sorts := initMetric(ilm.Metrics(), metadata.M.MysqlSorts).IntSum().DataPoints()
	threads := initMetric(ilm.Metrics(), metadata.M.MysqlThreads).DoubleGauge().DataPoints()

	// collect innodb metrics.
	innodbStats, err := m.client.getInnodbStats()
	for _, stat := range innodbStats {
		labels := pdata.NewStringMap()
		switch stat.key {
		case "buffer_pool_size":
			labels.Insert(metadata.L.BufferPoolSizeState, "size")
			addToDoubleLabeledMetric(bufferPoolSize, now, labels, parseFloat(stat.value))
		}
	}
	if err != nil {
		m.logger.Error("Failed to fetch InnoDB stats", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	// collect global status metrics.
	globalStats, err := m.client.getGlobalStats()
	if err != nil {
		m.logger.Error("Failed to fetch global stats", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	for _, stat := range globalStats {
		labels := pdata.NewStringMap()
		switch stat.key {
		// buffer_pool_pages
		case "Innodb_buffer_pool_pages_data":
			labels.Insert(metadata.L.BufferPoolPagesState, "data")
			addToDoubleLabeledMetric(bufferPoolPages, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_pages_dirty":
			labels.Insert(metadata.L.BufferPoolPagesState, "dirty")
			addToDoubleLabeledMetric(bufferPoolPages, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_pages_flushed":
			labels.Insert(metadata.L.BufferPoolPagesState, "flushed")
			addToDoubleLabeledMetric(bufferPoolPages, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_pages_free":
			labels.Insert(metadata.L.BufferPoolPagesState, "free")
			addToDoubleLabeledMetric(bufferPoolPages, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_pages_misc":
			labels.Insert(metadata.L.BufferPoolPagesState, "misc")
			addToDoubleLabeledMetric(bufferPoolPages, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_pages_total":
			labels.Insert(metadata.L.BufferPoolPagesState, "total")
			addToDoubleLabeledMetric(bufferPoolPages, now, labels, parseFloat(stat.value))
			// 	// buffer_pool_operations
		case "Innodb_buffer_pool_read_ahead_rnd":
			labels.Insert(metadata.L.BufferPoolOperationsState, "read_ahead_rnd")
			addToIntLabeledMetric(bufferPoolOperations, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_read_ahead":
			labels.Insert(metadata.L.BufferPoolOperationsState, "read_ahead")
			addToIntLabeledMetric(bufferPoolOperations, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_read_ahead_evicted":
			labels.Insert(metadata.L.BufferPoolOperationsState, "read_ahead_evicted")
			addToIntLabeledMetric(bufferPoolOperations, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_read_requests":
			labels.Insert(metadata.L.BufferPoolOperationsState, "read_requests")
			addToIntLabeledMetric(bufferPoolOperations, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_reads":
			labels.Insert(metadata.L.BufferPoolOperationsState, "reads")
			addToIntLabeledMetric(bufferPoolOperations, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_wait_free":
			labels.Insert(metadata.L.BufferPoolOperationsState, "wait_free")
			addToIntLabeledMetric(bufferPoolOperations, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_write_requests":
			labels.Insert(metadata.L.BufferPoolOperationsState, "write_requests")
			addToIntLabeledMetric(bufferPoolOperations, now, labels, parseInt(stat.value))
			// 	// buffer_pool_size
		case "Innodb_buffer_pool_bytes_data":
			labels.Insert(metadata.L.BufferPoolSizeState, "data")
			addToDoubleLabeledMetric(bufferPoolSize, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_bytes_dirty":
			labels.Insert(metadata.L.BufferPoolSizeState, "dirty")
			addToDoubleLabeledMetric(bufferPoolSize, now, labels, parseFloat(stat.value))
			// 	// commands
		case "Com_stmt_execute":
			labels.Insert(metadata.L.CommandState, "execute")
			addToIntLabeledMetric(commands, now, labels, parseInt(stat.value))
		case "Com_stmt_close":
			labels.Insert(metadata.L.CommandState, "close")
			addToIntLabeledMetric(commands, now, labels, parseInt(stat.value))
		case "Com_stmt_fetch":
			labels.Insert(metadata.L.CommandState, "fetch")
			addToIntLabeledMetric(commands, now, labels, parseInt(stat.value))
		case "Com_stmt_prepare":
			labels.Insert(metadata.L.CommandState, "prepare")
			addToIntLabeledMetric(commands, now, labels, parseInt(stat.value))
		case "Com_stmt_reset":
			labels.Insert(metadata.L.CommandState, "reset")
			addToIntLabeledMetric(commands, now, labels, parseInt(stat.value))
		case "Com_stmt_send_long_data":
			labels.Insert(metadata.L.CommandState, "send_long_data")
			addToIntLabeledMetric(commands, now, labels, parseInt(stat.value))
			// 	// handlers
		case "Handler_commit":
			labels.Insert(metadata.L.HandlerState, "commit")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_delete":
			labels.Insert(metadata.L.HandlerState, "delete")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_discover":
			labels.Insert(metadata.L.HandlerState, "discover")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_external_lock":
			labels.Insert(metadata.L.HandlerState, "lock")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_mrr_init":
			labels.Insert(metadata.L.HandlerState, "mrr_init")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_prepare":
			labels.Insert(metadata.L.HandlerState, "prepare")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_read_first":
			labels.Insert(metadata.L.HandlerState, "read_first")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_read_key":
			labels.Insert(metadata.L.HandlerState, "read_key")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_read_last":
			labels.Insert(metadata.L.HandlerState, "read_last")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_read_next":
			labels.Insert(metadata.L.HandlerState, "read_next")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_read_prev":
			labels.Insert(metadata.L.HandlerState, "read_prev")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_read_rnd":
			labels.Insert(metadata.L.HandlerState, "read_rnd")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_read_rnd_next":
			labels.Insert(metadata.L.HandlerState, "read_rnd_next")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_rollback":
			labels.Insert(metadata.L.HandlerState, "rollback")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_savepoint":
			labels.Insert(metadata.L.HandlerState, "savepoint")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_savepoint_rollback":
			labels.Insert(metadata.L.HandlerState, "savepoint_rollback")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_update":
			labels.Insert(metadata.L.HandlerState, "update")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
		case "Handler_write":
			labels.Insert(metadata.L.HandlerState, "write")
			addToIntLabeledMetric(handlers, now, labels, parseInt(stat.value))
			// 	// double_writes
		case "Innodb_dblwr_pages_written":
			labels.Insert(metadata.L.DoubleWritesState, "written")
			addToIntLabeledMetric(doubleWrites, now, labels, parseInt(stat.value))
		case "Innodb_dblwr_writes":
			labels.Insert(metadata.L.DoubleWritesState, "writes")
			addToIntLabeledMetric(doubleWrites, now, labels, parseInt(stat.value))
			// 	// log_operations
		case "Innodb_log_waits":
			labels.Insert(metadata.L.LogOperationsState, "waits")
			addToIntLabeledMetric(logOperations, now, labels, parseInt(stat.value))
		case "Innodb_log_write_requests":
			labels.Insert(metadata.L.LogOperationsState, "requests")
			addToIntLabeledMetric(logOperations, now, labels, parseInt(stat.value))
		case "Innodb_log_writes":
			labels.Insert(metadata.L.LogOperationsState, "writes")
			addToIntLabeledMetric(logOperations, now, labels, parseInt(stat.value))
			// 	// operations
		case "Innodb_data_fsyncs":
			labels.Insert(metadata.L.OperationsState, "fsyncs")
			addToIntLabeledMetric(operations, now, labels, parseInt(stat.value))
		case "Innodb_data_reads":
			labels.Insert(metadata.L.OperationsState, "reads")
			addToIntLabeledMetric(operations, now, labels, parseInt(stat.value))
		case "Innodb_data_writes":
			labels.Insert(metadata.L.OperationsState, "writes")
			addToIntLabeledMetric(operations, now, labels, parseInt(stat.value))
			// 	// page_operations
		case "Innodb_pages_created":
			labels.Insert(metadata.L.PageOperationsState, "created")
			addToIntLabeledMetric(pageOperations, now, labels, parseInt(stat.value))
		case "Innodb_pages_read":
			labels.Insert(metadata.L.PageOperationsState, "read")
			addToIntLabeledMetric(pageOperations, now, labels, parseInt(stat.value))
		case "Innodb_pages_written":
			labels.Insert(metadata.L.PageOperationsState, "written")
			addToIntLabeledMetric(pageOperations, now, labels, parseInt(stat.value))
			// 	// row_locks
		case "Innodb_row_lock_waits":
			labels.Insert(metadata.L.RowLocksState, "waits")
			addToIntLabeledMetric(rowLocks, now, labels, parseInt(stat.value))
		case "Innodb_row_lock_time":
			labels.Insert(metadata.L.RowLocksState, "time")
			addToIntLabeledMetric(rowLocks, now, labels, parseInt(stat.value))
			// 	// row_operations
		case "Innodb_rows_deleted":
			labels.Insert(metadata.L.RowOperationsState, "deleted")
			addToIntLabeledMetric(rowOperations, now, labels, parseInt(stat.value))
		case "Innodb_rows_inserted":
			labels.Insert(metadata.L.RowOperationsState, "inserted")
			addToIntLabeledMetric(rowOperations, now, labels, parseInt(stat.value))
		case "Innodb_rows_read":
			labels.Insert(metadata.L.RowOperationsState, "read")
			addToIntLabeledMetric(rowOperations, now, labels, parseInt(stat.value))
		case "Innodb_rows_updated":
			labels.Insert(metadata.L.RowOperationsState, "updated")
			addToIntLabeledMetric(rowOperations, now, labels, parseInt(stat.value))
			// 	// locks
		case "Table_locks_immediate":
			labels.Insert(metadata.L.LocksState, "immediate")
			addToIntLabeledMetric(locks, now, labels, parseInt(stat.value))
		case "Table_locks_waited":
			labels.Insert(metadata.L.LocksState, "waited")
			addToIntLabeledMetric(locks, now, labels, parseInt(stat.value))
			// 	// sorts
		case "Sort_merge_passes":
			labels.Insert(metadata.L.SortsState, "merge_passes")
			addToIntLabeledMetric(sorts, now, labels, parseInt(stat.value))
		case "Sort_range":
			labels.Insert(metadata.L.SortsState, "range")
			addToIntLabeledMetric(sorts, now, labels, parseInt(stat.value))
		case "Sort_rows":
			labels.Insert(metadata.L.SortsState, "rows")
			addToIntLabeledMetric(sorts, now, labels, parseInt(stat.value))
		case "Sort_scan":
			labels.Insert(metadata.L.SortsState, "scan")
			addToIntLabeledMetric(sorts, now, labels, parseInt(stat.value))
			// 	// threads
		case "Threads_cached":
			labels.Insert(metadata.L.ThreadsState, "cached")
			addToDoubleLabeledMetric(threads, now, labels, parseFloat(stat.value))
		case "Threads_connected":
			labels.Insert(metadata.L.ThreadsState, "connected")
			addToDoubleLabeledMetric(threads, now, labels, parseFloat(stat.value))
		case "Threads_created":
			labels.Insert(metadata.L.ThreadsState, "created")
			addToDoubleLabeledMetric(threads, now, labels, parseFloat(stat.value))
		case "Threads_running":
			labels.Insert(metadata.L.ThreadsState, "running")
			addToDoubleLabeledMetric(threads, now, labels, parseFloat(stat.value))
		}
	}
	return rms, nil
}

// parseFloat converts string to float64.
func parseFloat(value string) float64 {
	f, _ := strconv.ParseFloat(value, 64)
	return f
}

// parseInt converts string to int64.
func parseInt(value string) int64 {
	i, _ := strconv.ParseInt(value, 10, 64)
	return i
}
