// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package vcenterreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
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
	client *vcenterClient
	config *Config
	mb     *metadata.MetricsBuilder
	logger *zap.Logger

	dsRefToComputesByRef map[string]map[string]*mo.ComputeResource
	vmToComputeMap       map[string]string
	vmToRPool            map[string]*mo.ResourcePool
	datacenters          []*mo.Datacenter
	datastores           []*mo.Datastore
	rPoolIPathsByRef     map[string]*string
	rPoolsByRef          map[string]*mo.ResourcePool
	computesByRef        map[string]*mo.ComputeResource
	hostsByRef           map[string]*mo.HostSystem
	hostPerfMetricsByRef map[string]*performance.EntityMetric
	vmsByRef             map[string]*mo.VirtualMachine
	vmPerfMetricsByRef   map[string]*performance.EntityMetric
}

func newVmwareVcenterScraper(
	logger *zap.Logger,
	config *Config,
	settings receiver.CreateSettings,
) *vcenterMetricScraper {
	client := newVcenterClient(config)
	return &vcenterMetricScraper{
		client:               client,
		config:               config,
		logger:               logger,
		mb:                   metadata.NewMetricsBuilder(config.MetricsBuilderConfig, settings),
		dsRefToComputesByRef: make(map[string]map[string]*mo.ComputeResource),
		vmToComputeMap:       make(map[string]string),
		vmToRPool:            make(map[string]*mo.ResourcePool),
		datacenters:          make([]*mo.Datacenter, 0),
		datastores:           make([]*mo.Datastore, 0),
		rPoolIPathsByRef:     make(map[string]*string),
		computesByRef:        make(map[string]*mo.ComputeResource),
		hostsByRef:           make(map[string]*mo.HostSystem),
		hostPerfMetricsByRef: make(map[string]*performance.EntityMetric),
		rPoolsByRef:          make(map[string]*mo.ResourcePool),
		vmsByRef:             make(map[string]*mo.VirtualMachine),
		vmPerfMetricsByRef:   make(map[string]*performance.EntityMetric),
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
	v.logger.Warn("START SCRAPE")
	if v.client == nil {
		v.client = newVcenterClient(v.config)
	}

	// ensure connection before scraping
	if err := v.client.EnsureConnection(ctx); err != nil {
		return pmetric.NewMetrics(), fmt.Errorf("unable to connect to vSphere SDK: %w", err)
	}

	errs := &scrapererror.ScrapeErrors{}
	err := v.collectAllAndBuildMetrics(ctx, errs)

	v.logger.Warn("END SCRAPE")
	return v.mb.Emit(), err
}

// collectAllAndBuildMetrics collects & converts all relevant resources managed by vCenter to OTEL resources & metrics
func (v *vcenterMetricScraper) collectAllAndBuildMetrics(ctx context.Context, errs *scrapererror.ScrapeErrors) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	v.collectAllRPoolsWithInventoryLists(ctx, errs)
	v.collectDatacenters(ctx, errs)
	for _, dc := range v.datacenters {
		v.collectDatastores(ctx, dc, errs)
		v.collectComputes(ctx, dc, errs)
		v.collectHosts(ctx, dc, errs)
		v.collectResourcePools(ctx, dc, errs)
		v.collectVMs(ctx, dc, errs)
		runtime.ReadMemStats(&m)
		v.buildMetrics(dc, errs)
	}
	v.clearCollectData()
	runtime.ReadMemStats(&m)
	return errs.Combine()
}

