// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package vcenterreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"

import (
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/mo"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver/internal/metadata"
)

func (v *vcenterMetricScraper) recordClusterStats(
	colTime pcommon.Timestamp,
	cr *mo.ComputeResource,
	poweredOnVMs, poweredOffVMs int64,
) {
	v.mb.RecordVcenterClusterVMCountDataPoint(colTime, poweredOnVMs, metadata.AttributeVMCountPowerStateOn)
	v.mb.RecordVcenterClusterVMCountDataPoint(colTime, poweredOffVMs, metadata.AttributeVMCountPowerStateOff)

	s := cr.Summary.GetComputeResourceSummary()
	v.mb.RecordVcenterClusterCPULimitDataPoint(colTime, int64(s.TotalCpu))
	v.mb.RecordVcenterClusterCPUEffectiveDataPoint(colTime, int64(s.EffectiveCpu))
	v.mb.RecordVcenterClusterMemoryEffectiveDataPoint(colTime, s.EffectiveMemory)
	v.mb.RecordVcenterClusterMemoryLimitDataPoint(colTime, s.TotalMemory)
	v.mb.RecordVcenterClusterHostCountDataPoint(colTime, int64(s.NumHosts-s.NumEffectiveHosts), false)
	v.mb.RecordVcenterClusterHostCountDataPoint(colTime, int64(s.NumEffectiveHosts), true)
}

func (v *vcenterMetricScraper) recordHostSystemStats(
	colTime pcommon.Timestamp,
	hs *mo.HostSystem,
) {
	s := hs.Summary
	h := s.Hardware
	z := s.QuickStats

	v.mb.RecordVcenterHostMemoryUsageDataPoint(colTime, int64(z.OverallMemoryUsage))
	memUtilization := 100 * float64(z.OverallMemoryUsage) / float64(h.MemorySize>>20)
	v.mb.RecordVcenterHostMemoryUtilizationDataPoint(colTime, memUtilization)

	v.mb.RecordVcenterHostCPUUsageDataPoint(colTime, int64(z.OverallCpuUsage))
	cpuUtilization := 100 * float64(z.OverallCpuUsage) / float64(int32(h.NumCpuCores)*h.CpuMhz)
	v.mb.RecordVcenterHostCPUUtilizationDataPoint(colTime, cpuUtilization)
}

func (v *vcenterMetricScraper) recordVMStats(
	colTime pcommon.Timestamp,
	vm *mo.VirtualMachine,
	hs *mo.HostSystem,
) {
	memUsage := vm.Summary.QuickStats.GuestMemoryUsage
	balloonedMem := vm.Summary.QuickStats.BalloonedMemory
	swappedMem := vm.Summary.QuickStats.SwappedMemory
	swappedSSDMem := vm.Summary.QuickStats.SsdSwappedMemory

	if totalMemory := vm.Summary.Config.MemorySizeMB; totalMemory > 0 && memUsage > 0 {
		memoryUtilization := float64(memUsage) / float64(totalMemory) * 100
		v.mb.RecordVcenterVMMemoryUtilizationDataPoint(colTime, memoryUtilization)
	}

	v.mb.RecordVcenterVMMemoryUsageDataPoint(colTime, int64(memUsage))
	v.mb.RecordVcenterVMMemoryBalloonedDataPoint(colTime, int64(balloonedMem))
	v.mb.RecordVcenterVMMemorySwappedDataPoint(colTime, int64(swappedMem))
	v.mb.RecordVcenterVMMemorySwappedSsdDataPoint(colTime, swappedSSDMem)

	diskUsed := vm.Summary.Storage.Committed
	diskFree := vm.Summary.Storage.Uncommitted

	v.mb.RecordVcenterVMDiskUsageDataPoint(colTime, diskUsed, metadata.AttributeDiskStateUsed)
	v.mb.RecordVcenterVMDiskUsageDataPoint(colTime, diskFree, metadata.AttributeDiskStateAvailable)
	if diskFree != 0 {
		diskUtilization := float64(diskUsed) / float64(diskFree+diskUsed) * 100
		v.mb.RecordVcenterVMDiskUtilizationDataPoint(colTime, diskUtilization)
	}

	cpuUsage := vm.Summary.QuickStats.OverallCpuUsage
	if cpuUsage == 0 {
		// Most likely the VM is unavailable or is unreachable.
		return
	}
	v.mb.RecordVcenterVMCPUUsageDataPoint(colTime, int64(cpuUsage))

	// https://communities.vmware.com/t5/VMware-code-Documents/Resource-Management/ta-p/2783456
	// VirtualMachine.runtime.maxCpuUsage is a property of the virtual machine, indicating the limit value.
	// This value is always equal to the limit value set for that virtual machine.
	// If no limit, it has full host mhz * vm.Config.Hardware.NumCPU.
	cpuLimit := vm.Config.Hardware.NumCPU * hs.Summary.Hardware.CpuMhz
	if vm.Runtime.MaxCpuUsage != 0 {
		cpuLimit = vm.Runtime.MaxCpuUsage
	}
	if cpuLimit == 0 {
		// This shouldn't happen, but protect against division by zero.
		return
	}
	v.mb.RecordVcenterVMCPUUtilizationDataPoint(colTime, 100*float64(cpuUsage)/float64(cpuLimit))
}

