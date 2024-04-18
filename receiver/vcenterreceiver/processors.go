// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package vcenterreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"

import (
	"fmt"
	"time"

	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/receiver/scrapererror"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver/internal/metadata"
)

// processDatacenterData creates all of the vCenter metrics from the stored scraped data under a single Datacenter
func (v *vcenterMetricScraper) processDatacenterData(dc *mo.Datacenter, errs *scrapererror.ScrapeErrors) {
	// Init for current collection
	now := pcommon.NewTimestampFromTime(time.Now())

	v.processDatastores(now, dc, errs)
	v.processResourcePools(now, dc, errs)
	vmRefToComputeRef := v.processHosts(now, dc, errs)
	vmStateInfoByComputeRef := v.processVMs(now, dc, vmRefToComputeRef, errs)
	v.processClusters(now, dc, vmStateInfoByComputeRef, errs)
}

// processDatastores creates the vCenter Datastore metrics and resources from the stored scraped data under a single Datacenter
func (v *vcenterMetricScraper) processDatastores(
	ts pcommon.Timestamp,
	dc *mo.Datacenter,
	errs *scrapererror.ScrapeErrors,
) {
	for _, ds := range v.scrapeData.datastores {
		v.buildDatastoreMetrics(ts, dc, ds, errs)
	}
}

// buildDatastoreMetrics builds a resource and metrics for a given scraped vCenter Datastore
func (v *vcenterMetricScraper) buildDatastoreMetrics(
	ts pcommon.Timestamp,
	dc *mo.Datacenter,
	ds *mo.Datastore,
	errs *scrapererror.ScrapeErrors,
) {
	// Get a list of ComputeResources for this Datastore if there is one
	computes := []*mo.ComputeResource{}
	uniqueComputeRefs := v.scrapeData.computeRefsByDSRef[ds.Reference().Value]
	for crRef, _ := range uniqueComputeRefs {
		cr := v.scrapeData.computesByRef[crRef]
		// This shouldn't be possible as computeRefsByDSRef is populated during collecting ComputeResources, but better be safe
		if cr == nil {
			errs.AddPartial(1, fmt.Errorf("no ComputeResource collected for Datastore [%s]'s ComputeResource ref: %s", ds.Name, crRef))
			continue
		}
		computes = append(computes, cr)
	}

	// Create Datastore resource builder
	rb := v.createDatastoreResourceBuilder(dc, ds, computes)

	// Record & emit Datastore metric data points
	v.recordDatastoreStats(ts, ds)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))
}

// processResourcePools creates the vCenter Resource Pool metrics and resources from the stored scraped data under a single Datacenter
func (v *vcenterMetricScraper) processResourcePools(
	ts pcommon.Timestamp,
	dc *mo.Datacenter,
	errs *scrapererror.ScrapeErrors,
) {
	for _, rp := range v.scrapeData.rPoolsByRef {
		// Don't make metrics for vApps
		if rp.Reference().Type == "VirtualApp" {
			continue
		}

		if err := v.buildResourcePoolMetrics(ts, dc, rp); err != nil {
			errs.AddPartial(1, err)
		}
	}
}

// buildResourcePoolMetrics builds a resource and metrics for a given scraped vCenter Resource Pool
func (v *vcenterMetricScraper) buildResourcePoolMetrics(
	ts pcommon.Timestamp,
	dc *mo.Datacenter,
	rp *mo.ResourcePool,
) error {
	// Get related ResourcePool Compute info
	crRef := rp.Owner
	cr := v.scrapeData.computesByRef[crRef.Value]
	if cr == nil {
		return fmt.Errorf("no collected ComputeResource found for ResourcePool [%s]'s parent ref: %s", rp.Name, crRef.Value)
	}

	// Create ResourcePool resource builder
	rb, err := v.createResourcePoolResourceBuilder(dc, cr, rp)
	if err != nil {
		return err
	}

	// Record & emit Resource Pool metric data points
	v.recordResourcePoolStats(ts, rp)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return nil
}