// collectAllRPoolsWithInventoryLists collects and store all ResourcePools with their InventoryLists
func (v *vcenterMetricScraper) collectAllRPoolsWithInventoryLists(ctx context.Context, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.rPoolIPathsByRef = make(map[string]*string)

	// Get ResourcePools with InventoryLists for later retrieval
	rps, err := v.client.AllResourcePoolWithInventoryLists(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i := range rps {
		v.rPoolIPathsByRef[rps[i].Reference().Value] = &rps[i].InventoryPath
	}
}

// collectDatacenters collects and store all relevant property data
func (v *vcenterMetricScraper) collectDatacenters(ctx context.Context, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.datacenters = make([]*mo.Datacenter, 0)

	// Get Datacenters w/properties and store for later retrieval
	dcs, err := v.client.Datacenters(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i := range dcs {
		v.datacenters = append(v.datacenters, &dcs[i])
	}
}

// collectDatastores collects and store all relevant property data for a Datacenter's Datastores
func (v *vcenterMetricScraper) collectDatastores(ctx context.Context, dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.datastores = make([]*mo.Datastore, 0)

	// Get Datastores w/properties and store for later retrieval
	dss, err := v.client.Datastores(ctx, dc.Reference())
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i := range dss {
		v.datastores = append(v.datastores, &dss[i])
	}
}

// collectComputes collects and store all relevant property data for a Datacenter's ComputeResources/ClusterComputeResources
func (v *vcenterMetricScraper) collectComputes(ctx context.Context, dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.dsRefToComputesByRef = make(map[string]map[string]*mo.ComputeResource)
	v.computesByRef = make(map[string]*mo.ComputeResource)

	// Get ComputeResources/ClusterComputeResources w/properties and store for later retrieval
	crs, err := v.client.ComputeResources(ctx, dc.Reference())
	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	for i := range crs {
		// Store mapping of datastores to group of computes
		for _, dsRef := range crs[i].Datastore {
			csByRef := v.dsRefToComputesByRef[dsRef.Value]
			if csByRef == nil {
				csByRef = make(map[string]*mo.ComputeResource)
				v.dsRefToComputesByRef[dsRef.Value] = csByRef
			}
			csByRef[crs[i].Reference().Value] = &crs[i]
		}

		v.computesByRef[crs[i].Reference().Value] = &crs[i]
	}
}

// collectHosts collects and store all relevant metric/property data for a Datacenter's HostSystems
func (v *vcenterMetricScraper) collectHosts(ctx context.Context, dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.hostsByRef = make(map[string]*mo.HostSystem)

	// Get HostSystems w/properties and store for later retrieval
	hss, err := v.client.HostSystems(ctx, dc.Reference())
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	hsRefs := []types.ManagedObjectReference{}
	for i := range hss {
		hsRefs = append(hsRefs, hss[i].Reference())
		v.hostsByRef[hss[i].Reference().Value] = &hss[i]
	}

	spec := types.PerfQuerySpec{
		MaxSample: 5,
		Format:    string(types.PerfFormatNormal),
		// Just grabbing real time performance metrics of the current
		// supported metrics by this receiver. If more are added we may need
		// a system of making this user customizable or adapt to use a 5 minute interval per metric
		IntervalId: int32(20),
	}
	// Get all HostSystem performance metrics and store for later retrieval
	results, err := v.client.perfMetricsQuery(ctx, spec, hostPerfMetricList, hsRefs)
	if err != nil {
		errs.AddPartial(1, fmt.Errorf("failed to retrieve perf metrics for HostSystems: %w", err))
		return
	}
	v.hostPerfMetricsByRef = results.resultsByRef
}

// collectResourcePools collects and store all relevant property data for a Datacenter's ResourcePools
func (v *vcenterMetricScraper) collectResourcePools(ctx context.Context, dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.rPoolsByRef = make(map[string]*mo.ResourcePool)

	// Get ResourcePools w/properties and store for later retrieval
	rps, err := v.client.ResourcePools(ctx, dc.Reference())
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i := range rps {
		v.rPoolsByRef[rps[i].Reference().Value] = &rps[i]
	}
}

// collectVMs collects and store all relevant metric/property data for a Datacenter's VirtualMachines
func (v *vcenterMetricScraper) collectVMs(ctx context.Context, dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.vmsByRef = make(map[string]*mo.VirtualMachine)

	// Get VirtualMachines w/properties and store for later retrieval
	vms, err := v.client.VMs(ctx, dc.Reference())
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	vmRefs := []types.ManagedObjectReference{}
	for i := range vms {
		vmRefs = append(vmRefs, vms[i].Reference())
		v.vmsByRef[vms[i].Reference().Value] = &vms[i]
	}

	spec := types.PerfQuerySpec{
		Format: string(types.PerfFormatNormal),
		// Just grabbing real time performance metrics of the current
		// supported metrics by this receiver. If more are added we may need
		// a system of making this user customizable or adapt to use a 5 minute interval per metric
		IntervalId: int32(20),
	}
	// Get all VirtualMachine performance metrics and store for later retrieval
	results, err := v.client.perfMetricsQuery(ctx, spec, vmPerfMetricList, vmRefs)
	if err != nil {
		errs.AddPartial(1, fmt.Errorf("failed to retrieve perf metrics for VirtualMachines: %w", err))
		return
	}
	v.vmPerfMetricsByRef = results.resultsByRef
}

// clearCollectData clears collection data after it is no longer needed
func (v *vcenterMetricScraper) clearCollectData() {
	v.vmToComputeMap = nil
	v.vmToRPool = nil
	v.dsRefToComputesByRef = nil
	v.datacenters = nil
	v.datastores = nil
	v.rPoolIPathsByRef = nil
	v.computesByRef = nil
	v.hostsByRef = nil
	v.hostPerfMetricsByRef = nil
	v.rPoolsByRef = nil
	v.vmsByRef = nil
	v.vmPerfMetricsByRef = nil
}

// buildMetrics creates all of the metrics from the stored collected data
func (v *vcenterMetricScraper) buildMetrics(dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.vmToComputeMap = make(map[string]string)
	v.vmToRPool = make(map[string]*mo.ResourcePool)
	now := pcommon.NewTimestampFromTime(time.Now())

	v.buildDatastores(now, dc)
	v.buildResourcePools(now, dc, errs)
	for _, cr := range v.computesByRef {
		v.buildHostsForCompute(now, dc, cr, errs)
		poweredOnVMs, poweredOffVMs := v.buildVmsForCompute(now, dc, cr, errs)
		if cr.Reference().Type == "ClusterComputeResource" {
			v.buildClusters(now, dc, cr, poweredOnVMs, poweredOffVMs)
		}
	}
}

// buildDatastores builds the Datastore metrics and resources from the collected data
func (v *vcenterMetricScraper) buildDatastores(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
) {
	for _, ds := range v.datastores {
		crsByRef := v.dsRefToComputesByRef[ds.Reference().Value]
		crs := []*mo.ComputeResource{}
		if crsByRef == nil {
			v.buildDatastoreMetrics(colTime, dc, ds, crs)
			continue
		}

		for _, cr := range crsByRef {
			crs = append(crs, cr)
		}
		v.buildDatastoreMetrics(colTime, dc, ds, crs)
	}
}

// buildDatastores builds a resource and metrics for a given collected Datastore
func (v *vcenterMetricScraper) buildDatastoreMetrics(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	ds *mo.Datastore,
	crs []*mo.ComputeResource,
) {
	v.recordDatastoreStats(colTime, ds)

	rb := v.createDatastoreResourceBuilder(dc, ds, crs)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))
}

