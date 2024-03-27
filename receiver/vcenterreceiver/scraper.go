// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package vcenterreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scrapererror"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver/internal/metadata"
)

const (
	emitPerfMetricsWithObjectsFeatureGateID = "receiver.vcenter.emitPerfMetricsWithObjects"
)

var _ = featuregate.GlobalRegistry().MustRegister(
	emitPerfMetricsWithObjectsFeatureGateID,
	featuregate.StageStable,
	featuregate.WithRegisterToVersion("v0.97.0"),
)

var _ receiver.Metrics = (*vcenterMetricScraper)(nil)

type vcenterMetricScraper struct {
	client                *vcenterClient
	config                *Config
	mb                    *metadata.MetricsBuilder
	logger                *zap.Logger
	vmTotalCollectTime    time.Duration
	vmPrePerfNetworkTime  time.Duration
	vmPerfNetworkTime     time.Duration
	vmPostPerfNetworkTime time.Duration
	vmNetworkTime         time.Duration
	vmContainerViewTime   time.Duration
	vmRetrieveTime        time.Duration
	AltVMNetworkTime      time.Duration
	vmPartA1              time.Duration
	vmPartA2              time.Duration
	vmPartA3              time.Duration
	vmPartA4              time.Duration
	vmPartA5              time.Duration
	vmPartB               time.Duration
	vmPartC               time.Duration
	vmRecordVMUsagesTime  time.Duration
	hostPerfNetworkTime   time.Duration

	// map of vm name => compute name
	vmToComputeMap    map[string]string
	vmToResourcePool  map[string]*object.ResourcePool
	vmToResourcePool2 map[string]*mo.ResourcePool
}

