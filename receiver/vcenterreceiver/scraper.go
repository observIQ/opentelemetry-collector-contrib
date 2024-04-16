// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package vcenterreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"

import (
	"context"
	"fmt"
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

type clusterVMStates struct {
	poweredOn  int64
	poweredOff int64
}

type vcenterMetricScraper struct {
	client *vcenterClient
	config *Config
	mb     *metadata.MetricsBuilder
	logger *zap.Logger

	computeRefsByDSRef   map[string]map[string]bool
	datacenters          []*mo.Datacenter
	datastores           []*mo.Datastore
	rPoolIPathsByRef     map[string]*string
	vAppIPathsByRef      map[string]*string
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
		computeRefsByDSRef:   make(map[string]map[string]bool),
		datacenters:          make([]*mo.Datacenter, 0),
		datastores:           make([]*mo.Datastore, 0),
		rPoolIPathsByRef:     make(map[string]*string),
		vAppIPathsByRef:      make(map[string]*string),
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
	err := v.scrapeAndBuildAllMetrics(ctx, errs)

	v.logger.Warn("END SCRAPE")
	return v.mb.Emit(), err
}

// scrapeAndBuildAllMetrics collects & converts all relevant resources managed by vCenter to OTEL resources & metrics
func (v *vcenterMetricScraper) scrapeAndBuildAllMetrics(ctx context.Context, errs *scrapererror.ScrapeErrors) error {
	v.scrapeResourcePoolInventoryListObjects(ctx, errs)
	v.scrapeVAppInventoryListObjects(ctx, errs)
	v.scrapeDatacenters(ctx, errs)

	for _, dc := range v.datacenters {
		v.scrapeDatastores(ctx, dc, errs)
		v.scrapeComputes(ctx, dc, errs)
		v.scrapeHosts(ctx, dc, errs)
		v.scrapeResourcePools(ctx, dc, errs)
		v.scrapeVirtualMachines(ctx, dc, errs)

		// Build metrics now that all vCenter data has been scraped for a single datacenter
		v.buildMetrics(dc, errs)
	}
	v.clearScrapeData()

	return errs.Combine()
}

// scrapeResourcePoolInventoryListObjects scrapes and store all ResourcePool objects with their InventoryLists
func (v *vcenterMetricScraper) scrapeResourcePoolInventoryListObjects(ctx context.Context, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.rPoolIPathsByRef = make(map[string]*string)

	// Get ResourcePools with InventoryLists and store for later retrieval
	rPools, err := v.client.ResourcePoolInventoryListObjects(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i := range rPools {
		v.rPoolIPathsByRef[rPools[i].Reference().Value] = &rPools[i].InventoryPath
	}
}

// scrapeVAppInventoryListObjects scrapes and stores all vApp objects with their InventoryLists
func (v *vcenterMetricScraper) scrapeVAppInventoryListObjects(ctx context.Context, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.vAppIPathsByRef = make(map[string]*string)

	// Get vApps with InventoryLists and store for later retrieval
	vApps, err := v.client.VAppInventoryListObjects(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i := range vApps {
		v.vAppIPathsByRef[vApps[i].Reference().Value] = &vApps[i].InventoryPath
	}
}

// scrapeDatacenters scrapes and stores all relevant property data for all Datacenters
func (v *vcenterMetricScraper) scrapeDatacenters(ctx context.Context, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.datacenters = make([]*mo.Datacenter, 0)

	// Get Datacenters w/properties and store for later retrieval
	datacenters, err := v.client.Datacenters(ctx)
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i := range datacenters {
		v.datacenters = append(v.datacenters, &datacenters[i])
	}
}

// scrapeDatastores scrapes and stores all relevant property data for a Datacenter's Datastores
func (v *vcenterMetricScraper) scrapeDatastores(ctx context.Context, dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.datastores = make([]*mo.Datastore, 0)

	// Get Datastores w/properties and store for later retrieval
	datastores, err := v.client.Datastores(ctx, dc.Reference())
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i := range datastores {
		v.datastores = append(v.datastores, &datastores[i])
	}
}

// scrapeComputes scrapes and stores all relevant property data for a Datacenter's ComputeResources/ClusterComputeResources
func (v *vcenterMetricScraper) scrapeComputes(ctx context.Context, dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.computeRefsByDSRef = make(map[string]map[string]bool)
	v.computesByRef = make(map[string]*mo.ComputeResource)

	// Get ComputeResources/ClusterComputeResources w/properties and store for later retrieval
	computes, err := v.client.ComputeResources(ctx, dc.Reference())
	if err != nil {
		errs.AddPartial(1, err)
		return
	}

	for i := range computes {
		// Store relationships between each Datastore and the ComputeResources that use it
		for _, dsRef := range computes[i].Datastore {
			uniqueComputeRefs := v.computeRefsByDSRef[dsRef.Value]
			if uniqueComputeRefs == nil {
				uniqueComputeRefs = make(map[string]bool)
				v.computeRefsByDSRef[dsRef.Value] = uniqueComputeRefs
			}
			uniqueComputeRefs[computes[i].Reference().Value] = true
		}

		v.computesByRef[computes[i].Reference().Value] = &computes[i]
	}
}

// scrapeHosts scrapes and stores all relevant metric/property data for a Datacenter's HostSystems
func (v *vcenterMetricScraper) scrapeHosts(ctx context.Context, dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.hostsByRef = make(map[string]*mo.HostSystem)

	// Get HostSystems w/properties and store for later retrieval
	hosts, err := v.client.HostSystems(ctx, dc.Reference())
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	hsRefs := []types.ManagedObjectReference{}
	for i := range hosts {
		hsRefs = append(hsRefs, hosts[i].Reference())
		v.hostsByRef[hosts[i].Reference().Value] = &hosts[i]
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
	results, err := v.client.PerfMetricsQuery(ctx, spec, hostPerfMetricList, hsRefs)
	if err != nil {
		errs.AddPartial(1, fmt.Errorf("failed to retrieve perf metrics for HostSystems: %w", err))
		return
	}
	v.hostPerfMetricsByRef = results.resultsByRef
}

// scrapeResourcePools scrapes and stores all relevant property data for a Datacenter's ResourcePools/vApps
func (v *vcenterMetricScraper) scrapeResourcePools(ctx context.Context, dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	v.rPoolsByRef = make(map[string]*mo.ResourcePool)

	// Get ResourcePools/vApps w/properties and store for later retrieval
	rPools, err := v.client.ResourcePools(ctx, dc.Reference())
	if err != nil {
		errs.AddPartial(1, err)
		return
	}
	for i := range rPools {
		v.rPoolsByRef[rPools[i].Reference().Value] = &rPools[i]
	}
}

// scrapeVirtualMachines scrapes and stores all relevant metric/property data for a Datacenter's VirtualMachines
func (v *vcenterMetricScraper) scrapeVirtualMachines(ctx context.Context, dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
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
	results, err := v.client.PerfMetricsQuery(ctx, spec, vmPerfMetricList, vmRefs)
	if err != nil {
		errs.AddPartial(1, fmt.Errorf("failed to retrieve perf metrics for VirtualMachines: %w", err))
		return
	}
	v.vmPerfMetricsByRef = results.resultsByRef
}

// clearScrapeData clears scrape data after it is no longer needed
func (v *vcenterMetricScraper) clearScrapeData() {
	v.computeRefsByDSRef = nil
	v.datacenters = nil
	v.datastores = nil
	v.rPoolIPathsByRef = nil
	v.vAppIPathsByRef = nil
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
	now := pcommon.NewTimestampFromTime(time.Now())

	v.buildDatastores(now, dc, errs)
	v.buildResourcePools(now, dc, errs)
	vmRefToComputeRef := v.buildHosts(now, dc, errs)
	clusterVMStatesByCRRef := v.buildVMs(now, dc, vmRefToComputeRef, errs)
	v.buildClusters(now, dc, clusterVMStatesByCRRef, errs)
}

// buildDatastores builds the Datastore metrics and resources from the collected Datastores
func (v *vcenterMetricScraper) buildDatastores(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	errs *scrapererror.ScrapeErrors,
) {
	for _, ds := range v.datastores {
		v.buildDatastoreMetrics(colTime, dc, ds, errs)
	}
}

// buildDatastoreMetrics builds a resource and metrics for a given collected Datastore
func (v *vcenterMetricScraper) buildDatastoreMetrics(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	ds *mo.Datastore,
	errs *scrapererror.ScrapeErrors,
) {
	// Get a list of ComputeResources for this Datastore if there is one
	crs := []*mo.ComputeResource{}
	uniqueCSRefs := v.computeRefsByDSRef[ds.Reference().Value]
	for crRef, _ := range uniqueCSRefs {
		cr := v.computesByRef[crRef]
		// This shouldn't be possible as dsRefToComputeRefs is populated during collecting ComputeResources, but better be safe
		if cr == nil {
			errs.AddPartial(1, fmt.Errorf("no ComputeResource collected for Datastore [%s]'s ComputeResource ref: %s", ds.Name, crRef))
			continue
		}
		crs = append(crs, cr)
	}

	// Create Datastore resource builder
	rb := v.createDatastoreResourceBuilder(dc, ds, crs)

	// Record Datastore metric data points
	v.recordDatastoreStats(colTime, ds)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))
}

// buildResourcePools builds ResourcePool metrics and resources from the collected ResourcePools
func (v *vcenterMetricScraper) buildResourcePools(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	errs *scrapererror.ScrapeErrors,
) {
	for _, rp := range v.rPoolsByRef {
		// Don't make metrics for vApps
		if rp.Reference().Type == "VirtualApp" {
			continue
		}

		if err := v.buildResourcePoolMetrics(colTime, dc, rp); err != nil {
			errs.AddPartial(1, err)
		}
	}
}

// buildResourcePoolMetrics builds a resource and metrics for a given collected ResourcePool
func (v *vcenterMetricScraper) buildResourcePoolMetrics(colTime pcommon.Timestamp, dc *mo.Datacenter, rp *mo.ResourcePool,
) error {
	// Get related ResourcePool compute info
	crRef := rp.Owner
	cr := v.computesByRef[crRef.Value]
	if cr == nil {
		return fmt.Errorf("no collected ComputeResource found for ResourcePool [%s]'s parent ref: %s", rp.Name, crRef.Value)
	}

	// Create ResourcePool resource builder
	rb, err := v.createResourcePoolResourceBuilder(dc, cr, rp)
	if err != nil {
		return err
	}

	// Record ResourcePool metric data points
	v.recordResourcePoolStats(colTime, rp)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return nil
}

// buildHosts builds Host metrics and resources from the collected Hosts
// returns a map of of each ComputeResource for each VM
func (v *vcenterMetricScraper) buildHosts(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	errs *scrapererror.ScrapeErrors,
) map[string]string {
	vmRefToComputeRef := map[string]string{}

	for _, hs := range v.hostsByRef {
		if err := v.buildHostMetrics(colTime, dc, hs, vmRefToComputeRef); err != nil {
			errs.AddPartial(1, err)
		}
	}

	return vmRefToComputeRef
}

// buildHostMetrics builds a resource and metrics for a given collected Host
func (v *vcenterMetricScraper) buildHostMetrics(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	hs *mo.HostSystem,
	vmRefToComputeRef map[string]string,
) error {
	// Get related Host compute info
	crRef := hs.Parent
	if crRef == nil {
		return fmt.Errorf("no parent found for Host: %s", hs.Name)
	}
	cr := v.computesByRef[crRef.Value]
	if cr == nil {
		return fmt.Errorf("no collected ComputeResource found for Host [%s]'s parent ref: %s", hs.Name, crRef.Value)
	}

	// Store VM to ComputeResource relationship for all child VMs
	for _, vmRef := range hs.Vm {
		vmRefToComputeRef[vmRef.Value] = cr.Reference().Value
	}

	// Create Host resource builder
	rb := v.createHostResourceBuilder(dc, cr, hs)

	// Record Host metric data points
	v.recordHostSystemStats(colTime, hs)
	hostPerfMetrics := v.hostPerfMetricsByRef[hs.Reference().Value]
	if hostPerfMetrics != nil {
		v.recordHostPerformanceMetrics(hostPerfMetrics)
	}
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return nil
}

// buildVMs builds VM metrics and resources from the collected VMs
// returns a map of all VM State counts for each ComputeResource
func (v *vcenterMetricScraper) buildVMs(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	vmRefToComputeRef map[string]string,
	errs *scrapererror.ScrapeErrors,
) map[string]*clusterVMStates {
	vmStatesByCRRef := map[string]*clusterVMStates{}

	for _, vm := range v.vmsByRef {
		if err := v.buildVMMetrics(colTime, dc, vm, vmRefToComputeRef, vmStatesByCRRef); err != nil {
			errs.AddPartial(1, err)
		}
	}

	return vmStatesByCRRef
}

// buildVMMetrics builds a resource and metrics for a given collected VM
func (v *vcenterMetricScraper) buildVMMetrics(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	vm *mo.VirtualMachine,
	vmRefToComputeRef map[string]string,
	vmStatesByCRRef map[string]*clusterVMStates,
) error {
	// Get related VM compute info
	crRef := vmRefToComputeRef[vm.Reference().Value]
	if crRef == "" {
		return fmt.Errorf("no ComputeResource ref found for VM: %s", vm.Name)
	}
	cr := v.computesByRef[crRef]
	if cr == nil {
		return fmt.Errorf("no collected ComputeResource for VM [%s]'s ComputeResource ref: %s", vm.Name, crRef)
	}

	// Get related VM host info
	hsRef := vm.Summary.Runtime.Host
	if hsRef == nil {
		return fmt.Errorf("no Host ref for VM: %s", vm.Name)
	}
	hs := v.hostsByRef[hsRef.Value]
	if hs == nil {
		return fmt.Errorf("no collected Host for VM [%s]'s Host ref: %s", vm.Name, hsRef.Value)
	}

	// VMs may not have a ResourcePool reported (templates)
	// But grab it if available
	rpRef := vm.ResourcePool
	var rp *mo.ResourcePool
	if rpRef != nil {
		rp = v.rPoolsByRef[rpRef.Value]
	}

	// Update cluster's VM power state counts if applicable
	vmStates := vmStatesByCRRef[crRef]
	if vmStates == nil {
		vmStates = &clusterVMStates{poweredOff: 0, poweredOn: 0}
		vmStatesByCRRef[crRef] = vmStates
	}
	if string(vm.Runtime.PowerState) == "poweredOff" {
		vmStates.poweredOff++
	} else {
		vmStates.poweredOn++
	}

	// Create VM resource builder
	rb, err := v.createVMResourceBuilder(dc, cr, hs, rp, vm)
	if err != nil {
		return err
	}

	// Record VM metric data points
	perfMetrics := v.vmPerfMetricsByRef[vm.Reference().Value]
	v.recordVMStats(colTime, vm, hs)
	if perfMetrics != nil {
		v.recordVMPerformanceMetrics(perfMetrics)
	}
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return err
}

// buildClusters builds Cluster metrics and resources from the collected ComputeResources
func (v *vcenterMetricScraper) buildClusters(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	vmStatesByCRRef map[string]*clusterVMStates,
	errs *scrapererror.ScrapeErrors,
) {
	for crRef, cr := range v.computesByRef {
		// Don't make metrics for anything that's not a Cluster (otherwise it should be the same as a HostSystem)
		if cr.Reference().Type != "ClusterComputeResource" {
			continue
		}

		vmStates := vmStatesByCRRef[crRef]
		if err := v.buildClusterMetrics(colTime, dc, cr, vmStates); err != nil {
			errs.AddPartial(1, err)
		}
	}
}

// buildClusterMetrics builds a resource and metrics for a given collected ComputeCluster
func (v *vcenterMetricScraper) buildClusterMetrics(
	colTime pcommon.Timestamp,
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
	vmStates *clusterVMStates,
) (err error) {
	// Create Cluster resource builder
	rb := v.createClusterResourceBuilder(dc, cr)

	if vmStates == nil {
		err = fmt.Errorf("no VM power counts found for Cluster: %s", cr.Name)
	}
	// Record Cluster metric data points
	v.recordClusterStats(colTime, cr, vmStates)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return err
}