func (v *vcenterMetricScraper) buildResourcePools(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	errs *scrapererror.ScrapeErrors,
) {
	for _, rp := range v.rPoolsByRef {
		if err := v.buildResourcePoolMetrics(colTime, dc, rp); err != nil {
			errs.AddPartial(1, err)
		}
	}
}

// buildResourcePoolMetricsForCompute builds a resource and metrics for a given collected ResourcePool
func (v *vcenterMetricScraper) buildResourcePoolMetrics(colTime pcommon.Timestamp, dc *mo.Datacenter, rp *mo.ResourcePool,
) error {
	crRef := rp.Owner
	cr := v.computesByRef[crRef.Value]
	for _, vmRef := range rp.Vm {
		v.vmToComputeMap[vmRef.Value] = crRef.Value
		v.vmToRPool[vmRef.Value] = rp
	}

	rb, err := v.createResourcePoolResourceBuilder(dc, cr, rp)
	if err != nil {
		return err
	}

	v.recordResourcePoolStats(colTime, rp)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return nil
}

// buildHostsForCompute builds Host metrics and resources from the collected data for a single ComputeResource
func (v *vcenterMetricScraper) buildHostsForCompute(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
	errs *scrapererror.ScrapeErrors,
) {
	hsRefs := cr.Host
	if hsRefs == nil || len(hsRefs) == 0 {
		errs.AddPartial(1, fmt.Errorf("no Host refs for ComputeResource: %s", cr.Name))
		return
	}

	for i := range hsRefs {
		if err := v.buildHostMetricsForCompute(colTime, dc, cr, &hsRefs[i]); err != nil {
			errs.AddPartial(1, err)
		}
	}
}

