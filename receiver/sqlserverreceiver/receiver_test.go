// Copyright 2020, OpenTelemetry Authors
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

//go:build windows
// +build windows

package sqlserverreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/observIQ/opentelemetry-collector-contrib/receiver/sqlserverreceiver/internal/metadata"
	windowsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

func TestSqlServerReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	r := newSqlServerReceiver(
		componenttest.NewNopReceiverCreateSettings(),
		cfg,
		consumertest.NewNop(),
	)
	require.NotNil(t, r)

	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, r.Shutdown(context.Background()))
}

func TestCreateWindowsReceiverConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   *windowsreceiver.Config
	}{
		{
			name: "Test MetricSettings All Enabled",
			config: &Config{
				ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
					CollectionInterval: 10 * time.Second,
				},
				Metrics: getMetricSettingsEnabled(),
			},
			want: getWindowsReceiverConfigAll(),
		},
		{
			name: "Test MetricSettings Some Disabled",
			config: &Config{
				ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
					CollectionInterval: 10 * time.Second,
				},
				Metrics: getMetricSettingsMix(),
			},
			want: getWindowsReceiverConfigPartial(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricSlice := createMetricSlice(tt.config.Metrics)
			metricPerfCounterConfigs := createMetricPerfCounterConfigs()
			got, err := createWindowsReceiverConfig(tt.config.ScraperControllerSettings, metricSlice, metricPerfCounterConfigs)

			require.Nil(t, err)
			require.Equal(t, got.ScraperControllerSettings, tt.want.ScraperControllerSettings)
			require.Equal(t, got.MetricMetaData, tt.want.MetricMetaData)
			require.ElementsMatch(t, got.PerfCounters, tt.want.PerfCounters)
		})
	}
}

func getMetricSettingsEnabled() metadata.MetricsSettings {
	return metadata.MetricsSettings{
		SqlserverBatchRequestRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverBatchSQLCompilationRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverBatchSQLRecompilationRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverLockWaitRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverLockWaitTimeAvg: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageBufferCacheHitRatio: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageCheckpointFlushRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageLazyWriteRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageLifeExpectancy: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageOperationRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageSplitRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverTransactionLogFlushDataRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverTransactionLogFlushRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverTransactionLogFlushWaitRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverTransactionLogGrowthCount: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverTransactionLogShrinkCount: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverTransactionLogUsage: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverUserConnectionCount: metadata.MetricSettings{
			Enabled: true,
		},
	}
}

func getMetricSettingsMix() metadata.MetricsSettings {
	return metadata.MetricsSettings{
		SqlserverBatchRequestRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverBatchSQLCompilationRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverBatchSQLRecompilationRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverLockWaitRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverLockWaitTimeAvg: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageBufferCacheHitRatio: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageCheckpointFlushRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageLazyWriteRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageLifeExpectancy: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverPageOperationRate: metadata.MetricSettings{
			Enabled: false,
		},
		SqlserverPageSplitRate: metadata.MetricSettings{
			Enabled: true,
		},
		SqlserverTransactionLogFlushDataRate: metadata.MetricSettings{
			Enabled: false,
		},
		SqlserverTransactionLogFlushRate: metadata.MetricSettings{
			Enabled: false,
		},
		SqlserverTransactionLogFlushWaitRate: metadata.MetricSettings{
			Enabled: false,
		},
		SqlserverTransactionLogGrowthCount: metadata.MetricSettings{
			Enabled: false,
		},
		SqlserverTransactionLogShrinkCount: metadata.MetricSettings{
			Enabled: false,
		},
		SqlserverTransactionLogUsage: metadata.MetricSettings{
			Enabled: false,
		},
		SqlserverUserConnectionCount: metadata.MetricSettings{
			Enabled: true,
		},
	}
}

