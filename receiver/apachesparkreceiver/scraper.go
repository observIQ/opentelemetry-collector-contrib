// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apachesparkreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver"

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scrapererror"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver/internal/models"
)

var (
	errClientNotInit         = errors.New("client not initialized")
	errFailedAppIDCollection = errors.New("failed to retrieve app ids")
)

type sparkScraper struct {
	client   client
	logger   *zap.Logger
	config   *Config
	settings component.TelemetrySettings
	mb       *metadata.MetricsBuilder
}

func newSparkScraper(logger *zap.Logger, cfg *Config, settings receiver.CreateSettings) *sparkScraper {
	return &sparkScraper{
		logger:   logger,
		config:   cfg,
		settings: settings.TelemetrySettings,
		mb:       metadata.NewMetricsBuilder(cfg.MetricsBuilderConfig, settings),
	}
}

func (s *sparkScraper) start(_ context.Context, host component.Host) (err error) {
	httpClient, err := newApacheSparkClient(s.config, host, s.settings)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}
	s.client = httpClient
	return nil
}

func (s *sparkScraper) scrape(_ context.Context) (pmetric.Metrics, error) {
	now := pcommon.NewTimestampFromTime(time.Now())
	var scrapeErrors scrapererror.ScrapeErrors

	if s.client == nil {
		return pmetric.NewMetrics(), errClientNotInit
	}

	// call applications endpoint
	// not getting app name for now, just ids
	var apps *models.Applications
	apps, err := s.client.GetApplications()
	if err != nil {
		return pmetric.NewMetrics(), errFailedAppIDCollection
	}

	// get stats from the 'metrics' endpoint
	clusterStats, err := s.client.GetClusterStats()
	if err != nil {
		scrapeErrors.AddPartial(1, err)
		s.logger.Warn("Failed to scrape cluster stats", zap.Error(err))
	} else {
		for _, app := range *apps {
			s.collectCluster(clusterStats, now, app.ApplicationID, app.Name)
		}
	}

	// for each application id, get stats from stages & executors endpoints
	for _, app := range *apps {

		stageStats, err := s.client.GetStageStats(app.ApplicationID)
		if err != nil {
			scrapeErrors.AddPartial(1, err)
			s.logger.Warn("Failed to scrape stage stats", zap.Error(err))
		} else {
			s.collectStage(*stageStats, now, app.ApplicationID, app.Name)
		}

		executorStats, err := s.client.GetExecutorStats(app.ApplicationID)
		if err != nil {
			scrapeErrors.AddPartial(1, err)
			s.logger.Warn("Failed to scrape executor stats", zap.Error(err))
		} else {
			s.collectExecutor(*executorStats, now, app.ApplicationID, app.Name)
		}

		jobStats, err := s.client.GetJobStats(app.ApplicationID)
		if err != nil {
			scrapeErrors.AddPartial(1, err)
			s.logger.Warn("Failed to scrape job stats", zap.Error(err))
		} else {
			s.collectJob(*jobStats, now, app.ApplicationID, app.Name)
		}
	}
	return s.mb.Emit(), scrapeErrors.Combine()
}