// buildHostMetricsForCompute builds a resource and metrics for a given collected Host
func (v *vcenterMetricScraper) buildHostMetricsForCompute(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
	hsRefs *types.ManagedObjectReference,
) error {
	hs := v.hostsByRef[hsRefs.Value]
	if hs == nil {
		return fmt.Errorf("no collected Host for ComputeResource [%s]'s Host ref: %s", cr.Name, hsRefs.Value)
	}
	for _, vmRef := range hs.Vm {
		v.vmToComputeMap[vmRef.Value] = cr.Reference().Value
	}

	v.recordHostSystemStats(colTime, hs)
	hostPerfMetrics := v.hostPerfMetricsByRef[hsRefs.Value]
	if hostPerfMetrics != nil {
		v.recordHostPerformanceMetrics(hostPerfMetrics)
	}

	rb := v.createHostResourceBuilder(dc, cr, hs)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return nil
}

func (v *vcenterMetricScraper) buildClusters(
	now pcommon.Timestamp,
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
	poweredOnVMs, poweredOffVMs int64,
) {
	v.recordClusterStats(now, cr, poweredOnVMs, poweredOffVMs)

	rb := v.createClusterResourceBuilder(dc, cr)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))
}

func (v *vcenterMetricScraper) buildVmsForCompute(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
	errs *scrapererror.ScrapeErrors,
) (poweredOnVMs int64, poweredOffVMs int64) {
	for _, vm := range v.vmsByRef {
		incOn, incOff, err := v.buildVMForCompute(colTime, dc, cr, vm)
		if incOff {
			poweredOffVMs++
		}

		if incOn {
			poweredOnVMs++
		}

		if err != nil {
			errs.AddPartial(1, err)
		}
	}

	return poweredOnVMs, poweredOffVMs
}

func (v *vcenterMetricScraper) buildVMForCompute(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
	vm *mo.VirtualMachine,
) (incOn bool, incOff bool, err error) {
	crRef, ok := v.vmToComputeMap[vm.Reference().Value]
	if !ok {
		return incOn, incOff, err
	}

	if crRef != cr.Reference().Value {
		return incOn, incOff, err
	}

	if string(vm.Runtime.PowerState) == "poweredOff" {
		incOff = true
	} else {
		incOn = true
	}

	// vms are optional without a resource pool
	rpRef := vm.ResourcePool
	var rp *mo.ResourcePool
	if rpRef != nil {
		rp = v.rPoolsByRef[rpRef.Value]
	}

	if rp != nil {
		rpCompute := rp.Owner
		// not part of this cluster
		if rpCompute.Reference().Value != cr.Reference().Value {
			return incOn, incOff, err
		}
		stored, ok := v.vmToRPool[vm.Reference().Value]
		if ok {
			rp = stored
		}
	}

	if vm.Config == nil {
		return incOn, incOff, fmt.Errorf("config empty for VM: %s", vm.Name)
	}

	// Get related VM host info
	hsRef := vm.Summary.Runtime.Host
	if hsRef == nil {
		return incOn, incOff, fmt.Errorf("no Host ref for VM: %s", vm.Name)
	}
	hs := v.hostsByRef[hsRef.Value]
	if hs == nil {
		return incOn, incOff, fmt.Errorf("no collected Host for VM [%s]'s Host ref: %s", vm.Name, hsRef.Value)
	}

	rb, err := v.createVMResourceBuilder(dc, cr, hs, rp, vm)
	if err != nil {
		return incOn, incOff, err
	}

	perfMetrics := v.vmPerfMetricsByRef[vm.Reference().Value]
	v.recordVMStats(colTime, vm, hs)
	if perfMetrics != nil {
		v.recordVMPerformanceMetrics(perfMetrics)
	}
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return incOn, incOff, err
}
