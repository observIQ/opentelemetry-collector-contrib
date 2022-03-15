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

package sqlserverreceiver // import "github.com/observIQ/opentelemetry-collector-contrib/receiver/sqlserverreceiver"

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/observIQ/opentelemetry-collector-contrib/receiver/sqlserverreceiver/internal/metadata"
	windowsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver"
)

type sqlServerReceiver struct {
	params          component.ReceiverCreateSettings
	config          *Config
	consumer        consumer.Metrics
	windowsReceiver component.MetricsReceiver
}

// newSqlServerReceiver returns a sqlServerReceiver.
func newSqlServerReceiver(params component.ReceiverCreateSettings, cfg *Config, consumer consumer.Metrics) *sqlServerReceiver {
	return &sqlServerReceiver{params: params, config: cfg, consumer: consumer}
}

// Start creates and starts the Windows performance counter receiver.
func (r *sqlServerReceiver) Start(ctx context.Context, host component.Host) error {
	windowsFactory := windowsreceiver.NewFactory()

	metricSlice := createMetricSlice(r.config.Metrics)
	metricPerfCounterConfigs := createMetricPerfCounterConfigs()
	windowsConfig, err := createWindowsReceiverConfig(r.config.ScraperControllerSettings, metricSlice, metricPerfCounterConfigs)
	if err != nil {
		return fmt.Errorf("failed to create windows performance counter receiver config: %v", err)
	}

	windowsReceiver, err := windowsFactory.CreateMetricsReceiver(ctx, r.params, windowsConfig, r.consumer)
	if err != nil {
		return fmt.Errorf("failed to create windows performance counter receiver: %v", err)
	}

	r.windowsReceiver = windowsReceiver
	return r.windowsReceiver.Start(ctx, host)
}

// createWindowsReceiverConfig creates and returns the config needed for the windowsperfcountersreceiver.
func createWindowsReceiverConfig(scraperCfg scraperhelper.ScraperControllerSettings, metricSlice pdata.MetricSlice, metricPerfCounterCfgs map[string][]windowsreceiver.PerfCounterConfig) (*windowsreceiver.Config, error) {
	windowsCfg := &windowsreceiver.Config{
		ScraperControllerSettings: scraperCfg,
	}

	builtMetricCfgs := make(map[string]windowsreceiver.MetricConfig)
	for i := 0; i < metricSlice.Len(); i++ {
		metric := metricSlice.At(i)
		metricCfg := windowsreceiver.MetricConfig{
			Unit:        metric.Unit(),
			Description: metric.Description(),
		}

		if metric.DataType() == pdata.MetricDataTypeSum {
			switch metric.Sum().AggregationTemporality() {
			case pdata.MetricAggregationTemporalityCumulative:
				metricCfg.Sum.Aggregation = "cumulative"
			case pdata.MetricAggregationTemporalityDelta:
				metricCfg.Sum.Aggregation = "delta"
			}
			metricCfg.Sum.Monotonic = metric.Sum().IsMonotonic()
		} else {
			metricCfg.Gauge = windowsreceiver.GaugeMetric{}
		}

		builtMetricCfgs[metric.Name()] = metricCfg
	}

	windowsCfg.MetricMetaData = builtMetricCfgs
	for metricName := range builtMetricCfgs {
		if perfCounter, ok := metricPerfCounterCfgs[metricName]; ok {
			windowsCfg.PerfCounters = append(windowsCfg.PerfCounters, perfCounter...)
		}
	}

	return windowsCfg, nil
}

// Shutdown stops the underlying Windows performance counter receiver.
func (w *sqlServerReceiver) Shutdown(ctx context.Context) error {
	return w.windowsReceiver.Shutdown(ctx)
}