func newVmwareVcenterScraper(
	logger *zap.Logger,
	config *Config,
	settings receiver.CreateSettings,
) *vcenterMetricScraper {
	client := newVcenterClient(config)
	return &vcenterMetricScraper{
		client:            client,
		config:            config,
		logger:            logger,
		mb:                metadata.NewMetricsBuilder(config.MetricsBuilderConfig, settings),
		vmToComputeMap:    make(map[string]string),
		vmToResourcePool:  make(map[string]*object.ResourcePool),
		vmToResourcePool2: make(map[string]*mo.ResourcePool),
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
	v.vmPrePerfNetworkTime = 0
	v.vmPerfNetworkTime = 0
	v.vmPostPerfNetworkTime = 0
	v.vmPartA1 = 0
	v.vmPartA2 = 0
	v.vmPartA3 = 0
	v.vmPartA4 = 0
	v.vmPartA5 = 0
	v.vmPartB = 0
	v.vmPartC = 0
	v.vmRecordVMUsagesTime = 0
	v.logger.Warn("LOG: START SCRAPE")
	if v.client == nil {
		v.client = newVcenterClient(v.config)
	}

	// ensure connection before scraping
	if err := v.client.EnsureConnection(ctx); err != nil {
		return pmetric.NewMetrics(), fmt.Errorf("unable to connect to vSphere SDK: %w", err)
	}
	v.logger.Warn("LOG: START COLLECT DATACENTERS")
	err := v.collectDatacenters(ctx)
	v.logger.Warn("LOG: END COLLECT DATACENTERS")

	// cleanup so any inventory moves are accounted for
	v.vmToComputeMap = make(map[string]string)
	v.vmToResourcePool = make(map[string]*object.ResourcePool)
	v.vmToResourcePool2 = make(map[string]*mo.ResourcePool)
	v.logger.Warn("LOG: END SCRAPE")
	v.logger.Warn("LOG: OVERALL VM COLLECT TIME: " + v.vmTotalCollectTime.String())
	v.logger.Warn("LOG: TOTAL VM NETWORK TIME: " + v.vmNetworkTime.String())
	v.logger.Warn("LOG: TOTAL VM ALT NETWORK TIME: " + v.AltVMNetworkTime.String())
	v.logger.Warn("LOG: TOTAL VM CONTAINER VIEW TIME: " + v.vmContainerViewTime.String())
	v.logger.Warn("LOG: TOTAL VM RETRIEVE TIME: " + v.vmRetrieveTime.String())
	v.logger.Warn("LOG: TOTAL VM PART A1 TIME: " + v.vmPartA1.String())
	v.logger.Warn("LOG: TOTAL VM PART A2 TIME: " + v.vmPartA2.String())
	v.logger.Warn("LOG: TOTAL VM PART A3 TIME: " + v.vmPartA3.String())
	v.logger.Warn("LOG: TOTAL VM PART A4 TIME: " + v.vmPartA4.String())
	v.logger.Warn("LOG: TOTAL VM PART A5 TIME: " + v.vmPartA5.String())
	v.logger.Warn("LOG: TOTAL VM PART B TIME: " + v.vmPartB.String())
	v.logger.Warn("LOG: TOTAL VM PART C TIME: " + v.vmPartC.String())
	v.logger.Warn("LOG: TOTAL VM PRE PERF NETWORK TIME: " + v.vmPrePerfNetworkTime.String())
	v.logger.Warn("LOG: TOTAL VM PERF NETWORK TIME: " + v.vmPerfNetworkTime.String())
	v.logger.Warn("LOG: TOTAL VM POST PERF NETWORK TIME: " + v.vmPostPerfNetworkTime.String())
	v.logger.Warn("LOG: TOTAL HOST PERF NETWORK TIME: " + v.hostPerfNetworkTime.String())
	v.logger.Warn("LOG: END SCRAPE")
	return v.mb.Emit(), err
}

func (v *vcenterMetricScraper) collectDatacenters(ctx context.Context) error {
	v.logger.Warn("LOG: START NETWORK DATACENTERS")
	datacenters, err := v.client.Datacenters(ctx)
	v.logger.Warn("LOG: END NETWORK DATACENTERS")
	if err != nil {
		return err
	}
	errs := &scrapererror.ScrapeErrors{}
	for i, dc := range datacenters {
		v.logger.Warn("LOG: START COLLECT CLUSTERS" + strconv.FormatInt(int64(i), 10))
		v.collectClusters(ctx, dc, errs)
		v.logger.Warn("LOG: END COLLECT CLUSTERS" + strconv.FormatInt(int64(i), 10))
	}
	return errs.Combine()
}

func (v *vcenterMetricScraper) collectClusters(ctx context.Context, datacenter *object.Datacenter, errs *scrapererror.ScrapeErrors) {
	v.logger.Warn("LOG: START NETWORK COMPUTES")
	computes, err := v.client.Computes(ctx, datacenter)
	v.logger.Warn("LOG: END NETWORK COMPUTES")
	if err != nil {
		errs.Add(err)
		return
	}

	now := pcommon.NewTimestampFromTime(time.Now())

	v.logger.Warn("LOG: START COLLECT RPOOLS")
	v.collectResourcePools(ctx, now, errs)
	v.logger.Warn("LOG: END COLLECT RPOOLS")
	for i, c := range computes {
		v.logger.Warn("LOG: START COLLECT HOSTS" + strconv.FormatInt(int64(i), 10))
		v.collectHosts(ctx, now, c, errs)
		v.logger.Warn("LOG: END COLLECT HOSTS" + strconv.FormatInt(int64(i), 10))
		v.logger.Warn("LOG: START COLLECT DATASTORES" + strconv.FormatInt(int64(i), 10))
		v.collectDatastores(ctx, now, c, errs)
		v.logger.Warn("LOG: END COLLECT DATASTORES" + strconv.FormatInt(int64(i), 10))
		start := time.Now()
		poweredOnVMs, poweredOffVMs := v.collectVMs(ctx, now, c, errs)
		end := time.Now()
		v.vmTotalCollectTime = end.Sub(start)
		if c.Reference().Type == "ClusterComputeResource" {
			v.logger.Warn("LOG: START COLLECT CLUSTER" + strconv.FormatInt(int64(i), 10))
			v.collectCluster(ctx, now, c, poweredOnVMs, poweredOffVMs, errs)
			v.logger.Warn("LOG: END COLLECT CLUSTER" + strconv.FormatInt(int64(i), 10))
		}
	}
}

func (v *vcenterMetricScraper) collectCluster(
	ctx context.Context,
	now pcommon.Timestamp,
	c *object.ComputeResource,
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
	v.mb.RecordVcenterClusterHostCountDataPoint(now, int64(s.NumHosts-s.NumEffectiveHosts), false)
	v.mb.RecordVcenterClusterHostCountDataPoint(now, int64(s.NumEffectiveHosts), true)
	rb := v.mb.NewResourceBuilder()
	rb.SetVcenterClusterName(c.Name())
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))
}