func (v *vcenterMetricScraper) recordDatastoreStats(
	colTime pcommon.Timestamp,
	ds *mo.Datastore,
) {
	s := ds.Summary
	diskUsage := s.Capacity - s.FreeSpace
	diskUtilization := float64(diskUsage) / float64(s.Capacity) * 100
	v.mb.RecordVcenterDatastoreDiskUsageDataPoint(colTime, diskUsage, metadata.AttributeDiskStateUsed)
	v.mb.RecordVcenterDatastoreDiskUsageDataPoint(colTime, s.FreeSpace, metadata.AttributeDiskStateAvailable)
	v.mb.RecordVcenterDatastoreDiskUtilizationDataPoint(colTime, diskUtilization)
}

func (v *vcenterMetricScraper) recordResourcePoolStats(
	colTime pcommon.Timestamp,
	rp *mo.ResourcePool,
) {
	s := rp.Summary.GetResourcePoolSummary()
	if s.QuickStats != nil {
		v.mb.RecordVcenterResourcePoolCPUUsageDataPoint(colTime, s.QuickStats.OverallCpuUsage)
		v.mb.RecordVcenterResourcePoolMemoryUsageDataPoint(colTime, s.QuickStats.GuestMemoryUsage)
	}

	v.mb.RecordVcenterResourcePoolCPUSharesDataPoint(colTime, int64(s.Config.CpuAllocation.Shares.Shares))
	v.mb.RecordVcenterResourcePoolMemorySharesDataPoint(colTime, int64(s.Config.MemoryAllocation.Shares.Shares))

}

var hostPerfMetricList = []string{
	// network metrics
	"net.bytesTx.average",
	"net.bytesRx.average",
	"net.packetsTx.summation",
	"net.packetsRx.summation",
	"net.usage.average",
	"net.errorsRx.summation",
	"net.errorsTx.summation",

	// disk metrics
	"virtualDisk.totalWriteLatency.average",
	"disk.deviceReadLatency.average",
	"disk.deviceWriteLatency.average",
	"disk.kernelReadLatency.average",
	"disk.kernelWriteLatency.average",
	"disk.maxTotalLatency.latest",
	"disk.read.average",
	"disk.write.average",
}

// vmPerfMetricList may be customizable in the future but here is the full list of Virtual Machine Performance Counters
// https://docs.vmware.com/en/vRealize-Operations/8.6/com.vmware.vcom.metrics.doc/GUID-1322F5A4-DA1D-481F-BBEA-99B228E96AF2.html
var vmPerfMetricList = []string{
	// network metrics
	"net.packetsTx.summation",
	"net.packetsRx.summation",
	"net.bytesRx.average",
	"net.bytesTx.average",
	"net.usage.average",

	// disk metrics
	"disk.totalWriteLatency.average",
	"disk.totalReadLatency.average",
	"disk.maxTotalLatency.latest",
	"virtualDisk.totalWriteLatency.average",
	"virtualDisk.totalReadLatency.average",
}

