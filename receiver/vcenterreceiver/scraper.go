// Copyright  The OpenTelemetry Authors
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

package vcenterreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vsan/types"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/scrapererror"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver/internal/metadata"
)

var _ component.Receiver = (*vcenterMetricScraper)(nil)

type vcenterMetricScraper struct {
	client *vcenterClient
	config *Config
	mb     *metadata.MetricsBuilder
	logger *zap.Logger
}

func newVmwareVcenterScraper(
	logger *zap.Logger,
	config *Config,
	settings component.ReceiverCreateSettings,
) *vcenterMetricScraper {
	client := newVcenterClient(config)
	return &vcenterMetricScraper{
		client: client,
		config: config,
		logger: logger,
		mb:     metadata.NewMetricsBuilder(config.Metrics, settings.BuildInfo),
	}
}

func (v *vcenterMetricScraper) Start(ctx context.Context, _ component.Host) error {
	connectErr := v.client.EnsureConnection(ctx)
	// don't fail to start if we cannot establish connection, just log an error
	if connectErr != nil {
		v.logger.Error(fmt.Sprintf("unable to establish a connection to the vSphere SDK %s", connectErr.Error()))
	}
	return nil
}

func (v *vcenterMetricScraper) Shutdown(ctx context.Context) error {
	return v.client.Disconnect(ctx)
}

func (v *vcenterMetricScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	if v.client == nil {
		v.client = newVcenterClient(v.config)
	}

	// ensure connection before scraping
	if err := v.client.EnsureConnection(ctx); err != nil {
		return pmetric.NewMetrics(), fmt.Errorf("unable to connect to vSphere SDK: %w", err)
	}
	err := v.collectDatacenters(ctx)
	return v.mb.Emit(), err
}

func (v *vcenterMetricScraper) collectDatacenters(ctx context.Context) error {
	datacenters, err := v.client.Datacenters(ctx)
	if err != nil {
		return err
	}
	errs := &scrapererror.ScrapeErrors{}
	for _, dc := range datacenters {
		v.collectClusters(ctx, dc, errs)
	}
	return errs.Combine()
}

func (v *vcenterMetricScraper) collectClusters(ctx context.Context, datacenter *object.Datacenter, errs *scrapererror.ScrapeErrors) {
	clusters, err := v.client.Clusters(ctx, datacenter)
	if err != nil {
		errs.Add(err)
		return
	}
	now := pcommon.NewTimestampFromTime(time.Now())

	for _, c := range clusters {
		v.collectHosts(ctx, now, c, errs)
		v.collectDatastores(ctx, now, c, errs)
		poweredOnVMs, poweredOffVMs := v.collectVMs(ctx, now, c, errs)
		v.collectCluster(ctx, now, c, poweredOnVMs, poweredOffVMs, errs)
	}
	v.collectResourcePools(ctx, now, errs)
}

func (v *vcenterMetricScraper) collectCluster(
	ctx context.Context,
	now pcommon.Timestamp,
	c *object.ClusterComputeResource,
	poweredOnVMs, poweredOffVMs int64,
	errs *scrapererror.ScrapeErrors,
) {
	v.mb.RecordVcenterClusterVMCountDataPoint(now, poweredOnVMs, metadata.AttributeVMCountPowerStateOn)
	v.mb.RecordVcenterClusterVMCountDataPoint(now, poweredOffVMs, metadata.AttributeVMCountPowerStateOff)

	var moCluster mo.ClusterComputeResource
	err := c.Properties(ctx, c.Reference(), []string{"summary"}, &moCluster)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	s := moCluster.Summary.GetComputeResourceSummary()
	v.mb.RecordVcenterClusterCPULimitDataPoint(now, int64(s.TotalCpu))
	v.mb.RecordVcenterClusterCPUEffectiveDataPoint(now, int64(s.EffectiveCpu))
	v.mb.RecordVcenterClusterMemoryEffectiveDataPoint(now, s.EffectiveMemory)
	v.mb.RecordVcenterClusterMemoryLimitDataPoint(now, s.TotalMemory)
	v.mb.RecordVcenterClusterHostCountDataPoint(now, int64(s.NumHosts-s.NumEffectiveHosts), metadata.AttributeHostEffectiveFalse)
	v.mb.RecordVcenterClusterHostCountDataPoint(now, int64(s.NumEffectiveHosts), metadata.AttributeHostEffectiveTrue)

	mor := c.Reference()
	csvs, err := v.client.VSANCluster(ctx, &mor, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		errs.AddPartial(1, err)
	}
	v.addVSANMetrics(csvs, "*", clusterType, errs)

	v.mb.EmitForResource(
		metadata.WithVcenterClusterName(c.Name()),
	)
}