func (v *vcenterMetricScraper) collectDatastores(
	ctx context.Context,
	colTime pcommon.Timestamp,
	compute *object.ComputeResource,
	errs *scrapererror.ScrapeErrors,
) {
	datastores, err := compute.Datastores(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	for i, ds := range datastores {
		v.logger.Warn("LOG: START COLLECT DATASTORE" + strconv.FormatInt(int64(i), 10))
		v.collectDatastore(ctx, colTime, ds, compute, errs)
		v.logger.Warn("LOG: START COLLECT DATASTORE" + strconv.FormatInt(int64(i), 10))
	}
}

func (v *vcenterMetricScraper) collectDatastore(
	ctx context.Context,
	now pcommon.Timestamp,
	ds *object.Datastore,
	compute *object.ComputeResource,
	errs *scrapererror.ScrapeErrors,
) {
	var moDS mo.Datastore
	err := ds.Properties(ctx, ds.Reference(), []string{"summary", "name"}, &moDS)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	v.recordDatastoreProperties(now, moDS)
	rb := v.mb.NewResourceBuilder()
	if compute.Reference().Type == "ClusterComputeResource" {
		rb.SetVcenterClusterName(compute.Name())
	}
	rb.SetVcenterDatastoreName(moDS.Name)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))
}

func (v *vcenterMetricScraper) collectHosts(
	ctx context.Context,
	colTime pcommon.Timestamp,
	compute *object.ComputeResource,
	errs *scrapererror.ScrapeErrors,
) {
	hosts, err := compute.Hosts(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	for i, h := range hosts {
		v.logger.Warn("LOG: START COLLECT HOST" + strconv.FormatInt(int64(i), 10))
		v.collectHost(ctx, colTime, h, compute, errs)
		v.logger.Warn("LOG: END COLLECT HOST" + strconv.FormatInt(int64(i), 10))
	}
}

func (v *vcenterMetricScraper) collectHost(
	ctx context.Context,
	now pcommon.Timestamp,
	host *object.HostSystem,
	compute *object.ComputeResource,
	errs *scrapererror.ScrapeErrors,
) {

	var hwSum mo.HostSystem
	err := host.Properties(ctx, host.Reference(),
		[]string{
			"name",
			"config",
			"summary.hardware",
			"summary.quickStats",
			"vm",
		}, &hwSum)

	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	for _, vmRef := range hwSum.Vm {
		v.logger.Warn("LOG: START VM TO COMPUTEMAP")
		v.vmToComputeMap[vmRef.Value] = compute.Name()
		v.logger.Warn("LOG: END VM TO COMPUTEMAP")
	}

	v.recordHostSystemMemoryUsage(now, hwSum)
	v.hostPerfNetworkTime += v.recordHostPerformanceMetrics(ctx, hwSum, errs)
	rb := v.mb.NewResourceBuilder()
	rb.SetVcenterHostName(host.Name())
	rb.SetVcenterClusterName(compute.Name())
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))
}