// createMetricSlice makes use of the generated metrics builder to return a MetricsSlice of all enabled metrics.
func createMetricSlice(metricsSettings metadata.MetricsSettings) pdata.MetricSlice {
	metricsBuilder := metadata.NewMetricsBuilder(metricsSettings)
	timestamp := pdata.NewTimestampFromTime(time.Now())
	metricsBuilder.RecordSqlserverBatchRequestRateDataPoint(timestamp, 0)
	metricsBuilder.RecordSqlserverBatchSQLCompilationRateDataPoint(timestamp, 0)
	metricsBuilder.RecordSqlserverBatchSQLRecompilationRateDataPoint(timestamp, 0)
	metricsBuilder.RecordSqlserverLockWaitRateDataPoint(timestamp, 0)
	metricsBuilder.RecordSqlserverLockWaitTimeAvgDataPoint(timestamp, 0)
	metricsBuilder.RecordSqlserverPageBufferCacheHitRatioDataPoint(timestamp, 0)
	metricsBuilder.RecordSqlserverPageCheckpointFlushRateDataPoint(timestamp, 0)
	metricsBuilder.RecordSqlserverPageLazyWriteRateDataPoint(timestamp, 0)
	metricsBuilder.RecordSqlserverPageLifeExpectancyDataPoint(timestamp, 0)
	metricsBuilder.RecordSqlserverPageOperationRateDataPoint(timestamp, 0, "")
	metricsBuilder.RecordSqlserverPageSplitRateDataPoint(timestamp, 0)
	metricsBuilder.RecordSqlserverTransactionLogFlushDataRateDataPoint(timestamp, 0, "")
	metricsBuilder.RecordSqlserverTransactionLogFlushRateDataPoint(timestamp, 0, "")
	metricsBuilder.RecordSqlserverTransactionLogFlushWaitRateDataPoint(timestamp, 0, "")
	metricsBuilder.RecordSqlserverTransactionLogGrowthCountDataPoint(timestamp, 0, "")
	metricsBuilder.RecordSqlserverTransactionLogShrinkCountDataPoint(timestamp, 0, "")
	metricsBuilder.RecordSqlserverTransactionLogUsageDataPoint(timestamp, 0, "")
	metricsBuilder.RecordSqlserverUserConnectionCountDataPoint(timestamp, 0)
	metricSlice := metricsBuilder.Emit().ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()

	return metricSlice
}

