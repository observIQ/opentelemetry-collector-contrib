// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package vcenterreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"

import (
	"fmt"

	"github.com/vmware/govmomi/vim25/mo"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver/internal/metadata"
)

func (v *vcenterMetricScraper) createDatastoreResourceBuilder(
	dc *mo.Datacenter,
	ds *mo.Datastore,
	crs []*mo.ComputeResource,
) *metadata.ResourceBuilder {
	rb := v.mb.NewResourceBuilder()
	rb.SetVcenterDatacenterName(dc.Name)
	rb.SetVcenterDatastoreName(ds.Name)

	cNames := []any{}
	for _, cr := range crs {
		if cr.Reference().Type == "ClusterComputeResource" {
			cNames = append(cNames, cr.Name)
		}
	}
	rb.SetVcenterClusterNames(cNames)

	return rb
}

func (v *vcenterMetricScraper) createClusterResourceBuilder(
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
) *metadata.ResourceBuilder {
	rb := v.mb.NewResourceBuilder()
	rb.SetVcenterDatacenterName(dc.Name)
	rb.SetVcenterClusterName(cr.Name)

	return rb
}

func (v *vcenterMetricScraper) createHostResourceBuilder(
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
	hs *mo.HostSystem,
) *metadata.ResourceBuilder {
	rb := v.mb.NewResourceBuilder()
	rb.SetVcenterDatacenterName(dc.Name)
	if cr.Reference().Type == "ClusterComputeResource" {
		rb.SetVcenterClusterName(cr.Name)
	}
	rb.SetVcenterHostName(hs.Name)

	return rb
}

func (v *vcenterMetricScraper) createResourcePoolResourceBuilder(
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
	rp *mo.ResourcePool,
) (*metadata.ResourceBuilder, error) {
	rb := v.mb.NewResourceBuilder()
	rb.SetVcenterDatacenterName(dc.Name)
	if cr.Reference().Type == "ClusterComputeResource" {
		rb.SetVcenterClusterName(cr.Name)
	}
	rb.SetVcenterResourcePoolName(rp.Name)

	iPath := v.rPoolIPathsByRef[rp.Reference().Value]
	if iPath == nil {
		return nil, fmt.Errorf("no inventory path found for collected ResourcePool: %s", rp.Name)
	}
	rb.SetVcenterResourcePoolInventoryPath(*iPath)

	return rb, nil
}

func (v *vcenterMetricScraper) createVMResourceBuilder(
	dc *mo.Datacenter,
	cr *mo.ComputeResource,
	hs *mo.HostSystem,
	rp *mo.ResourcePool,
	vm *mo.VirtualMachine,
) (*metadata.ResourceBuilder, error) {
	rb := v.mb.NewResourceBuilder()
	rb.SetVcenterDatacenterName(dc.Name)
	if cr.Reference().Type == "ClusterComputeResource" {
		rb.SetVcenterClusterName(cr.Name)
	}
	rb.SetVcenterHostName(hs.Name)

	// Likely a VM Template
	if rp == nil {
		rb.SetVcenterVMName(vm.Name)
		rb.SetVcenterVMID(vm.Config.InstanceUuid)

		return rb, nil
	}

	if rp.Reference().Type == "VirtualApp" {
		rb.SetVcenterVirtualAppName(rp.Name)
		iPath := v.vAppIPathsByRef[rp.Reference().Value]
		if iPath == nil {
			return nil, fmt.Errorf("no inventory path found for VM [%s]'s collected vApp: %s", vm.Name, rp.Name)
		}
		rb.SetVcenterVirtualAppInventoryPath(*iPath)
	} else {
		rb.SetVcenterResourcePoolName(rp.Name)
		iPath := v.rPoolIPathsByRef[rp.Reference().Value]
		if iPath == nil {
			return nil, fmt.Errorf("no inventory path found for VM [%s]'s collected ResourcePool: %s", vm.Name, rp.Name)
		}
		rb.SetVcenterResourcePoolInventoryPath(*iPath)
	}
	rb.SetVcenterVMName(vm.Name)
	rb.SetVcenterVMID(vm.Config.InstanceUuid)

	return rb, nil
}