func (v *vcenterMetricScraper) collectDatastores(
	ctx context.Context,
	colTime pcommon.Timestamp,
	cluster *object.ClusterComputeResource,
	errs *scrapererror.ScrapeErrors,
) {
	datastores, err := cluster.Datastores(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	for _, ds := range datastores {
		v.collectDatastore(ctx, colTime, ds, errs)
	}
}

func (v *vcenterMetricScraper) collectDatastore(
	ctx context.Context,
	now pcommon.Timestamp,
	ds *object.Datastore,
	errs *scrapererror.ScrapeErrors,
) {
	var moDS mo.Datastore
	err := ds.Properties(ctx, ds.Reference(), []string{"summary", "name"}, &moDS)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	v.recordDatastoreProperties(now, moDS)
	v.mb.EmitForResource(
		metadata.WithVcenterDatastoreName(moDS.Name),
	)
}

func (v *vcenterMetricScraper) collectHosts(
	ctx context.Context,
	colTime pcommon.Timestamp,
	cluster *object.ClusterComputeResource,
	errs *scrapererror.ScrapeErrors,
) {
	hosts, err := cluster.Hosts(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	mor := cluster.Reference()
	csvs, err := v.client.VSANHosts(ctx, &mor, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		errs.AddPartial(1, err)
	}

	for _, h := range hosts {
		v.collectHost(ctx, colTime, h, cluster, csvs, errs)
	}
}

func (v *vcenterMetricScraper) collectHost(
	ctx context.Context,
	now pcommon.Timestamp,
	host *object.HostSystem,
	cluster *object.ClusterComputeResource,
	vsanCsvs []types.VsanPerfEntityMetricCSV,
	errs *scrapererror.ScrapeErrors,
) {
	var hwSum mo.HostSystem
	err := host.Properties(ctx, host.Reference(),
		[]string{
			"config",
			"summary.hardware",
			"summary.quickStats",
		}, &hwSum)

	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	v.recordHostSystemMemoryUsage(now, hwSum)
	v.recordHostPerformanceMetrics(ctx, hwSum, errs)
	entityRef := fmt.Sprintf("host-domclient:%v",
		hwSum.Config.VsanHostConfig.ClusterInfo.NodeUuid,
	)
	v.addVSANMetrics(vsanCsvs, entityRef, hostType, errs)
	v.mb.EmitForResource(
		metadata.WithVcenterHostName(host.Name()),
		metadata.WithVcenterClusterName(cluster.Name()),
	)
}

func (v *vcenterMetricScraper) collectResourcePools(
	ctx context.Context,
	ts pcommon.Timestamp,
	errs *scrapererror.ScrapeErrors,
) {
	rps, err := v.client.ResourcePools(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for _, rp := range rps {
		var moRP mo.ResourcePool
		err = rp.Properties(ctx, rp.Reference(), []string{
			"summary",
			"summary.quickStats",
			"name",
		}, &moRP)
		if err != nil {
			errs.AddPartial(1, err)
			continue
		}
		v.recordResourcePool(ts, moRP)
		v.mb.EmitForResource(metadata.WithVcenterResourcePoolName(rp.Name()))
	}
}

func (v *vcenterMetricScraper) collectVMs(
	ctx context.Context,
	colTime pcommon.Timestamp,
	cluster *object.ClusterComputeResource,
	errs *scrapererror.ScrapeErrors,
) (poweredOnVMs int64, poweredOffVMs int64) {
	vms, err := v.client.VMs(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for _, vm := range vms {
		var moVM mo.VirtualMachine
		err := vm.Properties(ctx, vm.Reference(), []string{
			"config",
			"runtime",
			"summary",
		}, &moVM)

		if err != nil {
			errs.AddPartial(1, err)
			continue
		}

		vmUUID := moVM.Config.InstanceUuid

		if string(moVM.Runtime.PowerState) == "poweredOff" {
			poweredOffVMs++
		} else {
			poweredOnVMs++
		}

		host, err := vm.HostSystem(ctx)
		if err != nil {
			errs.AddPartial(1, err)
			return
		}
		hostname, err := host.ObjectName(ctx)
		if err != nil {
			errs.AddPartial(1, err)
			return
		}

		mor := cluster.Reference()
		vsanCsvs, err := v.client.VSANVirtualMachines(ctx, &mor, time.Now().UTC(), time.Now().UTC())
		if err != nil {
			errs.AddPartial(1, err)
		}

		v.collectVM(ctx, colTime, moVM, vsanCsvs, errs)
		v.mb.EmitForResource(
			metadata.WithVcenterVMName(vm.Name()),
			metadata.WithVcenterVMID(vmUUID),
			metadata.WithVcenterClusterName(cluster.Name()),
			metadata.WithVcenterHostName(hostname),
		)
	}
	return poweredOnVMs, poweredOffVMs
}

func (v *vcenterMetricScraper) collectVM(
	ctx context.Context,
	colTime pcommon.Timestamp,
	vm mo.VirtualMachine,
	vsanCsvs []types.VsanPerfEntityMetricCSV,
	errs *scrapererror.ScrapeErrors,
) {
	v.recordVMUsages(colTime, vm)
	v.recordVMPerformance(ctx, vm, errs)

	vmUUID := vm.Config.InstanceUuid
	entityRefID := fmt.Sprintf("virtual-machine:%s", vmUUID)
	v.addVSANMetrics(vsanCsvs, entityRefID, vmType, errs)
}

type vsanType int

const (
	clusterType vsanType = iota
	hostType
	vmType
)

// example 2022-03-10 14:15:00
const timeFormat = "2006-01-02 15:04:05"

func (v *vcenterMetricScraper) addVSANMetrics(
	csvs []types.VsanPerfEntityMetricCSV,
	entityID string,
	vsanType vsanType,
	errs *scrapererror.ScrapeErrors,
) {
	for _, r := range csvs {
		// can't correlate this point to a timestamp so just skip it
		if r.SampleInfo == "" {
			continue
		}
		// If not this entity ID, then skip it
		if vsanType != clusterType && r.EntityRefId != entityID {
			continue
		}

		time, err := time.Parse(timeFormat, r.SampleInfo)
		if err != nil {
			errs.AddPartial(1, err)
			continue
		}

		ts := pcommon.NewTimestampFromTime(time)
		for _, val := range r.Value {
			values := strings.Split(val.Values, ",")
			for _, value := range values {
				switch vsanType {
				case clusterType:
					v.recordClusterVsanMetric(ts, val.MetricId.Label, value, errs)
				case hostType:
					v.recordHostVsanMetric(ts, val.MetricId.Label, value, errs)
				case vmType:
					v.recordVMVsanMetric(ts, val.MetricId.Label, value, errs)
				}
			}
		}
	}
}