// processHosts creates the vCenter HostS metrics and resources from the stored scraped data under a single Datacenter
//
// returns a map containing the ComputeResource info for each VM
func (v *vcenterMetricScraper) processHosts(
	ts pcommon.Timestamp,
	dc *mo.Datacenter,
	errs *scrapererror.ScrapeErrors,
) map[string]*types.ManagedObjectReference {
	vmRefToComputeRef := map[string]*types.ManagedObjectReference{}

	for _, hs := range v.scrapeData.hostsByRef {
		hsVMRefToComputeRef, err := v.buildHostMetrics(ts, dc, hs)
		if err != nil {
			errs.AddPartial(1, err)
			continue
		}

		// Populate master VM to CR relationship map from
		// single Host based version of it
		for vmRef, csRef := range hsVMRefToComputeRef {
			vmRefToComputeRef[vmRef] = csRef
		}
	}

	return vmRefToComputeRef
}

// buildHostMetrics builds a resource and metrics for a given scraped Host
//
// returns a map containing the ComputeResource info for each VM running on the Host
func (v *vcenterMetricScraper) buildHostMetrics(
	ts pcommon.Timestamp,
	dc *mo.Datacenter,
	hs *mo.HostSystem,
) (vmRefToComputeRef map[string]*types.ManagedObjectReference, err error) {
	vmRefToComputeRef = map[string]*types.ManagedObjectReference{}
	// Get related Host ComputeResource info
	crRef := hs.Parent
	if crRef == nil {
		return vmRefToComputeRef, fmt.Errorf("no parent found for Host: %s", hs.Name)
	}
	cr := v.scrapeData.computesByRef[crRef.Value]
	if cr == nil {
		return vmRefToComputeRef, fmt.Errorf("no collected ComputeResource found for Host [%s]'s parent ref: %s", hs.Name, crRef.Value)
	}

	// Store VM to ComputeResource relationship for all child VMs
	for _, vmRef := range hs.Vm {
		vmRefToComputeRef[vmRef.Value] = crRef
	}

	// Create Host resource builder
	rb := v.createHostResourceBuilder(dc, cr, hs)

	// Record & emit Host metric data points
	v.recordHostSystemStats(ts, hs)
	hostPerfMetrics := v.scrapeData.hostPerfMetricsByRef[hs.Reference().Value]
	if hostPerfMetrics != nil {
		v.recordHostPerformanceMetrics(hostPerfMetrics)
	}
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return vmRefToComputeRef, nil
}

// processVMs creates the vCenter VM metrics and resources from the stored scraped data under a single Datacenter
//
// returns a map of all VM State counts for each ComputeResource
func (v *vcenterMetricScraper) processVMs(
	ts pcommon.Timestamp,
	dc *mo.Datacenter,
	vmRefToComputeRef map[string]*types.ManagedObjectReference,
	errs *scrapererror.ScrapeErrors,
) map[string]*vmStateInfo {
	vmStateInfoByComputeRef := map[string]*vmStateInfo{}

	for _, vm := range v.scrapeData.vmsByRef {
		crRef, singleVMStateInfo, err := v.buildVMMetrics(ts, dc, vm, vmRefToComputeRef)
		if err != nil {
			errs.AddPartial(1, err)
		}

		// Update master ComputeResource VM power state counts with VM power info
		if crRef != nil && singleVMStateInfo != nil {
			crVMStateInfo := vmStateInfoByComputeRef[crRef.Value]
			if crVMStateInfo == nil {
				crVMStateInfo = &vmStateInfo{poweredOff: 0, poweredOn: 0}
				vmStateInfoByComputeRef[crRef.Value] = crVMStateInfo
			}
			if singleVMStateInfo.poweredOn > 0 {
				crVMStateInfo.poweredOn++
			}
			if singleVMStateInfo.poweredOff > 0 {
				crVMStateInfo.poweredOff++
			}
		}
	}

	return vmStateInfoByComputeRef
}