// createMetricPerfCounterConfigs creates the map of SQL Server specific PerfCounterConfigs needed for the windowsperfcountersreceiver config.
func createMetricPerfCounterConfigs() map[string][]windowsreceiver.PerfCounterConfig {
	metricPerfCounterConfigs := make(map[string][]windowsreceiver.PerfCounterConfig)

	metricPerfCounterConfigs["sqlserver.user.connection.count"] = []windowsreceiver.PerfCounterConfig{
		{
			Object: "SQLServer:General Statistics",
			Counters: []windowsreceiver.CounterConfig{
				{
					Metric: "sqlserver.user.connection.count",
					Name:   "User Connections",
				},
			},
		},
	}
	metricPerfCounterConfigs["sqlserver.batch.request.rate"] = []windowsreceiver.PerfCounterConfig{
		{
			Object: "SQLServer:SQL Statistics",
			Counters: []windowsreceiver.CounterConfig{
				{
					Metric: "sqlserver.batch.request.rate",
					Name:   "Batch Requests/sec",
				},
			},
		},
	}
	metricPerfCounterConfigs["sqlserver.batch.sql_compilation.rate"] = []windowsreceiver.PerfCounterConfig{
		{
			Object: "SQLServer:SQL Statistics",
			Counters: []windowsreceiver.CounterConfig{
				{
					Metric: "sqlserver.batch.sql_compilation.rate",
					Name:   "SQL Compilations/sec",
				},
			},
		},
	}
	metricPerfCounterConfigs["sqlserver.batch.sql_recompilation.rate"] = []windowsreceiver.PerfCounterConfig{
		{
			Object: "SQLServer:SQL Statistics",
			Counters: []windowsreceiver.CounterConfig{
				{
					Metric: "sqlserver.batch.sql_recompilation.rate",
					Name:   "SQL Re-Compilations/sec",
				},
			},
		},
	}
	metricPerfCounterConfigs["sqlserver.lock.wait.rate"] = []windowsreceiver.PerfCounterConfig{
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
	}
	metricPerfCounterConfigs["sqlserver.lock.wait_time.avg"] = []windowsreceiver.PerfCounterConfig{
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
	}
	metricPerfCounterConfigs["sqlserver.page.buffer_cache.hit_ratio"] = []windowsreceiver.PerfCounterConfig{
		{
			Object: "SQLServer:Buffer Manager",
			Counters: []windowsreceiver.CounterConfig{
				{
					Metric: "sqlserver.page.buffer_cache.hit_ratio",
					Name:   "Buffer cache hit ratio",
				},
			},
		},
	}
	metricPerfCounterConfigs["sqlserver.page.checkpoint.flush.rate"] = []windowsreceiver.PerfCounterConfig{
		{
			Object: "SQLServer:Buffer Manager",
			Counters: []windowsreceiver.CounterConfig{
				{
					Metric: "sqlserver.page.checkpoint.flush.rate",
					Name:   "Checkpoint pages/sec",
				},
			},
		},
	}
	metricPerfCounterConfigs["sqlserver.page.lazy_write.rate"] = []windowsreceiver.PerfCounterConfig{
		{
			Object: "SQLServer:Buffer Manager",
			Counters: []windowsreceiver.CounterConfig{
				{
					Metric: "sqlserver.page.lazy_write.rate",
					Name:   "Lazy Writes/sec",
				},
			},
		},
	}
	metricPerfCounterConfigs["sqlserver.page.life_expectancy"] = []windowsreceiver.PerfCounterConfig{
		{
			Object: "SQLServer:Buffer Manager",
			Counters: []windowsreceiver.CounterConfig{
				{
					Metric: "sqlserver.page.life_expectancy",
					Name:   "Page life expectancy",
				},
			},
		},
	}
	metricPerfCounterConfigs["sqlserver.page.operation.rate"] = []windowsreceiver.PerfCounterConfig{
		{
			Object: "SQLServer:Buffer Manager",
			Counters: []windowsreceiver.CounterConfig{
				{
					Metric:     "sqlserver.page.operation.rate",
					Name:       "Page reads/sec",
					Attributes: map[string]string{"operation": "read"},
				},
				{
					Metric:     "sqlserver.page.operation.rate",
					Name:       "Page writes/sec",
					Attributes: map[string]string{"operation": "write"},
				},
			},
		},
	}
	metricPerfCounterConfigs["sqlserver.page.split.rate"] = []windowsreceiver.PerfCounterConfig{
		{
			Object: "SQLServer:Access Methods",
			Counters: []windowsreceiver.CounterConfig{
				{
					Metric: "sqlserver.page.split.rate",
					Name:   "Page Splits/sec",
				},
			},
		},
	}
	metricPerfCounterConfigs["sqlserver.transaction_log.flush.data.rate"] = []windowsreceiver.PerfCounterConfig{
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
	}
	metricPerfCounterConfigs["sqlserver.transaction_log.flush.rate"] = []windowsreceiver.PerfCounterConfig{
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
	}
	metricPerfCounterConfigs["sqlserver.transaction_log.flush.wait.rate"] = []windowsreceiver.PerfCounterConfig{
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
	}
	metricPerfCounterConfigs["sqlserver.transaction_log.growth.count"] = []windowsreceiver.PerfCounterConfig{
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
	}
	metricPerfCounterConfigs["sqlserver.transaction_log.shrink.count"] = []windowsreceiver.PerfCounterConfig{
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
	}
	metricPerfCounterConfigs["sqlserver.transaction_log.usage"] = []windowsreceiver.PerfCounterConfig{
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
	}

	return metricPerfCounterConfigs
}