func (s *sparkScraper) collectCluster(clusterStats *models.ClusterProperties, now pcommon.Timestamp, appID string, appName string) {
	key := fmt.Sprintf("%s.driver.BlockManager.disk.diskSpaceUsed", appID)
	s.mb.RecordSparkDriverBlockManagerDiskDiskSpaceUsedDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName)

	key = fmt.Sprintf("%s.driver.BlockManager.memory.offHeapMemUsed_MB", appID)
	s.mb.RecordSparkDriverBlockManagerMemoryUsedDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeLocationOffHeap)
	key = fmt.Sprintf("%s.driver.BlockManager.memory.onHeapMemUsed_MB", appID)
	s.mb.RecordSparkDriverBlockManagerMemoryUsedDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeLocationOnHeap)

	key = fmt.Sprintf("%s.driver.BlockManager.memory.remainingOffHeapMem_MB", appID)
	s.mb.RecordSparkDriverBlockManagerMemoryRemainingDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeLocationOffHeap)
	key = fmt.Sprintf("%s.driver.BlockManager.memory.remainingOnHeapMem_MB", appID)
	s.mb.RecordSparkDriverBlockManagerMemoryRemainingDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeLocationOnHeap)

	key = fmt.Sprintf("%s.driver.HiveExternalCatalog.fileCacheHits", appID)
	s.mb.RecordSparkDriverHiveExternalCatalogFileCacheHitsDataPoint(now, int64(clusterStats.Counters[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.HiveExternalCatalog.filesDiscovered", appID)
	s.mb.RecordSparkDriverHiveExternalCatalogFilesDiscoveredDataPoint(now, int64(clusterStats.Counters[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.HiveExternalCatalog.hiveClientCalls", appID)
	s.mb.RecordSparkDriverHiveExternalCatalogHiveClientCallsDataPoint(now, int64(clusterStats.Counters[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.HiveExternalCatalog.parallelListingJobCount", appID)
	s.mb.RecordSparkDriverHiveExternalCatalogParallelListingJobsDataPoint(now, int64(clusterStats.Counters[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.HiveExternalCatalog.partitionsFetched", appID)
	s.mb.RecordSparkDriverHiveExternalCatalogPartitionsFetchedDataPoint(now, int64(clusterStats.Counters[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.CodeGenerator.compilationTime", appID)
	s.mb.RecordSparkDriverCodeGeneratorCompilationCountDataPoint(now, int64(clusterStats.Histograms[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.CodeGenerator.compilationTime", appID)
	s.mb.RecordSparkDriverCodeGeneratorCompilationAverageTimeDataPoint(now, clusterStats.Histograms[key].Mean, appID, appName)

	key = fmt.Sprintf("%s.driver.CodeGenerator.generatedClassSize", appID)
	s.mb.RecordSparkDriverCodeGeneratorGeneratedClassCountDataPoint(now, int64(clusterStats.Histograms[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.CodeGenerator.generatedClassSize", appID)
	s.mb.RecordSparkDriverCodeGeneratorGeneratedClassAverageSizeDataPoint(now, clusterStats.Histograms[key].Mean, appID, appName)

	key = fmt.Sprintf("%s.driver.CodeGenerator.generatedMethodSize", appID)
	s.mb.RecordSparkDriverCodeGeneratorGeneratedMethodCountDataPoint(now, int64(clusterStats.Histograms[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.CodeGenerator.generatedMethodSize", appID)
	s.mb.RecordSparkDriverCodeGeneratorGeneratedMethodAverageSizeDataPoint(now, clusterStats.Histograms[key].Mean, appID, appName)

	key = fmt.Sprintf("%s.driver.CodeGenerator.sourceCodeSize", appID)
	s.mb.RecordSparkDriverCodeGeneratorSourceCodeCountDataPoint(now, int64(clusterStats.Histograms[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.CodeGenerator.sourceCodeSize", appID)
	s.mb.RecordSparkDriverCodeGeneratorSourceCodeAverageSizeDataPoint(now, clusterStats.Histograms[key].Mean, appID, appName)

	key = fmt.Sprintf("%s.driver.DAGScheduler.job.activeJobs", appID)
	s.mb.RecordSparkDriverDagSchedulerJobActiveJobsDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName)

	key = fmt.Sprintf("%s.driver.DAGScheduler.job.allJobs", appID)
	s.mb.RecordSparkDriverDagSchedulerJobAllJobsDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName)

	key = fmt.Sprintf("%s.driver.DAGScheduler.stage.failedStages", appID)
	s.mb.RecordSparkDriverDagSchedulerStageFailedStagesDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName)

	key = fmt.Sprintf("%s.driver.DAGScheduler.stage.runningStages", appID)
	s.mb.RecordSparkDriverDagSchedulerStageRunningStagesDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName)

	key = fmt.Sprintf("%s.driver.DAGScheduler.stage.waitingStages", appID)
	s.mb.RecordSparkDriverDagSchedulerStageWaitingStagesDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName)

	key = fmt.Sprintf("%s.driver.LiveListenerBus.numEventsPosted", appID)
	s.mb.RecordSparkDriverLiveListenerBusEventsPostedDataPoint(now, int64(clusterStats.Counters[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.LiveListenerBus.queue.appStatus.listenerProcessingTime", appID)
	s.mb.RecordSparkDriverLiveListenerBusListenerProcessingTimeAverageDataPoint(now, clusterStats.Histograms[key].Mean, appID, appName)

	key = fmt.Sprintf("%s.driver.LiveListenerBus.queue.appStatus.numDroppedEvents", appID)
	s.mb.RecordSparkDriverLiveListenerBusEventsDroppedDataPoint(now, int64(clusterStats.Counters[key].Count), appID, appName)

	key = fmt.Sprintf("%s.driver.LiveListenerBus.queue.appStatus.size", appID)
	s.mb.RecordSparkDriverLiveListenerBusQueueSizeDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName)

	key = fmt.Sprintf("%s.driver.JVMCPU.jvmCpuTime", appID)
	s.mb.RecordSparkDriverJvmCPUTimeDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName)

	key = fmt.Sprintf("%s.driver.ExecutorMetrics.JVMOffHeapMemory", appID)
	s.mb.RecordSparkDriverExecutorMetricsJvmMemoryDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeLocationOffHeap)
	key = fmt.Sprintf("%s.driver.ExecutorMetrics.JVMHeapMemory", appID)
	s.mb.RecordSparkDriverExecutorMetricsJvmMemoryDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeLocationOnHeap)

	key = fmt.Sprintf("%s.driver.ExecutorMetrics.OffHeapExecutionMemory", appID)
	s.mb.RecordSparkDriverExecutorMetricsExecutionMemoryDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeLocationOffHeap)
	key = fmt.Sprintf("%s.driver.ExecutorMetrics.OnHeapExecutionMemory", appID)
	s.mb.RecordSparkDriverExecutorMetricsExecutionMemoryDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeLocationOnHeap)

	key = fmt.Sprintf("%s.driver.ExecutorMetrics.OffHeapStorageMemory", appID)
	s.mb.RecordSparkDriverExecutorMetricsStorageMemoryDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeLocationOffHeap)
	key = fmt.Sprintf("%s.driver.ExecutorMetrics.OnHeapStorageMemory", appID)
	s.mb.RecordSparkDriverExecutorMetricsStorageMemoryDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeLocationOnHeap)

	key = fmt.Sprintf("%s.driver.ExecutorMetrics.DirectPoolMemory", appID)
	s.mb.RecordSparkDriverExecutorMetricsPoolMemoryDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributePoolMemoryTypeDirect)
	key = fmt.Sprintf("%s.driver.ExecutorMetrics.MappedPoolMemory", appID)
	s.mb.RecordSparkDriverExecutorMetricsPoolMemoryDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributePoolMemoryTypeMapped)

	key = fmt.Sprintf("%s.driver.ExecutorMetrics.MinorGCCount", appID)
	s.mb.RecordSparkDriverExecutorMetricsGcCountDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeGcTypeMinor)
	key = fmt.Sprintf("%s.driver.ExecutorMetrics.MajorGCCount", appID)
	s.mb.RecordSparkDriverExecutorMetricsGcCountDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeGcTypeMajor)

	key = fmt.Sprintf("%s.driver.ExecutorMetrics.MinorGCTime", appID)
	s.mb.RecordSparkDriverExecutorMetricsGcTimeDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeGcTypeMinor)
	key = fmt.Sprintf("%s.driver.ExecutorMetrics.MajorGCTime", appID)
	s.mb.RecordSparkDriverExecutorMetricsGcTimeDataPoint(now, int64(clusterStats.Gauges[key].Value), appID, appName, metadata.AttributeGcTypeMajor)
}

func (s *sparkScraper) collectStage(stageStats models.Stages, now pcommon.Timestamp, appID string, appName string) {
	for i := range stageStats {
		var stageStatus metadata.AttributeStageStatus
		switch stageStats[i].Status {
		case "ACTIVE":
			stageStatus = metadata.AttributeStageStatusACTIVE
		case "COMPLETE":
			stageStatus = metadata.AttributeStageStatusCOMPLETE
		case "PENDING":
			stageStatus = metadata.AttributeStageStatusPENDING
		case "FAILED":
			stageStatus = metadata.AttributeStageStatusFAILED
		}

		s.mb.RecordSparkStageActiveTasksDataPoint(now, int64(stageStats[i].ExecutorRunTime), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageCompleteTasksDataPoint(now, int64(stageStats[i].ExecutorRunTime), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageFailedTasksDataPoint(now, int64(stageStats[i].ExecutorRunTime), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageKilledTasksDataPoint(now, int64(stageStats[i].ExecutorRunTime), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageExecutorRunTimeDataPoint(now, int64(stageStats[i].ExecutorRunTime), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageExecutorCPUTimeDataPoint(now, int64(stageStats[i].ExecutorCPUTime), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageResultSizeDataPoint(now, int64(stageStats[i].ResultSize), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageJvmGcTimeDataPoint(now, int64(stageStats[i].JvmGcTime), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageMemorySpilledDataPoint(now, int64(stageStats[i].MemoryBytesSpilled), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageDiskSpaceSpilledDataPoint(now, int64(stageStats[i].DiskBytesSpilled), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStagePeakExecutionMemoryDataPoint(now, int64(stageStats[i].PeakExecutionMemory), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageInputBytesDataPoint(now, int64(stageStats[i].InputBytes), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageInputRecordsDataPoint(now, int64(stageStats[i].InputRecords), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageOutputBytesDataPoint(now, int64(stageStats[i].OutputBytes), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageOutputRecordsDataPoint(now, int64(stageStats[i].OutputRecords), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageShuffleBlocksFetchedDataPoint(now, int64(stageStats[i].ShuffleRemoteBlocksFetched), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus, metadata.AttributeSourceRemote)
		s.mb.RecordSparkStageShuffleBlocksFetchedDataPoint(now, int64(stageStats[i].ShuffleLocalBlocksFetched), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus, metadata.AttributeSourceLocal)
		s.mb.RecordSparkStageShuffleFetchWaitTimeDataPoint(now, int64(stageStats[i].ShuffleFetchWaitTime), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageShuffleBytesReadDataPoint(now, int64(stageStats[i].ShuffleRemoteBytesRead), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus, metadata.AttributeSourceRemote)
		s.mb.RecordSparkStageShuffleBytesReadDataPoint(now, int64(stageStats[i].ShuffleLocalBytesRead), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus, metadata.AttributeSourceLocal)
		s.mb.RecordSparkStageShuffleRemoteBytesReadToDiskDataPoint(now, int64(stageStats[i].ShuffleRemoteBytesReadToDisk), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageShuffleReadBytesDataPoint(now, int64(stageStats[i].ShuffleReadBytes), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageShuffleReadRecordsDataPoint(now, int64(stageStats[i].ShuffleReadRecords), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageShuffleWriteBytesDataPoint(now, int64(stageStats[i].ShuffleWriteBytes), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageShuffleWriteRecordsDataPoint(now, int64(stageStats[i].ShuffleWriteRecords), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
		s.mb.RecordSparkStageShuffleWriteTimeDataPoint(now, int64(stageStats[i].ShuffleWriteTime), appID, appName, stageStats[i].StageID, stageStats[i].AttemptID, stageStatus)
	}
}

func (s *sparkScraper) collectExecutor(executorStats models.Executors, now pcommon.Timestamp, appID string, appName string) {
	for i := range executorStats {
		s.mb.RecordSparkExecutorMemoryUsedDataPoint(now, executorStats[i].MemoryUsed, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorDiskUsedDataPoint(now, executorStats[i].DiskUsed, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorMaxTasksDataPoint(now, executorStats[i].MaxTasks, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorActiveTasksDataPoint(now, executorStats[i].ActiveTasks, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorFailedTasksDataPoint(now, executorStats[i].FailedTasks, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorCompletedTasksDataPoint(now, executorStats[i].CompletedTasks, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorDurationDataPoint(now, executorStats[i].TotalDuration, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorGcTimeDataPoint(now, executorStats[i].TotalGCTime, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorInputBytesDataPoint(now, executorStats[i].TotalInputBytes, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorShuffleReadBytesDataPoint(now, executorStats[i].TotalShuffleRead, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorShuffleWriteBytesDataPoint(now, executorStats[i].TotalShuffleWrite, appID, appName, executorStats[i].ExecutorID)
		s.mb.RecordSparkExecutorUsedStorageMemoryDataPoint(now, executorStats[i].UsedOnHeapStorageMemory, appID, appName, executorStats[i].ExecutorID, metadata.AttributeLocationOnHeap)
		s.mb.RecordSparkExecutorUsedStorageMemoryDataPoint(now, executorStats[i].UsedOffHeapStorageMemory, appID, appName, executorStats[i].ExecutorID, metadata.AttributeLocationOffHeap)
		s.mb.RecordSparkExecutorTotalStorageMemoryDataPoint(now, executorStats[i].TotalOnHeapStorageMemory, appID, appName, executorStats[i].ExecutorID, metadata.AttributeLocationOnHeap)
		s.mb.RecordSparkExecutorTotalStorageMemoryDataPoint(now, executorStats[i].TotalOffHeapStorageMemory, appID, appName, executorStats[i].ExecutorID, metadata.AttributeLocationOffHeap)
	}
}

func (s *sparkScraper) collectJob(jobStats models.Jobs, now pcommon.Timestamp, appID string, appName string) {
	for i := range jobStats {
		s.mb.RecordSparkJobActiveTasksDataPoint(now, jobStats[i].NumActiveTasks, appID, appName, jobStats[i].JobID)
		s.mb.RecordSparkJobCompletedTasksDataPoint(now, jobStats[i].NumCompletedTasks, appID, appName, jobStats[i].JobID)
		s.mb.RecordSparkJobSkippedTasksDataPoint(now, jobStats[i].NumSkippedTasks, appID, appName, jobStats[i].JobID)
		s.mb.RecordSparkJobFailedTasksDataPoint(now, jobStats[i].NumFailedTasks, appID, appName, jobStats[i].JobID)
		s.mb.RecordSparkJobActiveStagesDataPoint(now, jobStats[i].NumActiveStages, appID, appName, jobStats[i].JobID)
		s.mb.RecordSparkJobCompletedStagesDataPoint(now, jobStats[i].NumCompletedStages, appID, appName, jobStats[i].JobID)
		s.mb.RecordSparkJobSkippedStagesDataPoint(now, jobStats[i].NumSkippedStages, appID, appName, jobStats[i].JobID)
		s.mb.RecordSparkJobFailedStagesDataPoint(now, jobStats[i].NumFailedStages, appID, appName, jobStats[i].JobID)
	}
}
