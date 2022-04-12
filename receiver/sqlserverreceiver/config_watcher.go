// Copyright The OpenTelemetry Authors
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

package sqlserverreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver"

import (
	windowsapi "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters"
)

// getWatcherConfigs returns established performance counter configs for each metric.
func getWatcherConfigs() []windowsapi.ObjectConfig {
	return []windowsapi.ObjectConfig{
		{
			Object: "SQLServer:General Statistics",
			Counters: []windowsapi.CounterConfig{
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.user.connection.count",
					},
					Name: "User Connections",
				},
			},
		},
		{
			Object: "SQLServer:SQL Statistics",
			Counters: []windowsapi.CounterConfig{
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.batch.request.rate",
					},

					Name: "Batch Requests/sec",
				},
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.batch.sql_compilation.rate",
					},

					Name: "SQL Compilations/sec",
				},
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.batch.sql_recompilation.rate",
					},
					Name: "SQL Re-Compilations/sec",
				},
			},
		},
		{
			Object:    "SQLServer:Locks",
			Instances: []string{"_Total"},
			Counters: []windowsapi.CounterConfig{
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.lock.wait.rate",
					},
					Name: "Lock Waits/sec",
				},
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.lock.wait_time.avg",
					},
					Name: "Average Wait Time (ms)",
				},
			},
		},
		{
			Object: "SQLServer:Buffer Manager",
			Counters: []windowsapi.CounterConfig{
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.page.buffer_cache.hit_ratio",
					},
					Name: "Buffer cache hit ratio",
				},
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.page.checkpoint.flush.rate",
					},
					Name: "Checkpoint pages/sec",
				},
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.page.lazy_write.rate",
					},
					Name: "Lazy Writes/sec",
				},
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.page.life_expectancy",
					},
					Name: "Page life expectancy",
				},
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.page.operation.rate",
						Attributes: map[string]string{
							"type": "read",
						},
					},
					Name: "Page reads/sec",
				},
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.page.operation.rate",
						Attributes: map[string]string{
							"type": "write",
						},
					},
					Name: "Page writes/sec",
				},
			},
		},
		{
			Object:    "SQLServer:Access Methods",
			Instances: []string{"_Total"},
			Counters: []windowsapi.CounterConfig{
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.page.split.rate",
					},
					Name: "Page Splits/sec",
				},
			},
		},
		{
			Object:    "SQLServer:Databases",
			Instances: []string{"*"},
			Counters: []windowsapi.CounterConfig{
				{
					MetricRep: windowsapi.MetricRep{
						Name: "sqlserver.transaction_log.flush.data.rate",
					},
					Name: "Log Bytes Flushed/sec",
				},
				// 		{
				// 			MetricRep: windowsapi.MetricRep{
				// 				Name: "sqlserver.transaction_log.flush.rate",
				// 			},
				// 			Name: "Log Flushes/sec",
				// 		},
				// 		{
				// 			MetricRep: windowsapi.MetricRep{
				// 				Name: "sqlserver.transaction_log.flush.wait.rate",
				// 			},
				// 			Name: "Log Flush Waits/sec",
				// 		},
				// 		{
				// 			MetricRep: windowsapi.MetricRep{
				// 				Name: "sqlserver.transaction_log.growth.count",
				// 			},
				// 			Name: "Log Growths",
				// 		},
				// 		{
				// 			MetricRep: windowsapi.MetricRep{
				// 				Name: "sqlserver.transaction_log.shrink.count",
				// 			},
				// 			Name: "Log Shrinks",
				// 		},
				// 		{
				// 			MetricRep: windowsapi.MetricRep{
				// 				Name: "sqlserver.transaction_log.usage",
				// 			},
				// 			Name: "Percent Log Used",
				// 		},
				// 		{
				// 			MetricRep: windowsapi.MetricRep{
				// 				Name: "sqlserver.transaction.rate",
				// 			},
				// 			Name: "Transactions/sec",
				// 		},
				// 		{
				// 			MetricRep: windowsapi.MetricRep{
				// 				Name: "sqlserver.transaction.write.rate",
				// 			},
				// 			Name: "Write Transactions/sec",
				// 		},
			},
		},
	}
}