func (v *vcenterMetricScraper) collectResourcePools(
	ctx context.Context,
	ts pcommon.Timestamp,
	errs *scrapererror.ScrapeErrors,
) {
	v.logger.Warn("LOG: START NETWORK RPOOLS")
	rps, err := v.client.ResourcePools(ctx)
	v.logger.Warn("LOG: END NETWORK RPOOLS")
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i, rp := range rps {
		v.logger.Warn("LOG: START COLLECT RPOOL" + strconv.FormatInt(int64(i), 10))
		var moRP mo.ResourcePool
		err = rp.Properties(ctx, rp.Reference(), []string{
			"summary",
			"summary.quickStats",
			"name",
			"vm",
		}, &moRP)
		if err != nil {
			errs.AddPartial(1, err)
			continue
		}

		computeRef, err := rp.Owner(ctx)
		if err != nil {
			errs.AddPartial(1, err)
			continue
		}
		for _, vmRef := range moRP.Vm {
			v.logger.Warn("LOG: START VM TO COMPUTEMAP")
			v.vmToComputeMap[vmRef.Value] = computeRef.Reference().Value
			v.logger.Warn("LOG: END VM TO COMPUTEMAP")
			v.logger.Warn("LOG: START VM TO RPOOL")
			v.vmToResourcePool[vmRef.Value] = rp
			v.logger.Warn("LOG: END VM TO RPOOL")
		}

		v.recordResourcePool(ts, moRP)
		rb := v.mb.NewResourceBuilder()
		rb.SetVcenterResourcePoolName(rp.Name())
		rb.SetVcenterResourcePoolInventoryPath(rp.InventoryPath)
		v.mb.EmitForResource(metadata.WithResource(rb.Emit()))
		v.logger.Warn("LOG: END COLLECT RPOOL" + strconv.FormatInt(int64(i), 10))
	}

	moRPS, err := v.client.AltResourcePools(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i, rp := range moRPS {
		computeRef := rp.Owner

		for _, vmRef := range rp.Vm {
			v.vmToComputeMap[vmRef.Value] = computeRef.Reference().Value
			v.vmToResourcePool2[vmRef.Value] = &rp
		}

		v.recordResourcePool(ts, rp)
		rb := v.mb.NewResourceBuilder()
		rb.SetVcenterResourcePoolName(rp.Name)
		rb.SetVcenterResourcePoolInventoryPath(rp.Reference().ServerGUID)
		v.mb.EmitForResource(metadata.WithResource(rb.Emit()))
		v.logger.Warn("LOG: END COLLECT RPOOL" + strconv.FormatInt(int64(i), 10))
	}
}

func (v *vcenterMetricScraper) collectVMs(
	ctx context.Context,
	colTime pcommon.Timestamp,
	compute *object.ComputeResource,
	errs *scrapererror.ScrapeErrors,
) (poweredOnVMs int64, poweredOffVMs int64) {
	start := time.Now()
	// _, err := v.client.VMs(ctx)
	end := time.Now()
	v.vmNetworkTime = end.Sub(start)
	// if err != nil {
	// 	errs.AddPartial(1, err)
	// 	return
	// }

	start = time.Now()
	vms, err := v.client.AltVMs(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	end = time.Now()
	v.AltVMNetworkTime = end.Sub(start)
	v.logger.Warn("LOG: ALT VM COUNT: " + (strconv.Itoa(len(vms))))

	for _, vm := range vms {
		start = time.Now()
		computeName, ok := v.vmToComputeMap[vm.Reference().Value]
		if !ok {
			v.logger.Warn("LOG: WHOOPS 1")
			continue
		}
		end = time.Now()
		v.vmPartA1 += end.Sub(start)

		start = time.Now()

		if computeName != compute.Reference().Value && computeName != compute.Name() {
			v.logger.Warn("LOG: WHOOPS 2")
			continue
		}
		end = time.Now()
		v.vmPartA2 += end.Sub(start)

		// start = time.Now()

		// var moVM mo.VirtualMachine
		// err := vm.Properties(ctx, vm.Reference(), []string{
		// 	"config",
		// 	"runtime",
		// 	"summary",
		// }, &moVM)

		// if err != nil {
		// 	errs.AddPartial(1, err)
		// 	continue
		// }
		// end = time.Now()
		// v.vmPartA3 += end.Sub(start)

		if string(vm.Runtime.PowerState) == "poweredOff" {
			poweredOffVMs++
		} else {
			poweredOnVMs++
		}

		start = time.Now()
		hostRef := vm.Summary.Runtime.Host
		if hostRef == nil {
			errs.AddPartial(1, errors.New("VM doesn't have a HostSystem"))
			v.logger.Warn("LOG: BIIIIIG WHOOPS 1")
			return
		}

		vmHost := object.NewHostSystem(v.client.vimDriver, *hostRef)

		// vmHost, err := vm.HostSystem(ctx)
		// if err != nil {
		// 	errs.AddPartial(1, err)
		// 	return
		// }
		end = time.Now()
		v.vmPartA5 += end.Sub(start)

		start = time.Now()
		// vms are optional without a resource pool
		// rp, _ := vm.ResourcePool(ctx)
		rpRef := vm.ResourcePool
		var rp *object.ResourcePool
		if rpRef != nil {
			rp = object.NewResourcePool(v.client.vimDriver, *rpRef)
		}

		if rp != nil {
			rpCompute, rpErr := rp.Owner(ctx)
			if rpErr != nil {
				errs.AddPartial(1, err)
				v.logger.Warn("LOG: BIIIIIG WHOOPS 2")
				return
			}
			// not part of this cluster
			if rpCompute.Reference().Value != compute.Reference().Value {
				v.logger.Warn("LOG: WHOOPS 3")
				continue
			}
			stored, ok := v.vmToResourcePool[vm.Reference().Value]
			if ok {
				rp = stored
			}
		}
		end = time.Now()
		v.vmPartB += end.Sub(start)

		start = time.Now()

		// hostname, err := vmHost.ObjectName(ctx)
		// if err != nil {
		// 	errs.AddPartial(1, err)
		// 	return
		// }

		var hwSum mo.HostSystem
		err = vmHost.Properties(ctx, vmHost.Reference(),
			[]string{
				"name",
				"summary.hardware",
				"vm",
			}, &hwSum)

		if err != nil {
			errs.AddPartial(1, err)
			v.logger.Warn("LOG: BIIIIIG WHOOPS 3")
			return
		}

		if vm.Config == nil {
			errs.AddPartial(1, fmt.Errorf("vm config empty for %s", hwSum.Name))
			v.logger.Warn("LOG: WHOOPS 4")
			continue
		}
		vmUUID := vm.Config.InstanceUuid

		end = time.Now()
		v.vmPartC += end.Sub(start)
		v.collectVM(ctx, colTime, vm, hwSum, errs)
		rb := v.mb.NewResourceBuilder()
		rb.SetVcenterVMName(vm.Summary.Config.Name)
		rb.SetVcenterVMID(vmUUID)
		if compute.Reference().Type == "ClusterComputeResource" {
			rb.SetVcenterClusterName(compute.Name())
		}
		rb.SetVcenterHostName(hwSum.Name)
		if rp != nil && rp.Name() != "" {
			rb.SetVcenterResourcePoolName(rp.Name())
			rb.SetVcenterResourcePoolInventoryPath(rp.InventoryPath)
		}
		v.mb.EmitForResource(metadata.WithResource(rb.Emit()))
	}
	return poweredOnVMs, poweredOffVMs
}

func (v *vcenterMetricScraper) collectVM(
	ctx context.Context,
	colTime pcommon.Timestamp,
	vm mo.VirtualMachine,
	hs mo.HostSystem,
	errs *scrapererror.ScrapeErrors,
) {
	start := time.Now()
	v.recordVMUsages(colTime, vm, hs)
	end := time.Now()
	v.vmRecordVMUsagesTime += end.Sub(start)
	pre, request, post := v.recordVMPerformance(ctx, vm, errs)
	v.vmPrePerfNetworkTime += pre
	v.vmPerfNetworkTime += request
	v.vmPostPerfNetworkTime += post
}
