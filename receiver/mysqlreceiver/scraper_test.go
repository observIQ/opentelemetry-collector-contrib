package mysqlreceiver

import (
	"context"
	"fmt"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver/internal/metadata"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	mysqlMock := fakeClient{}
	sc := newMySQLScraper(zap.NewNop(), &Config{
		User:     "otel",
		Password: "otel",
		Addr:     "127.0.0.1",
		Port:     3306,
	})
	sc.client = &mysqlMock

	// err := sc.start(context.Background(), componenttest.NewNopHost())
	// require.NoError(t, err)

	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 67, ms.Len())
	require.Equal(t, 14, len(metadata.M.Names()))

	floatMetrics := map[string]float64{}
	intMetrics := map[string]int64{}

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)

		fmt.Println(m.DataType())
		switch m.DataType() {
		case pdata.MetricDataTypeDoubleGauge:
			fmt.Println(m.Name(), m.DoubleGauge().DataPoints().At(0).Value())
			floatMetrics[m.Name()] = m.DoubleGauge().DataPoints().At(0).Value()
		case pdata.MetricDataTypeIntSum:
			fmt.Println(m.Name(), m.IntSum().DataPoints().At(0).Value())
			intMetrics[m.Name()] = m.IntSum().DataPoints().At(0).Value()
		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}

	require.Equal(t, 13, len(floatMetrics))
	require.Equal(t, map[string]float64{
		"Innodb_buffer_pool_bytes_data":    1.6072704e+07,
		"Innodb_buffer_pool_bytes_dirty":   0,
		"Innodb_buffer_pool_pages_data":    981,
		"Innodb_buffer_pool_pages_dirty":   0,
		"Innodb_buffer_pool_pages_flushed": 168,
		"Innodb_buffer_pool_pages_free":    7207,
		"Innodb_buffer_pool_pages_misc":    4,
		"Innodb_buffer_pool_pages_total":   8192,
		"Threads_cached":                   0,
		"Threads_connected":                1,
		"Threads_created":                  1,
		"Threads_running":                  2,
		"buffer_pool_size":                 1.34217728e+08}, floatMetrics)

	require.Equal(t, map[string]int64{
		"Com_stmt_close":                        0,
		"Com_stmt_execute":                      0,
		"Com_stmt_fetch":                        0,
		"Com_stmt_prepare":                      0,
		"Com_stmt_reset":                        0,
		"Com_stmt_send_long_data":               0,
		"Handler_commit":                        564,
		"Handler_delete":                        0,
		"Handler_discover":                      0,
		"Handler_external_lock":                 7127,
		"Handler_mrr_init":                      0,
		"Handler_prepare":                       0,
		"Handler_read_first":                    52,
		"Handler_read_key":                      1680,
		"Handler_read_last":                     0,
		"Handler_read_next":                     3960,
		"Handler_read_prev":                     0,
		"Handler_read_rnd":                      0,
		"Handler_read_rnd_next":                 505063,
		"Handler_rollback":                      0,
		"Handler_savepoint":                     0,
		"Handler_savepoint_rollback":            0,
		"Handler_update":                        315,
		"Handler_write":                         256734,
		"Innodb_buffer_pool_read_ahead":         0,
		"Innodb_buffer_pool_read_ahead_evicted": 0,
		"Innodb_buffer_pool_read_ahead_rnd":     0,
		"Innodb_buffer_pool_read_requests":      14837,
		"Innodb_buffer_pool_reads":              838,
		"Innodb_buffer_pool_wait_free":          0,
		"Innodb_buffer_pool_write_requests":     1669,
		"Innodb_data_fsyncs":                    46,
		"Innodb_data_reads":                     860,
		"Innodb_data_writes":                    215,
		"Innodb_dblwr_pages_written":            27,
		"Innodb_dblwr_writes":                   8,
		"Innodb_log_waits":                      0,
		"Innodb_log_write_requests":             646,
		"Innodb_log_writes":                     14,
		"Innodb_pages_created":                  144,
		"Innodb_pages_read":                     837,
		"Innodb_pages_written":                  168,
		"Innodb_row_lock_time":                  0,
		"Innodb_row_lock_waits":                 0,
		"Innodb_rows_deleted":                   0,
		"Innodb_rows_inserted":                  0,
		"Innodb_rows_read":                      0,
		"Innodb_rows_updated":                   0,
		"Sort_merge_passes":                     0,
		"Sort_range":                            0,
		"Sort_rows":                             0,
		"Sort_scan":                             0,
		"Table_locks_immediate":                 521,
		"Table_locks_waited":                    0,
	}, intMetrics)
}

func TestScrapeErrorBadConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		user     string
		password string
		address  string
		port     int64
	}{
		{
			desc:     "no user",
			password: "otel",
			address:  "127.0.0.1",
			port:     3306,
		},
		{
			user:    "",
			address: "127.0.0.1",
			port:    3306,
		},
		{
			desc:     "no address",
			user:     "",
			password: "otel",
			port:     3306,
		},
		{
			desc:     "no port",
			user:     "",
			password: "otel",
			address:  "127.0.0.1",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mysqlMock := fakeClient{}
			sc := newMySQLScraper(zap.NewNop(), &Config{
				User:     tC.user,
				Password: tC.password,
				Addr:     tC.address,
				Port:     int(tC.port),
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
			Addr:     "127.0.0.1",
			Port:     3306,
		})
		sc.client = &mysqlMock
		err := sc.start(context.Background(), componenttest.NewNopHost())
		require.Nil(t, err)
	})
}