func getWindowsReceiverConfigAll() *windowsreceiver.Config {
	return &windowsreceiver.Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 10 * time.Second,
		},
		MetricMetaData: map[string]windowsreceiver.MetricConfig{
			"sqlserver.batch.request.rate": {
				Unit:        "{requests}/s",
				Description: "Number of batch requests received by SQL Server.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.batch.sql_compilation.rate": {
				Unit:        "{compilations}/s",
				Description: "Number of SQL compilations needed.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.batch.sql_recompilation.rate": {
				Unit:        "{compilations}/s",
				Description: "Number of SQL recompilations needed.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.lock.wait.rate": {
				Unit:        "{requests}/s",
				Description: "Number of lock requests resulting in a wait.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.lock.wait_time.avg": {
				Unit:        "ms",
				Description: "Average wait time for all lock requests that had to wait.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.buffer_cache.hit_ratio": {
				Unit:        "%",
				Description: "Pages found in the buffer pool without having to read from disk.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.checkpoint.flush.rate": {
				Unit:        "{pages}/s",
				Description: "Number of pages flushed by operations requiring dirty pages to be flushed.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.lazy_write.rate": {
				Unit:        "{writes}/s",
				Description: "Number of lazy writes moving dirty pages to disk.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.life_expectancy": {
				Unit:        "s",
				Description: "Time a page will stay in the buffer pool.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.operation.rate": {
				Unit:        "{operations}/s",
				Description: "Number of physical database page operations issued.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.split.rate": {
				Unit:        "{pages}/s",
				Description: "Number of pages split as a result of overflowing index pages.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.transaction_log.flush.data.rate": {
				Unit:        "By/s",
				Description: "Total number of log bytes flushed.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.transaction_log.flush.rate": {
				Unit:        "{flushes}/s",
				Description: "Number of log flushes.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.transaction_log.flush.wait.rate": {
				Unit:        "{commits}/s",
				Description: "Number of commits waiting for a transaction log flush.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.transaction_log.growth.count": {
				Unit:        "{growths}",
				Description: "Total number of transaction log expansions for a database.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.transaction_log.shrink.count": {
				Unit:        "{shrinks}",
				Description: "Total number of transaction log shrinks for a database.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.transaction_log.usage": {
				Unit:        "%",
				Description: "Percent of transaction log space used.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.user.connection.count": {
				Unit:        "{connections}",
				Description: "Number of users connected to the SQL Server.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
		},
		PerfCounters: []windowsreceiver.PerfCounterConfig{
			{
				Object: "SQLServer:General Statistics",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.user.connection.count",
						Name:   "User Connections",
					},
				},
			},
			{
				Object: "SQLServer:SQL Statistics",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.batch.request.rate",
						Name:   "Batch Requests/sec",
					},
				},
			},
			{
				Object: "SQLServer:SQL Statistics",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.batch.sql_compilation.rate",
						Name:   "SQL Compilations/sec",
					},
				},
			},
			{
				Object: "SQLServer:SQL Statistics",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.batch.sql_recompilation.rate",
						Name:   "SQL Re-Compilations/sec",
					},
				},
			},
			{
				Object:    "SQLServer:Locks",
				Instances: []string{"_Total"},
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.lock.wait.rate",
						Name:   "Lock Waits/sec",
					},
				},
			},
			{
				Object:    "SQLServer:Locks",
				Instances: []string{"_Total"},
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.lock.wait_time.avg",
						Name:   "Average Wait Time (ms)",
					},
				},
			},
			{
				Object: "SQLServer:Buffer Manager",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.page.buffer_cache.hit_ratio",
						Name:   "Buffer cache hit ratio",
					},
				},
			},
			{
				Object: "SQLServer:Buffer Manager",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.page.checkpoint.flush.rate",
						Name:   "Checkpoint pages/sec",
					},
				},
			},
			{
				Object: "SQLServer:Buffer Manager",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.page.lazy_write.rate",
						Name:   "Lazy Writes/sec",
					},
				},
			},
			{
				Object: "SQLServer:Buffer Manager",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.page.life_expectancy",
						Name:   "Page life expectancy",
					},
				},
			},
			{
				Object: "SQLServer:Buffer Manager",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric:     "page.operation.rate",
						Name:       "Page reads/sec",
						Attributes: map[string]string{"operation": "read"},
					},
					{
						Metric:     "page.operation.rate",
						Name:       "Page writes/sec",
						Attributes: map[string]string{"operation": "write"},
					},
				},
			},
			{
				Object: "SQLServer:Access Methods",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.page.split.rate",
						Name:   "Page Splits/sec",
					},
				},
			},
			{
				Object:    "SQLServer:Databases",
				Instances: []string{"*"},
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.transaction_log.flush.data.rate",
						Name:   "Log Bytes Flushed/sec",
					},
				},
			},
			{
				Object:    "SQLServer:Databases",
				Instances: []string{"*"},
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.transaction_log.flush.rate",
						Name:   "Log Flushes/sec",
					},
				},
			},
			{
				Object:    "SQLServer:Databases",
				Instances: []string{"*"},
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.transaction_log.flush.wait.rate",
						Name:   "Log Flush Waits/sec",
					},
				},
			},
			{
				Object:    "SQLServer:Databases",
				Instances: []string{"*"},
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.transaction_log.growth.count",
						Name:   "Log Growths",
					},
				},
			},
			{
				Object:    "SQLServer:Databases",
				Instances: []string{"*"},
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.transaction_log.shrink.count",
						Name:   "Log Shrinks",
					},
				},
			},
			{
				Object:    "SQLServer:Databases",
				Instances: []string{"*"},
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.transaction_log.usage",
						Name:   "Percent Log Used",
					},
				},
			},
		},
	}
}

