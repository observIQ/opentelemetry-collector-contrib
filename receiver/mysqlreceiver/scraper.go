package mysqlreceiver

import (
	"context"
	"errors"
	"strconv"

	// "time"

	// "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"

	//"go.opentelemetry.io/collector/receiver/scraperhelper"

	//"go.opentelemetry.io/collector/consumer/simple"
	"go.uber.org/zap"
)

type mySQLScraper struct {
	client client

	logger *zap.Logger
	config *Config
}

// func newMySQLScraper(
// 	logger *zap.Logger,
// 	config *Config,
// ) scraperhelper.Scraper {
// 	sc := mySQLScraper{
// 		logger: logger,
// 		config: config,
// 	}
// 	return scraperhelper.NewResourceMetricsScraper(config.ID(), sc.scrape())
// }

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
	m.logger.Error("Hello world")
	// if m.client == nil {
	// 	return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to http client")
	// }

	// now := pdata.TimestampFromTime(time.Now())
	// metrics := pdata.NewMetrics()
	// ilm := metrics.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	// ilm.InstrumentationLibrary().SetName("otel/mysql")

	// metrics := simple.Metrics{
	// 	Metrics:   pdata.NewMetrics(),
	// 	Timestamp: time.Now(),
	// 	//MetricFactoriesByName:      metadata.M.FactoriesByName(),
	// 	InstrumentationLibraryName: "otelcol/mysql",
	// }

	// innodbStats, _ := m.client.getInnodbStats()

	// for _, stat := range innodbStats {
	// 	switch stat.key {
	// 	case "buffer_pool_size":
	// 		// metrics.WithLabels(map[string]string{metadata.L.BufferPoolSizeState: "size"}).AddDGaugeDataPoint(metadata.M.MysqlBufferPoolSize.Name(), parseFloat(stat.value))
	// 		metric := ilm.Metrics().AppendEmpty()
	// 		//initFunc(metric)
	// 		dp := metric.DoubleGauge().DataPoints().AppendEmpty()
	// 		dp.SetStartTimestamp(now)
	// 		dp.SetValue(parseFloat(stat.value))
	// 	}
	// }

	// for _, stat := range innodbStats {
	// 	switch stat.key {
	// 	case "buffer_pool_size":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolSizeState: "size"}).AddDGaugeDataPoint(metadata.M.MysqlBufferPoolSize.Name(), parseFloat(stat.value))
	// 	}
	// }

	// if err != nil {
	// 	m.logger.Error("Failed to fetch InnoDB stats", zap.Error(err))
	// 	return pdata.ResourceMetricsSlice{}, err
	// }

	// globalStats, err := m.client.getGlobalStats()
	// if err != nil {
	// 	m.logger.Error("Failed to fetch global stats", zap.Error(err))
	// 	return pdata.ResourceMetricsSlice{}, err
	// }

	// for _, stat := range globalStats {
	// 	switch stat.key {
	// 	// buffer_pool_pages
	// 	case "Innodb_buffer_pool_pages_data":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolPagesState: "data"}).AddDGaugeDataPoint(metadata.M.MysqlBufferPoolPages.Name(), parseFloat(stat.value))
	// 	case "Innodb_buffer_pool_pages_dirty":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolPagesState: "dirty"}).AddDGaugeDataPoint(metadata.M.MysqlBufferPoolPages.Name(), parseFloat(stat.value))
	// 	case "Innodb_buffer_pool_pages_flushed":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolPagesState: "flushed"}).AddDGaugeDataPoint(metadata.M.MysqlBufferPoolPages.Name(), parseFloat(stat.value))
	// 	case "Innodb_buffer_pool_pages_free":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolPagesState: "free"}).AddDGaugeDataPoint(metadata.M.MysqlBufferPoolPages.Name(), parseFloat(stat.value))
	// 	case "Innodb_buffer_pool_pages_misc":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolPagesState: "misc"}).AddDGaugeDataPoint(metadata.M.MysqlBufferPoolPages.Name(), parseFloat(stat.value))
	// 	case "Innodb_buffer_pool_pages_total":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolPagesState: "total"}).AddDGaugeDataPoint(metadata.M.MysqlBufferPoolPages.Name(), parseFloat(stat.value))
	// 	// buffer_pool_operations
	// 	case "Innodb_buffer_pool_read_ahead_rnd":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolOperationsState: "read_ahead_rnd"}).AddSumDataPoint(metadata.M.MysqlBufferPoolOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_buffer_pool_read_ahead":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolOperationsState: "read_ahead"}).AddSumDataPoint(metadata.M.MysqlBufferPoolOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_buffer_pool_read_ahead_evicted":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolOperationsState: "read_ahead_evicted"}).AddSumDataPoint(metadata.M.MysqlBufferPoolOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_buffer_pool_read_requests":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolOperationsState: "read_requests"}).AddSumDataPoint(metadata.M.MysqlBufferPoolOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_buffer_pool_reads":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolOperationsState: "reads"}).AddSumDataPoint(metadata.M.MysqlBufferPoolOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_buffer_pool_wait_free":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolOperationsState: "wait_free"}).AddSumDataPoint(metadata.M.MysqlBufferPoolOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_buffer_pool_write_requests":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolOperationsState: "write_requests"}).AddSumDataPoint(metadata.M.MysqlBufferPoolOperations.Name(), parseInt(stat.value))
	// 	// buffer_pool_size
	// 	case "Innodb_buffer_pool_bytes_data":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolSizeState: "data"}).AddDGaugeDataPoint(metadata.M.MysqlBufferPoolSize.Name(), parseFloat(stat.value))
	// 	case "Innodb_buffer_pool_bytes_dirty":
	// 		metrics.WithLabels(map[string]string{metadata.L.BufferPoolSizeState: "dirty"}).AddDGaugeDataPoint(metadata.M.MysqlBufferPoolSize.Name(), parseFloat(stat.value))
	// 	// commands
	// 	case "Com_stmt_execute":
	// 		metrics.WithLabels(map[string]string{metadata.L.CommandState: "execute"}).AddSumDataPoint(metadata.M.MysqlCommands.Name(), parseInt(stat.value))
	// 	case "Com_stmt_close":
	// 		metrics.WithLabels(map[string]string{metadata.L.CommandState: "close"}).AddSumDataPoint(metadata.M.MysqlCommands.Name(), parseInt(stat.value))
	// 	case "Com_stmt_fetch":
	// 		metrics.WithLabels(map[string]string{metadata.L.CommandState: "fetch"}).AddSumDataPoint(metadata.M.MysqlCommands.Name(), parseInt(stat.value))
	// 	case "Com_stmt_prepare":
	// 		metrics.WithLabels(map[string]string{metadata.L.CommandState: "prepare"}).AddSumDataPoint(metadata.M.MysqlCommands.Name(), parseInt(stat.value))
	// 	case "Com_stmt_reset":
	// 		metrics.WithLabels(map[string]string{metadata.L.CommandState: "reset"}).AddSumDataPoint(metadata.M.MysqlCommands.Name(), parseInt(stat.value))
	// 	case "Com_stmt_send_long_data":
	// 		metrics.WithLabels(map[string]string{metadata.L.CommandState: "send_long_data"}).AddSumDataPoint(metadata.M.MysqlCommands.Name(), parseInt(stat.value))
	// 	// handlers
	// 	case "Handler_commit":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "commit"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_delete":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "delete"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_discover":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "discover"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_external_lock":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "lock"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_mrr_init":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "mrr_init"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_prepare":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "prepare"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_read_first":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "read_first"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_read_key":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "read_key"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_read_last":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "read_last"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_read_next":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "read_next"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_read_prev":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "read_prev"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_read_rnd":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "read_rnd"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_read_rnd_next":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "read_rnd_next"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_rollback":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "rollback"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_savepoint":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "savepoint"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_savepoint_rollback":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "savepoint_rollback"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_update":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "update"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	case "Handler_write":
	// 		metrics.WithLabels(map[string]string{metadata.L.HandlerState: "write"}).AddSumDataPoint(metadata.M.MysqlHandlers.Name(), parseInt(stat.value))
	// 	// double_writes
	// 	case "Innodb_dblwr_pages_written":
	// 		metrics.WithLabels(map[string]string{metadata.L.DoubleWritesState: "written"}).AddSumDataPoint(metadata.M.MysqlDoubleWrites.Name(), parseInt(stat.value))
	// 	case "Innodb_dblwr_writes":
	// 		metrics.WithLabels(map[string]string{metadata.L.DoubleWritesState: "writes"}).AddSumDataPoint(metadata.M.MysqlDoubleWrites.Name(), parseInt(stat.value))
	// 	// log_operations
	// 	case "Innodb_log_waits":
	// 		metrics.WithLabels(map[string]string{metadata.L.LogOperationsState: "waits"}).AddSumDataPoint(metadata.M.MysqlLogOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_log_write_requests":
	// 		metrics.WithLabels(map[string]string{metadata.L.LogOperationsState: "requests"}).AddSumDataPoint(metadata.M.MysqlLogOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_log_writes":
	// 		metrics.WithLabels(map[string]string{metadata.L.LogOperationsState: "writes"}).AddSumDataPoint(metadata.M.MysqlLogOperations.Name(), parseInt(stat.value))
	// 	// operations
	// 	case "Innodb_data_fsyncs":
	// 		metrics.WithLabels(map[string]string{metadata.L.OperationsState: "fsyncs"}).AddSumDataPoint(metadata.M.MysqlOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_data_reads":
	// 		metrics.WithLabels(map[string]string{metadata.L.OperationsState: "reads"}).AddSumDataPoint(metadata.M.MysqlOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_data_writes":
	// 		metrics.WithLabels(map[string]string{metadata.L.OperationsState: "writes"}).AddSumDataPoint(metadata.M.MysqlOperations.Name(), parseInt(stat.value))
	// 	// page_operations
	// 	case "Innodb_pages_created":
	// 		metrics.WithLabels(map[string]string{metadata.L.PageOperationsState: "created"}).AddSumDataPoint(metadata.M.MysqlPageOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_pages_read":
	// 		metrics.WithLabels(map[string]string{metadata.L.PageOperationsState: "read"}).AddSumDataPoint(metadata.M.MysqlPageOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_pages_written":
	// 		metrics.WithLabels(map[string]string{metadata.L.PageOperationsState: "written"}).AddSumDataPoint(metadata.M.MysqlPageOperations.Name(), parseInt(stat.value))
	// 	// row_locks
	// 	case "Innodb_row_lock_waits":
	// 		metrics.WithLabels(map[string]string{metadata.L.RowLocksState: "waits"}).AddSumDataPoint(metadata.M.MysqlRowLocks.Name(), parseInt(stat.value))
	// 	case "Innodb_row_lock_time":
	// 		metrics.WithLabels(map[string]string{metadata.L.RowLocksState: "time"}).AddSumDataPoint(metadata.M.MysqlRowLocks.Name(), parseInt(stat.value))
	// 	// row_operations
	// 	case "Innodb_rows_deleted":
	// 		metrics.WithLabels(map[string]string{metadata.L.RowOperationsState: "deleted"}).AddSumDataPoint(metadata.M.MysqlRowOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_rows_inserted":
	// 		metrics.WithLabels(map[string]string{metadata.L.RowOperationsState: "inserted"}).AddSumDataPoint(metadata.M.MysqlRowOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_rows_read":
	// 		metrics.WithLabels(map[string]string{metadata.L.RowOperationsState: "read"}).AddSumDataPoint(metadata.M.MysqlRowOperations.Name(), parseInt(stat.value))
	// 	case "Innodb_rows_updated":
	// 		metrics.WithLabels(map[string]string{metadata.L.RowOperationsState: "updated"}).AddSumDataPoint(metadata.M.MysqlRowOperations.Name(), parseInt(stat.value))
	// 	// locks
	// 	case "Table_locks_immediate":
	// 		metrics.WithLabels(map[string]string{metadata.L.LocksState: "immediate"}).AddSumDataPoint(metadata.M.MysqlLocks.Name(), parseInt(stat.value))
	// 	case "Table_locks_waited":
	// 		metrics.WithLabels(map[string]string{metadata.L.LocksState: "waited"}).AddSumDataPoint(metadata.M.MysqlLocks.Name(), parseInt(stat.value))
	// 	// sorts
	// 	case "Sort_merge_passes":
	// 		metrics.WithLabels(map[string]string{metadata.L.SortsState: "merge_passes"}).AddSumDataPoint(metadata.M.MysqlSorts.Name(), parseInt(stat.value))
	// 	case "Sort_range":
	// 		metrics.WithLabels(map[string]string{metadata.L.SortsState: "range"}).AddSumDataPoint(metadata.M.MysqlSorts.Name(), parseInt(stat.value))
	// 	case "Sort_rows":
	// 		metrics.WithLabels(map[string]string{metadata.L.SortsState: "rows"}).AddSumDataPoint(metadata.M.MysqlSorts.Name(), parseInt(stat.value))
	// 	case "Sort_scan":
	// 		metrics.WithLabels(map[string]string{metadata.L.SortsState: "scan"}).AddSumDataPoint(metadata.M.MysqlSorts.Name(), parseInt(stat.value))
	// 	// threads
	// 	case "Threads_cached":
	// 		metrics.WithLabels(map[string]string{metadata.L.ThreadsState: "cached"}).AddDGaugeDataPoint(metadata.M.MysqlThreads.Name(), parseFloat(stat.value))
	// 	case "Threads_connected":
	// 		metrics.WithLabels(map[string]string{metadata.L.ThreadsState: "connected"}).AddDGaugeDataPoint(metadata.M.MysqlThreads.Name(), parseFloat(stat.value))
	// 	case "Threads_created":
	// 		metrics.WithLabels(map[string]string{metadata.L.ThreadsState: "created"}).AddDGaugeDataPoint(metadata.M.MysqlThreads.Name(), parseFloat(stat.value))
	// 	case "Threads_running":
	// 		metrics.WithLabels(map[string]string{metadata.L.ThreadsState: "running"}).AddDGaugeDataPoint(metadata.M.MysqlThreads.Name(), parseFloat(stat.value))
	// 	}
	// }
	return pdata.NewResourceMetricsSlice(), nil
	//return metrics.ResourceMetrics(), nil
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
