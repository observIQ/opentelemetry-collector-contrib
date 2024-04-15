// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceBuilder(t *testing.T) {
	for _, test := range []string{"default", "all_set", "none_set"} {
		t.Run(test, func(t *testing.T) {
			cfg := loadResourceAttributesConfig(t, test)
			rb := NewResourceBuilder(cfg)
			rb.SetVcenterClusterName("vcenter.cluster.name-val")
			rb.SetVcenterClusterNames([]any{"vcenter.cluster.names-item1", "vcenter.cluster.names-item2"})
			rb.SetVcenterDatacenterName("vcenter.datacenter.name-val")
			rb.SetVcenterDatastoreName("vcenter.datastore.name-val")
			rb.SetVcenterHostName("vcenter.host.name-val")
			rb.SetVcenterResourcePoolInventoryPath("vcenter.resource_pool.inventory_path-val")
			rb.SetVcenterResourcePoolName("vcenter.resource_pool.name-val")
			rb.SetVcenterVirtualAppInventoryPath("vcenter.virtual_app.inventory_path-val")
			rb.SetVcenterVirtualAppName("vcenter.virtual_app.name-val")
			rb.SetVcenterVMID("vcenter.vm.id-val")
			rb.SetVcenterVMName("vcenter.vm.name-val")

			res := rb.Emit()
			assert.Equal(t, 0, rb.Emit().Attributes().Len()) // Second call should return empty Resource

			switch test {
			case "default":
				assert.Equal(t, 11, res.Attributes().Len())
			case "all_set":
				assert.Equal(t, 11, res.Attributes().Len())
			case "none_set":
				assert.Equal(t, 0, res.Attributes().Len())
				return
			default:
				assert.Failf(t, "unexpected test case: %s", test)
			}

			val, ok := res.Attributes().Get("vcenter.cluster.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "vcenter.cluster.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("vcenter.cluster.names")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, []any{"vcenter.cluster.names-item1", "vcenter.cluster.names-item2"}, val.Slice().AsRaw())
			}
			val, ok = res.Attributes().Get("vcenter.datacenter.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "vcenter.datacenter.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("vcenter.datastore.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "vcenter.datastore.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("vcenter.host.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "vcenter.host.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("vcenter.resource_pool.inventory_path")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "vcenter.resource_pool.inventory_path-val", val.Str())
			}
			val, ok = res.Attributes().Get("vcenter.resource_pool.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "vcenter.resource_pool.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("vcenter.virtual_app.inventory_path")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "vcenter.virtual_app.inventory_path-val", val.Str())
			}
			val, ok = res.Attributes().Get("vcenter.virtual_app.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "vcenter.virtual_app.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("vcenter.vm.id")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "vcenter.vm.id-val", val.Str())
			}
			val, ok = res.Attributes().Get("vcenter.vm.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "vcenter.vm.name-val", val.Str())
			}
		})
	}
}
