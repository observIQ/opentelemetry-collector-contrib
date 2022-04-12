package metadata

import (
	"go.opentelemetry.io/collector/model/pdata"
)

func (mb *MetricsBuilder) RecordAnyDataPoint(ts pdata.Timestamp, val float64, name string, attributes map[string]string) {
	switch name {
	case "sqlserver.user.connection.count":
		mb.RecordSqlserverUserConnectionCountDataPoint(ts, val)
	case "sqlserver.batch.request.rate":
		mb.RecordSqlserverBatchRequestRateDataPoint(ts, val)
	case "sqlserver.batch.sql_compilation.rate":
		mb.RecordSqlserverBatchSQLCompilationRateDataPoint(ts, val)
	case "sqlserver.batch.sql_recompilation.rate":
		mb.RecordSqlserverBatchSQLRecompilationRateDataPoint(ts, val)
	case "sqlserver.lock.wait.rate":
		mb.RecordSqlserverLockWaitRateDataPoint(ts, val)
	case "sqlserver.lock.wait_time.avg":
		mb.RecordSqlserverLockWaitTimeAvgDataPoint(ts, val)
	case "sqlserver.page.buffer_cache.hit_ratio":
		mb.RecordSqlserverPageBufferCacheHitRatioDataPoint(ts, val)
	case "sqlserver.page.checkpoint.flush.rate":
		mb.RecordSqlserverPageCheckpointFlushRateDataPoint(ts, val)
	case "sqlserver.page.lazy_write.rate":
		mb.RecordSqlserverPageLazyWriteRateDataPoint(ts, val)
	case "sqlserver.page.life_expectancy":
		mb.RecordSqlserverPageLifeExpectancyDataPoint(ts, val)
	case "sqlserver.page.operation.rate":
		mb.RecordSqlserverPageOperationRateDataPoint(ts, val, attributes["type"])
	case "sqlserver.page.split.rate":
		mb.RecordSqlserverPageSplitRateDataPoint(ts, val)
	case "sqlserver.transaction_log.flush.data.rate":
		mb.RecordSqlserverTransactionLogFlushDataRateDataPoint(ts, val, attributes["database"])
		// case "sqlserver.transaction_log.flush.rate":
		// 	mb.RecordSqlserverTransactionLogFlushRateDataPoint(ts, val, attributes["database"])
		// case "sqlserver.transaction_log.flush.wait.rate":
		// 	mb.RecordSqlserverTransactionLogFlushWaitRateDataPoint(ts, val, attributes["database"])
		// case "sqlserver.transaction_log.growth.count":
		// 	mb.RecordSqlserverTransactionLogGrowthCountDataPoint(ts, val, attributes["database"])
		// case "sqlserver.transaction_log.shrink.count":
		// 	mb.RecordSqlserverTransactionLogShrinkCountDataPoint(ts, val, attributes["database"])
		// case "sqlserver.transaction_log.usage":
		// 	mb.RecordSqlserverTransactionLogUsageDataPoint(ts, val, attributes["database"])
		// case "sqlserver.transaction.rate":
		// 	mb.RecordSqlserverTransactionRateDataPoint(ts, val, attributes["database"])
		// case "sqlserver.transaction.write.rate":
		// 	mb.RecordSqlserverTransactionWriteRateDataPoint(ts, val, attributes["database"])
	}
}
