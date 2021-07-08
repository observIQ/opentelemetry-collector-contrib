package mysqlreceiver

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"

	"go.uber.org/zap"
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

func (m *mySQLScraper) start(_ context.Context, host component.Host) error {
	if m.config.User == "" || m.config.Password == "" || m.config.Addr == "" || m.config.Port == 0 {
		return errors.New("missing database configuration parameters")
	}
	client, err := newMySQLClient(mySQLConfig{
		user: m.config.User,
		pass: m.config.Password,
		addr: m.config.Addr,
		port: m.config.Port,
	})
	if err != nil {
		return err
	}
	m.client = client

	return nil
}

func (m *mySQLScraper) shutdown(context.Context) error {
	if !m.client.Closed() {
		m.logger.Info("gracefully shutdown")
		return m.client.Close()
	}
	return nil
}

func (m *mySQLScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {

	if m.client == nil {
		return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to http client")
	}

	now := pdata.TimestampFromTime(time.Now())
	metrics := pdata.NewMetrics()
	ilm := metrics.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/mysql")

	innodbStats, err := m.client.getInnodbStats()

	for _, stat := range innodbStats {
		labels := pdata.NewStringMap()
		switch stat.key {
		case "buffer_pool_size":
			labels.Insert(metadata.L.BufferPoolSizeState, "size")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		}
	}

	if err != nil {
		m.logger.Error("Failed to fetch InnoDB stats", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

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
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_pages_dirty":
			labels.Insert(metadata.L.BufferPoolPagesState, "dirty")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_pages_flushed":
			labels.Insert(metadata.L.BufferPoolPagesState, "flushed")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_pages_free":
			labels.Insert(metadata.L.BufferPoolPagesState, "free")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_pages_misc":
			labels.Insert(metadata.L.BufferPoolPagesState, "misc")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_pages_total":
			labels.Insert(metadata.L.BufferPoolPagesState, "total")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
			// 	// buffer_pool_operations
		case "Innodb_buffer_pool_read_ahead_rnd":
			labels.Insert(metadata.L.BufferPoolOperationsState, "read_ahead_rnd")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))

		case "Innodb_buffer_pool_read_ahead":
			labels.Insert(metadata.L.BufferPoolOperationsState, "read_ahead")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_read_ahead_evicted":
			labels.Insert(metadata.L.BufferPoolOperationsState, "read_ahead_evicted")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_read_requests":
			labels.Insert(metadata.L.BufferPoolOperationsState, "read_requests")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_reads":
			labels.Insert(metadata.L.BufferPoolOperationsState, "reads")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_wait_free":
			labels.Insert(metadata.L.BufferPoolOperationsState, "wait_free")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_buffer_pool_write_requests":
			labels.Insert(metadata.L.BufferPoolOperationsState, "write_requests")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// buffer_pool_size
		case "Innodb_buffer_pool_bytes_data":
			labels.Insert(metadata.L.BufferPoolSizeState, "data")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		case "Innodb_buffer_pool_bytes_dirty":
			labels.Insert(metadata.L.BufferPoolSizeState, "dirty")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
			// 	// commands
		case "Com_stmt_execute":
			labels.Insert(metadata.L.CommandState, "execute")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Com_stmt_close":
			labels.Insert(metadata.L.CommandState, "close")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Com_stmt_fetch":
			labels.Insert(metadata.L.CommandState, "fetch")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Com_stmt_prepare":
			labels.Insert(metadata.L.CommandState, "prepare")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Com_stmt_reset":
			labels.Insert(metadata.L.CommandState, "reset")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Com_stmt_send_long_data":
			labels.Insert(metadata.L.CommandState, "send_long_data")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// handlers
		case "Handler_commit":
			labels.Insert(metadata.L.HandlerState, "commit")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_delete":
			labels.Insert(metadata.L.HandlerState, "delete")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_discover":
			labels.Insert(metadata.L.HandlerState, "discover")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_external_lock":
			labels.Insert(metadata.L.HandlerState, "lock")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_mrr_init":
			labels.Insert(metadata.L.HandlerState, "mrr_init")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_prepare":
			labels.Insert(metadata.L.HandlerState, "prepare")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_read_first":
			labels.Insert(metadata.L.HandlerState, "read_first")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_read_key":
			labels.Insert(metadata.L.HandlerState, "read_key")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_read_last":
			labels.Insert(metadata.L.HandlerState, "read_last")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_read_next":
			labels.Insert(metadata.L.HandlerState, "read_next")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_read_prev":
			labels.Insert(metadata.L.HandlerState, "read_prev")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_read_rnd":
			labels.Insert(metadata.L.HandlerState, "read_rnd")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_read_rnd_next":
			labels.Insert(metadata.L.HandlerState, "read_rnd_next")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_rollback":
			labels.Insert(metadata.L.HandlerState, "rollback")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_savepoint":
			labels.Insert(metadata.L.HandlerState, "savepoint")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_savepoint_rollback":
			labels.Insert(metadata.L.HandlerState, "savepoint_rollback")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_update":
			labels.Insert(metadata.L.HandlerState, "update")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Handler_write":
			labels.Insert(metadata.L.HandlerState, "write")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// double_writes
		case "Innodb_dblwr_pages_written":
			labels.Insert(metadata.L.DoubleWritesState, "written")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_dblwr_writes":
			labels.Insert(metadata.L.DoubleWritesState, "writes")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// log_operations
		case "Innodb_log_waits":
			labels.Insert(metadata.L.LogOperationsState, "waits")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_log_write_requests":
			labels.Insert(metadata.L.LogOperationsState, "requests")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_log_writes":
			labels.Insert(metadata.L.LogOperationsState, "writes")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// operations
		case "Innodb_data_fsyncs":
			labels.Insert(metadata.L.OperationsState, "fsyncs")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_data_reads":
			labels.Insert(metadata.L.OperationsState, "reads")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_data_writes":
			labels.Insert(metadata.L.OperationsState, "writes")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// page_operations
		case "Innodb_pages_created":
			labels.Insert(metadata.L.PageOperationsState, "created")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_pages_read":
			labels.Insert(metadata.L.PageOperationsState, "read")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_pages_written":
			labels.Insert(metadata.L.PageOperationsState, "written")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// row_locks
		case "Innodb_row_lock_waits":
			labels.Insert(metadata.L.RowLocksState, "waits")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_row_lock_time":
			labels.Insert(metadata.L.RowLocksState, "time")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// row_operations
		case "Innodb_rows_deleted":
			labels.Insert(metadata.L.RowOperationsState, "deleted")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_rows_inserted":
			labels.Insert(metadata.L.RowOperationsState, "inserted")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_rows_read":
			labels.Insert(metadata.L.RowOperationsState, "read")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Innodb_rows_updated":
			labels.Insert(metadata.L.RowOperationsState, "updated")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// locks
		case "Table_locks_immediate":
			labels.Insert(metadata.L.LocksState, "immediate")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Table_locks_waited":
			labels.Insert(metadata.L.LocksState, "waited")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// sorts
		case "Sort_merge_passes":
			labels.Insert(metadata.L.SortsState, "merge_passes")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Sort_range":
			labels.Insert(metadata.L.SortsState, "range")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Sort_rows":
			labels.Insert(metadata.L.SortsState, "rows")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
		case "Sort_scan":
			labels.Insert(metadata.L.SortsState, "scan")
			addIntSum(ilm.Metrics(), stat.key, now, labels, parseInt(stat.value))
			// 	// threads
		case "Threads_cached":
			labels.Insert(metadata.L.ThreadsState, "cached")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		case "Threads_connected":
			labels.Insert(metadata.L.ThreadsState, "connected")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		case "Threads_created":
			labels.Insert(metadata.L.ThreadsState, "created")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		case "Threads_running":
			labels.Insert(metadata.L.ThreadsState, "running")
			addDoubleGauge(ilm.Metrics(), stat.key, now, labels, parseFloat(stat.value))
		}
	}
	return metrics.ResourceMetrics(), nil
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

func addDoubleGauge(ms pdata.MetricSlice, name string, now pdata.Timestamp, labels pdata.StringMap, value float64) {
	m := ms.AppendEmpty()
	m.SetName(name)
	m.SetDataType(pdata.MetricDataTypeDoubleGauge)
	dp := m.DoubleGauge().DataPoints().AppendEmpty()
	dp.SetTimestamp(now)
	dp.SetValue(value)
	labels.CopyTo(dp.LabelsMap())
}

func addIntSum(ms pdata.MetricSlice, name string, now pdata.Timestamp, labels pdata.StringMap, value int64) {
	m := ms.AppendEmpty()
	m.SetName(name)
	m.SetDataType(pdata.MetricDataTypeIntSum)
	dp := m.IntSum().DataPoints().AppendEmpty()
	dp.SetTimestamp(now)
	dp.SetValue(value)
	labels.CopyTo(dp.LabelsMap())
}