func (v *vcenterMetricScraper) recordVMPerformanceMetrics(entityMetric *performance.EntityMetric) {
	for _, val := range entityMetric.Value {
		for j, nestedValue := range val.Value {
			si := entityMetric.SampleInfo[j]
			switch val.Name {
			// Performance monitoring level 1 metrics
			case "net.bytesTx.average":
				v.mb.RecordVcenterVMNetworkThroughputDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeThroughputDirectionTransmitted, val.Instance)
			case "net.bytesRx.average":
				v.mb.RecordVcenterVMNetworkThroughputDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeThroughputDirectionReceived, val.Instance)
			case "net.usage.average":
				v.mb.RecordVcenterVMNetworkUsageDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, val.Instance)
			case "net.packetsTx.summation":
				v.mb.RecordVcenterVMNetworkPacketCountDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeThroughputDirectionTransmitted, val.Instance)
			case "net.packetsRx.summation":
				v.mb.RecordVcenterVMNetworkPacketCountDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeThroughputDirectionReceived, val.Instance)

			// Performance monitoring level 2 metrics required
			case "disk.totalReadLatency.average":
				v.mb.RecordVcenterVMDiskLatencyAvgDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeDiskDirectionRead, metadata.AttributeDiskTypePhysical, val.Instance)
			case "virtualDisk.totalReadLatency.average":
				v.mb.RecordVcenterVMDiskLatencyAvgDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeDiskDirectionRead, metadata.AttributeDiskTypeVirtual, val.Instance)
			case "disk.totalWriteLatency.average":
				v.mb.RecordVcenterVMDiskLatencyAvgDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeDiskDirectionWrite, metadata.AttributeDiskTypePhysical, val.Instance)
			case "virtualDisk.totalWriteLatency.average":
				v.mb.RecordVcenterVMDiskLatencyAvgDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeDiskDirectionWrite, metadata.AttributeDiskTypeVirtual, val.Instance)
			case "disk.maxTotalLatency.latest":
				v.mb.RecordVcenterVMDiskLatencyMaxDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, val.Instance)
			}
		}
	}
}

func (v *vcenterMetricScraper) recordHostPerformanceMetrics(entityMetric *performance.EntityMetric) {
	for _, val := range entityMetric.Value {
		for j, nestedValue := range val.Value {
			si := entityMetric.SampleInfo[j]
			switch val.Name {
			// Performance monitoring level 1 metrics
			case "net.usage.average":
				v.mb.RecordVcenterHostNetworkUsageDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, val.Instance)
			case "net.bytesTx.average":
				v.mb.RecordVcenterHostNetworkThroughputDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeThroughputDirectionTransmitted, val.Instance)
			case "net.bytesRx.average":
				v.mb.RecordVcenterHostNetworkThroughputDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeThroughputDirectionReceived, val.Instance)
			case "net.packetsTx.summation":
				v.mb.RecordVcenterHostNetworkPacketCountDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeThroughputDirectionTransmitted, val.Instance)
			case "net.packetsRx.summation":
				v.mb.RecordVcenterHostNetworkPacketCountDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeThroughputDirectionReceived, val.Instance)

			// Following requires performance level 2
			case "net.errorsRx.summation":
				v.mb.RecordVcenterHostNetworkPacketErrorsDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeThroughputDirectionReceived, val.Instance)
			case "net.errorsTx.summation":
				v.mb.RecordVcenterHostNetworkPacketErrorsDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeThroughputDirectionTransmitted, val.Instance)
			case "disk.totalWriteLatency.average":
				v.mb.RecordVcenterHostDiskLatencyAvgDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeDiskDirectionWrite, val.Instance)
			case "disk.totalReadLatency.average":
				v.mb.RecordVcenterHostDiskLatencyAvgDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeDiskDirectionRead, val.Instance)
			case "disk.maxTotalLatency.latest":
				v.mb.RecordVcenterHostDiskLatencyMaxDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, val.Instance)

			// Following requires performance level 4
			case "disk.read.average":
				v.mb.RecordVcenterHostDiskThroughputDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeDiskDirectionRead, val.Instance)
			case "disk.write.average":
				v.mb.RecordVcenterHostDiskThroughputDataPoint(pcommon.NewTimestampFromTime(si.Timestamp), nestedValue, metadata.AttributeDiskDirectionWrite, val.Instance)
			}
		}
	}
}