// buildVMMetrics builds a resource and metrics for a given scraped VM
//
// returns the ComputeResource and power state info associated with this VM
func (v *vcenterMetricScraper) buildVMMetrics(
	ts pcommon.Timestamp,
	dc *mo.Datacenter,
	vm *mo.VirtualMachine,
	vmRefToComputeRef map[string]*types.ManagedObjectReference,
) (crRef *types.ManagedObjectReference, stateInfo *vmStateInfo, err error) {
	// Get related VM compute info
	crRef = vmRefToComputeRef[vm.Reference().Value]
	if crRef == nil {
		return crRef, stateInfo, fmt.Errorf("no ComputeResource ref found for VM: %s", vm.Name)
	}
	cr := v.scrapeData.computesByRef[crRef.Value]
	if cr == nil {
		return crRef, stateInfo, fmt.Errorf("no collected ComputeResource for VM [%s]'s ComputeResource ref: %s", vm.Name, crRef)
	}

	// Get related VM host info
	hsRef := vm.Summary.Runtime.Host
	if hsRef == nil {
		return crRef, stateInfo, fmt.Errorf("no Host ref for VM: %s", vm.Name)
	}
	hs := v.scrapeData.hostsByRef[hsRef.Value]
	if hs == nil {
		return crRef, stateInfo, fmt.Errorf("no collected Host for VM [%s]'s Host ref: %s", vm.Name, hsRef.Value)
	}

	// VMs may not have a ResourcePool reported (templates)
	// But grab it if available
	rpRef := vm.ResourcePool
	var rp *mo.ResourcePool
	if rpRef != nil {
		rp = v.scrapeData.rPoolsByRef[rpRef.Value]
	}

	// Update VM power info
	stateInfo = &vmStateInfo{poweredOff: 0, poweredOn: 0}
	if string(vm.Runtime.PowerState) == "poweredOff" {
		stateInfo.poweredOff++
	} else {
		stateInfo.poweredOn++
	}

	// Create VM resource builder
	rb, err := v.createVMResourceBuilder(dc, cr, hs, rp, vm)
	if err != nil {
		return crRef, stateInfo, err
	}

	// Record VM metric data points
	perfMetrics := v.scrapeData.vmPerfMetricsByRef[vm.Reference().Value]
	v.recordVMStats(ts, vm, hs)
	if perfMetrics != nil {
		v.recordVMPerformanceMetrics(perfMetrics)
	}
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return crRef, stateInfo, err
}

// buildClusters creates the vCenter Cluster metrics and resources from the stored scraped data under a single Datacenter
func (v *vcenterMetricScraper) processClusters(
	ts pcommon.Timestamp,
	dc *mo.Datacenter,
	vmStatesByComputeRef map[string]*vmStateInfo,
	errs *scrapererror.ScrapeErrors,
) {
	for crRef, cr := range v.scrapeData.computesByRef {
		// Don't make metrics for anything that's not a Cluster (otherwise it should be the same as a HostSystem)
		if cr.Reference().Type != "ClusterComputeResource" {
			continue
		}

		vmStateInfo := vmStatesByComputeRef[crRef]
		if err := v.buildClusterMetrics(ts, dc, cr, vmStateInfo); err != nil {
			errs.AddPartial(1, err)
		}
	}
}

// buildClusterMetrics builds a resource and metrics for a given scraped Cluster
func (v *vcenterMetricScraper) buildClusterMetrics(
	ts pcommon.Timestamp,
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
	vmStateInfo *vmStateInfo,
) (err error) {
	// Create Cluster resource builder
	rb := v.createClusterResourceBuilder(dc, cr)

	if vmStateInfo == nil {
		err = fmt.Errorf("no VM power counts found for Cluster: %s", cr.Name)
	}
	// Record and emit Cluster metric data points
	v.recordClusterStats(ts, cr, vmStateInfo)
	v.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return err
}
