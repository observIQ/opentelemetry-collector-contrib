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
	"time"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
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
) *vcenterMetricScraper {
	client := newVmwarevcenterClient(config)
	return &vcenterMetricScraper{
		client: client,
		config: config,
		logger: logger,
		mb:     metadata.NewMetricsBuilder(config.MetricsConfig.Settings),
	}
}

func (v *vcenterMetricScraper) Start(ctx context.Context, _ component.Host) error {
	return v.client.EnsureConnection(ctx)
}

func (v *vcenterMetricScraper) Shutdown(ctx context.Context) error {
	return v.client.Disconnect(ctx)
}

func (v *vcenterMetricScraper) scrape(ctx context.Context) (pdata.Metrics, error) {
	if v.client == nil {
		v.client = newVmwarevcenterClient(v.config)
	}

	// ensure connection before scraping
	if err := v.client.EnsureConnection(ctx); err != nil {
		return pdata.NewMetrics(), fmt.Errorf("unable to connect to vSphere SDK: %w", err)
	}

	err := v.collectClusters(ctx)
	return v.mb.Emit(), err
}

func (v *vcenterMetricScraper) collectClusters(ctx context.Context) error {
	errs := &scrapererror.ScrapeErrors{}

	clusters, err := v.client.Clusters(ctx)
	if err != nil {
		errs.Add(err)
		return errs.Combine()
	}
	now := pdata.NewTimestampFromTime(time.Now())

	for _, c := range clusters {
		v.collectHosts(ctx, now, c, errs)
		v.collectDatastores(ctx, now, c, errs)
		v.collectVMs(ctx, now, c, errs)
		v.collectCluster(ctx, now, c)
	}
	v.collectResourcePools(ctx, now, errs)

	return errs.Combine()
}

func (v *vcenterMetricScraper) collectCluster(
	ctx context.Context,
	now pdata.Timestamp,
	c *object.ClusterComputeResource,
) {
	var moCluster mo.ClusterComputeResource
	c.Properties(ctx, c.Reference(), []string{"summary"}, &moCluster)
	s := moCluster.Summary.GetComputeResourceSummary()
	v.mb.RecordVcenterClusterCPULimitDataPoint(now, int64(s.TotalCpu))
	v.mb.RecordVcenterClusterHostCountDataPoint(now, int64(s.NumHosts-s.NumEffectiveHosts), "false")
	v.mb.RecordVcenterClusterHostCountDataPoint(now, int64(s.NumEffectiveHosts), "true")
	v.mb.EmitForResource(
		metadata.WithVcenterClusterName(c.Name()),
	)
}

func (v *vcenterMetricScraper) collectDatastores(
	ctx context.Context,
	colTime pdata.Timestamp,
	cluster *object.ClusterComputeResource,
	errs *scrapererror.ScrapeErrors,
) {
	datastores, err := cluster.Datastores(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	for _, ds := range datastores {
		v.collectDatastore(ctx, colTime, ds)
	}
}

func (v *vcenterMetricScraper) collectDatastore(
	ctx context.Context,
	now pdata.Timestamp,
	ds *object.Datastore,
) {
	var moDS mo.Datastore
	ds.Properties(ctx, ds.Reference(), []string{"summary", "name"}, &moDS)

	v.recordDatastoreProperties(now, moDS)
	v.mb.EmitForResource(
		metadata.WithVcenterDatastoreName(moDS.Name),
	)
}

func (v *vcenterMetricScraper) collectHosts(
	ctx context.Context,
	colTime pdata.Timestamp,
	cluster *object.ClusterComputeResource,
	errs *scrapererror.ScrapeErrors,
) {
	hosts, err := cluster.Hosts(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	for _, h := range hosts {
		v.collectHost(ctx, colTime, h, cluster, errs)
	}
}

func (v *vcenterMetricScraper) collectHost(
	ctx context.Context,
	now pdata.Timestamp,
	host *object.HostSystem,
	cluster *object.ClusterComputeResource,
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
	v.mb.EmitForResource(
		metadata.WithVcenterHostName(host.Name()),
		metadata.WithVcenterClusterName(cluster.Name()),
	)
}

func (v *vcenterMetricScraper) collectResourcePools(
	ctx context.Context,
	ts pdata.Timestamp,
	errs *scrapererror.ScrapeErrors,
) {
	rps, err := v.client.ResourcePools(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for _, rp := range rps {
		var moRP mo.ResourcePool
		rp.Properties(ctx, rp.Reference(), []string{
			"summary",
			"summary.quickStats",
			"name",
		}, &moRP)
		v.recordResourcePool(ts, moRP)
		v.mb.EmitForResource(metadata.WithVcenterResourcePoolName(rp.Name()))
	}
}

func (v *vcenterMetricScraper) collectVMs(
	ctx context.Context,
	colTime pdata.Timestamp,
	cluster *object.ClusterComputeResource,
	errs *scrapererror.ScrapeErrors,
) {
	vms, err := v.client.VMs(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	poweredOffVMs := 0
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

		v.collectVM(ctx, colTime, moVM, errs)
		v.mb.EmitForResource(
			metadata.WithVcenterVMName(vm.Name()),
			metadata.WithVcenterVMID(vmUUID),
			metadata.WithVcenterClusterName(cluster.Name()),
			metadata.WithVcenterHostName(hostname),
		)
	}

	v.mb.RecordVcenterClusterVMCountDataPoint(colTime, int64(len(vms))-int64(poweredOffVMs), metadata.AttributeVMCountPowerState.On)
	v.mb.RecordVcenterClusterVMCountDataPoint(colTime, int64(poweredOffVMs), metadata.AttributeVMCountPowerState.Off)
}

func (v *vcenterMetricScraper) collectVM(
	ctx context.Context,
	colTime pdata.Timestamp,
	vm mo.VirtualMachine,
	errs *scrapererror.ScrapeErrors,
) {
	v.recordVMUsages(colTime, vm)
	v.recordVMPerformance(ctx, vm, errs)
}