func getWindowsReceiverConfigPartial() *windowsreceiver.Config {
	return &windowsreceiver.Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 10 * time.Second,
		},
		MetricMetaData: map[string]windowsreceiver.MetricConfig{
			"sqlserver.batch.request.rate": {
				Unit:        "{requests}/s",
				Description: "Number of batch requests received by SQL Server.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.batch.sql_compilation.rate": {
				Unit:        "{compilations}/s",
				Description: "Number of SQL compilations needed.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.batch.sql_recompilation.rate": {
				Unit:        "{compilations}/s",
				Description: "Number of SQL recompilations needed.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.lock.wait.rate": {
				Unit:        "{requests}/s",
				Description: "Number of lock requests resulting in a wait.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.lock.wait_time.avg": {
				Unit:        "ms",
				Description: "Average wait time for all lock requests that had to wait.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.buffer_cache.hit_ratio": {
				Unit:        "%",
				Description: "Pages found in the buffer pool without having to read from disk.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.checkpoint.flush.rate": {
				Unit:        "{pages}/s",
				Description: "Number of pages flushed by operations requiring dirty pages to be flushed.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.lazy_write.rate": {
				Unit:        "{writes}/s",
				Description: "Number of lazy writes moving dirty pages to disk.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.life_expectancy": {
				Unit:        "s",
				Description: "Time a page will stay in the buffer pool.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.page.split.rate": {
				Unit:        "{pages}/s",
				Description: "Number of pages split as a result of overflowing index pages.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
			"sqlserver.user.connection.count": {
				Unit:        "{connections}",
				Description: "Number of users connected to the SQL Server.",
				Gauge:       windowsreceiver.GaugeMetric{},
			},
		},
		PerfCounters: []windowsreceiver.PerfCounterConfig{
			{
				Object: "SQLServer:General Statistics",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.user.connection.count",
						Name:   "User Connections",
					},
				},
			},
			{
				Object: "SQLServer:SQL Statistics",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.batch.request.rate",
						Name:   "Batch Requests/sec",
					},
				},
			},
			{
				Object: "SQLServer:SQL Statistics",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.batch.sql_compilation.rate",
						Name:   "SQL Compilations/sec",
					},
				},
			},
			{
				Object: "SQLServer:SQL Statistics",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.batch.sql_recompilation.rate",
						Name:   "SQL Re-Compilations/sec",
					},
				},
			},
			{
				Object:    "SQLServer:Locks",
				Instances: []string{"_Total"},
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.lock.wait.rate",
						Name:   "Lock Waits/sec",
					},
				},
			},
			{
				Object:    "SQLServer:Locks",
				Instances: []string{"_Total"},
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.lock.wait_time.avg",
						Name:   "Average Wait Time (ms)",
					},
				},
			},
			{
				Object: "SQLServer:Buffer Manager",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.page.buffer_cache.hit_ratio",
						Name:   "Buffer cache hit ratio",
					},
				},
			},
			{
				Object: "SQLServer:Buffer Manager",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.page.checkpoint.flush.rate",
						Name:   "Checkpoint pages/sec",
					},
				},
			},
			{
				Object: "SQLServer:Buffer Manager",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.page.lazy_write.rate",
						Name:   "Lazy Writes/sec",
					},
				},
			},
			{
				Object: "SQLServer:Buffer Manager",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.page.life_expectancy",
						Name:   "Page life expectancy",
					},
				},
			},
			{
				Object: "SQLServer:Access Methods",
				Counters: []windowsreceiver.CounterConfig{
					{
						Metric: "sqlserver.page.split.rate",
						Name:   "Page Splits/sec",
					},
				},
			},
		},
	}
}
