// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	common "github.com/streamnative/pulsarctl/pkg/pulsar/common"
	mock "github.com/stretchr/testify/mock"

	utils "github.com/streamnative/pulsarctl/pkg/pulsar/utils"
)

// MockNamespaces is an autogenerated mock type for the MockNamespaces type
type MockNamespaces struct {
	mock.Mock
}

// ClearNamespaceBacklog provides a mock function with given fields: namespace
func (_m *MockNamespaces) ClearNamespaceBacklog(namespace utils.NameSpaceName) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ClearNamespaceBacklogForSubscription provides a mock function with given fields: namespace, sName
func (_m *MockNamespaces) ClearNamespaceBacklogForSubscription(namespace utils.NameSpaceName, sName string) error {
	ret := _m.Called(namespace, sName)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, string) error); ok {
		r0 = rf(namespace, sName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ClearNamespaceBundleBacklog provides a mock function with given fields: namespace, bundle
func (_m *MockNamespaces) ClearNamespaceBundleBacklog(namespace utils.NameSpaceName, bundle string) error {
	ret := _m.Called(namespace, bundle)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, string) error); ok {
		r0 = rf(namespace, bundle)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ClearNamespaceBundleBacklogForSubscription provides a mock function with given fields: namespace, bundle, sName
func (_m *MockNamespaces) ClearNamespaceBundleBacklogForSubscription(namespace utils.NameSpaceName, bundle string, sName string) error {
	ret := _m.Called(namespace, bundle, sName)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, string, string) error); ok {
		r0 = rf(namespace, bundle, sName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ClearOffloadDeleteLag provides a mock function with given fields: namespace
func (_m *MockNamespaces) ClearOffloadDeleteLag(namespace utils.NameSpaceName) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateNamespace provides a mock function with given fields: namespace
func (_m *MockNamespaces) CreateNamespace(namespace string) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateNsWithBundlesData provides a mock function with given fields: namespace, bundleData
func (_m *MockNamespaces) CreateNsWithBundlesData(namespace string, bundleData *utils.BundlesData) error {
	ret := _m.Called(namespace, bundleData)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *utils.BundlesData) error); ok {
		r0 = rf(namespace, bundleData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateNsWithNumBundles provides a mock function with given fields: namespace, numBundles
func (_m *MockNamespaces) CreateNsWithNumBundles(namespace string, numBundles int) error {
	ret := _m.Called(namespace, numBundles)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int) error); ok {
		r0 = rf(namespace, numBundles)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateNsWithPolices provides a mock function with given fields: namespace, polices
func (_m *MockNamespaces) CreateNsWithPolices(namespace string, polices utils.Policies) error {
	ret := _m.Called(namespace, polices)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, utils.Policies) error); ok {
		r0 = rf(namespace, polices)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteBookieAffinityGroup provides a mock function with given fields: namespace
func (_m *MockNamespaces) DeleteBookieAffinityGroup(namespace string) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteNamespace provides a mock function with given fields: namespace
func (_m *MockNamespaces) DeleteNamespace(namespace string) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteNamespaceAntiAffinityGroup provides a mock function with given fields: namespace
func (_m *MockNamespaces) DeleteNamespaceAntiAffinityGroup(namespace string) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteNamespaceBundle provides a mock function with given fields: namespace, bundleRange
func (_m *MockNamespaces) DeleteNamespaceBundle(namespace string, bundleRange string) error {
	ret := _m.Called(namespace, bundleRange)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(namespace, bundleRange)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAntiAffinityNamespaces provides a mock function with given fields: tenant, cluster, namespaceAntiAffinityGroup
func (_m *MockNamespaces) GetAntiAffinityNamespaces(tenant string, cluster string, namespaceAntiAffinityGroup string) ([]string, error) {
	ret := _m.Called(tenant, cluster, namespaceAntiAffinityGroup)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, string) ([]string, error)); ok {
		return rf(tenant, cluster, namespaceAntiAffinityGroup)
	}
	if rf, ok := ret.Get(0).(func(string, string, string) []string); ok {
		r0 = rf(tenant, cluster, namespaceAntiAffinityGroup)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(tenant, cluster, namespaceAntiAffinityGroup)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBacklogQuotaMap provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetBacklogQuotaMap(namespace string) (map[utils.BacklogQuotaType]utils.BacklogQuota, error) {
	ret := _m.Called(namespace)

	var r0 map[utils.BacklogQuotaType]utils.BacklogQuota
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (map[utils.BacklogQuotaType]utils.BacklogQuota, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(string) map[utils.BacklogQuotaType]utils.BacklogQuota); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[utils.BacklogQuotaType]utils.BacklogQuota)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBookieAffinityGroup provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetBookieAffinityGroup(namespace string) (*utils.BookieAffinityGroupData, error) {
	ret := _m.Called(namespace)

	var r0 *utils.BookieAffinityGroupData
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*utils.BookieAffinityGroupData, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(string) *utils.BookieAffinityGroupData); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*utils.BookieAffinityGroupData)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCompactionThreshold provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetCompactionThreshold(namespace utils.NameSpaceName) (int64, error) {
	ret := _m.Called(namespace)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (int64, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) int64); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDispatchRate provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetDispatchRate(namespace utils.NameSpaceName) (utils.DispatchRate, error) {
	ret := _m.Called(namespace)

	var r0 utils.DispatchRate
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (utils.DispatchRate, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) utils.DispatchRate); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(utils.DispatchRate)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMaxConsumersPerSubscription provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetMaxConsumersPerSubscription(namespace utils.NameSpaceName) (int, error) {
	ret := _m.Called(namespace)

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (int, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) int); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMaxConsumersPerTopic provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetMaxConsumersPerTopic(namespace utils.NameSpaceName) (int, error) {
	ret := _m.Called(namespace)

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (int, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) int); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMaxProducersPerTopic provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetMaxProducersPerTopic(namespace utils.NameSpaceName) (int, error) {
	ret := _m.Called(namespace)

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (int, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) int); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNamespaceAntiAffinityGroup provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetNamespaceAntiAffinityGroup(namespace string) (string, error) {
	ret := _m.Called(namespace)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNamespaceMessageTTL provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetNamespaceMessageTTL(namespace string) (int, error) {
	ret := _m.Called(namespace)

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (int, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNamespacePermissions provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetNamespacePermissions(namespace utils.NameSpaceName) (map[string][]common.AuthAction, error) {
	ret := _m.Called(namespace)

	var r0 map[string][]common.AuthAction
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (map[string][]common.AuthAction, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) map[string][]common.AuthAction); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]common.AuthAction)
		}
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNamespaceReplicationClusters provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetNamespaceReplicationClusters(namespace string) ([]string, error) {
	ret := _m.Called(namespace)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNamespaces provides a mock function with given fields: tenant
func (_m *MockNamespaces) GetNamespaces(tenant string) ([]string, error) {
	ret := _m.Called(tenant)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(tenant)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(tenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOffloadDeleteLag provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetOffloadDeleteLag(namespace utils.NameSpaceName) (int64, error) {
	ret := _m.Called(namespace)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (int64, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) int64); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOffloadThreshold provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetOffloadThreshold(namespace utils.NameSpaceName) (int64, error) {
	ret := _m.Called(namespace)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (int64, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) int64); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPersistence provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetPersistence(namespace string) (*utils.PersistencePolicies, error) {
	ret := _m.Called(namespace)

	var r0 *utils.PersistencePolicies
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*utils.PersistencePolicies, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(string) *utils.PersistencePolicies); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*utils.PersistencePolicies)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPolicies provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetPolicies(namespace string) (*utils.Policies, error) {
	ret := _m.Called(namespace)

	var r0 *utils.Policies
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*utils.Policies, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(string) *utils.Policies); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*utils.Policies)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPublishRate provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetPublishRate(namespace utils.NameSpaceName) (utils.PublishRate, error) {
	ret := _m.Called(namespace)

	var r0 utils.PublishRate
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (utils.PublishRate, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) utils.PublishRate); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(utils.PublishRate)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetReplicatorDispatchRate provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetReplicatorDispatchRate(namespace utils.NameSpaceName) (utils.DispatchRate, error) {
	ret := _m.Called(namespace)

	var r0 utils.DispatchRate
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (utils.DispatchRate, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) utils.DispatchRate); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(utils.DispatchRate)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRetention provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetRetention(namespace string) (*utils.RetentionPolicies, error) {
	ret := _m.Called(namespace)

	var r0 *utils.RetentionPolicies
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*utils.RetentionPolicies, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(string) *utils.RetentionPolicies); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*utils.RetentionPolicies)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSchemaAutoUpdateCompatibilityStrategy provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetSchemaAutoUpdateCompatibilityStrategy(namespace utils.NameSpaceName) (utils.SchemaCompatibilityStrategy, error) {
	ret := _m.Called(namespace)

	var r0 utils.SchemaCompatibilityStrategy
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (utils.SchemaCompatibilityStrategy, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) utils.SchemaCompatibilityStrategy); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(utils.SchemaCompatibilityStrategy)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSchemaValidationEnforced provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetSchemaValidationEnforced(namespace utils.NameSpaceName) (bool, error) {
	ret := _m.Called(namespace)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (bool, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) bool); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSubscribeRate provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetSubscribeRate(namespace utils.NameSpaceName) (utils.SubscribeRate, error) {
	ret := _m.Called(namespace)

	var r0 utils.SubscribeRate
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (utils.SubscribeRate, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) utils.SubscribeRate); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(utils.SubscribeRate)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSubscriptionDispatchRate provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetSubscriptionDispatchRate(namespace utils.NameSpaceName) (utils.DispatchRate, error) {
	ret := _m.Called(namespace)

	var r0 utils.DispatchRate
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) (utils.DispatchRate, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) utils.DispatchRate); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(utils.DispatchRate)
	}

	if rf, ok := ret.Get(1).(func(utils.NameSpaceName) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTopics provides a mock function with given fields: namespace
func (_m *MockNamespaces) GetTopics(namespace string) ([]string, error) {
	ret := _m.Called(namespace)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(namespace)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GrantNamespacePermission provides a mock function with given fields: namespace, role, action
func (_m *MockNamespaces) GrantNamespacePermission(namespace utils.NameSpaceName, role string, action []common.AuthAction) error {
	ret := _m.Called(namespace, role, action)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, string, []common.AuthAction) error); ok {
		r0 = rf(namespace, role, action)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GrantSubPermission provides a mock function with given fields: namespace, sName, roles
func (_m *MockNamespaces) GrantSubPermission(namespace utils.NameSpaceName, sName string, roles []string) error {
	ret := _m.Called(namespace, sName, roles)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, string, []string) error); ok {
		r0 = rf(namespace, sName, roles)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveBacklogQuota provides a mock function with given fields: namespace
func (_m *MockNamespaces) RemoveBacklogQuota(namespace string) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveTopicAutoCreation provides a mock function with given fields: namespace
func (_m *MockNamespaces) RemoveTopicAutoCreation(namespace utils.NameSpaceName) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RevokeNamespacePermission provides a mock function with given fields: namespace, role
func (_m *MockNamespaces) RevokeNamespacePermission(namespace utils.NameSpaceName, role string) error {
	ret := _m.Called(namespace, role)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, string) error); ok {
		r0 = rf(namespace, role)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RevokeSubPermission provides a mock function with given fields: namespace, sName, role
func (_m *MockNamespaces) RevokeSubPermission(namespace utils.NameSpaceName, sName string, role string) error {
	ret := _m.Called(namespace, sName, role)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, string, string) error); ok {
		r0 = rf(namespace, sName, role)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetBacklogQuota provides a mock function with given fields: namespace, backlogQuota, backlogQuotaType
func (_m *MockNamespaces) SetBacklogQuota(namespace string, backlogQuota utils.BacklogQuota, backlogQuotaType utils.BacklogQuotaType) error {
	ret := _m.Called(namespace, backlogQuota, backlogQuotaType)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, utils.BacklogQuota, utils.BacklogQuotaType) error); ok {
		r0 = rf(namespace, backlogQuota, backlogQuotaType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetBookieAffinityGroup provides a mock function with given fields: namespace, bookieAffinityGroup
func (_m *MockNamespaces) SetBookieAffinityGroup(namespace string, bookieAffinityGroup utils.BookieAffinityGroupData) error {
	ret := _m.Called(namespace, bookieAffinityGroup)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, utils.BookieAffinityGroupData) error); ok {
		r0 = rf(namespace, bookieAffinityGroup)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetCompactionThreshold provides a mock function with given fields: namespace, threshold
func (_m *MockNamespaces) SetCompactionThreshold(namespace utils.NameSpaceName, threshold int64) error {
	ret := _m.Called(namespace, threshold)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, int64) error); ok {
		r0 = rf(namespace, threshold)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetDeduplicationStatus provides a mock function with given fields: namespace, enableDeduplication
func (_m *MockNamespaces) SetDeduplicationStatus(namespace string, enableDeduplication bool) error {
	ret := _m.Called(namespace, enableDeduplication)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, bool) error); ok {
		r0 = rf(namespace, enableDeduplication)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetDispatchRate provides a mock function with given fields: namespace, rate
func (_m *MockNamespaces) SetDispatchRate(namespace utils.NameSpaceName, rate utils.DispatchRate) error {
	ret := _m.Called(namespace, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, utils.DispatchRate) error); ok {
		r0 = rf(namespace, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetEncryptionRequiredStatus provides a mock function with given fields: namespace, encrypt
func (_m *MockNamespaces) SetEncryptionRequiredStatus(namespace utils.NameSpaceName, encrypt bool) error {
	ret := _m.Called(namespace, encrypt)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, bool) error); ok {
		r0 = rf(namespace, encrypt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetMaxConsumersPerSubscription provides a mock function with given fields: namespace, max
func (_m *MockNamespaces) SetMaxConsumersPerSubscription(namespace utils.NameSpaceName, max int) error {
	ret := _m.Called(namespace, max)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, int) error); ok {
		r0 = rf(namespace, max)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetMaxConsumersPerTopic provides a mock function with given fields: namespace, max
func (_m *MockNamespaces) SetMaxConsumersPerTopic(namespace utils.NameSpaceName, max int) error {
	ret := _m.Called(namespace, max)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, int) error); ok {
		r0 = rf(namespace, max)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetMaxProducersPerTopic provides a mock function with given fields: namespace, max
func (_m *MockNamespaces) SetMaxProducersPerTopic(namespace utils.NameSpaceName, max int) error {
	ret := _m.Called(namespace, max)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, int) error); ok {
		r0 = rf(namespace, max)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetNamespaceAntiAffinityGroup provides a mock function with given fields: namespace, namespaceAntiAffinityGroup
func (_m *MockNamespaces) SetNamespaceAntiAffinityGroup(namespace string, namespaceAntiAffinityGroup string) error {
	ret := _m.Called(namespace, namespaceAntiAffinityGroup)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(namespace, namespaceAntiAffinityGroup)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetNamespaceMessageTTL provides a mock function with given fields: namespace, ttlInSeconds
func (_m *MockNamespaces) SetNamespaceMessageTTL(namespace string, ttlInSeconds int) error {
	ret := _m.Called(namespace, ttlInSeconds)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int) error); ok {
		r0 = rf(namespace, ttlInSeconds)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetNamespaceReplicationClusters provides a mock function with given fields: namespace, clusterIds
func (_m *MockNamespaces) SetNamespaceReplicationClusters(namespace string, clusterIds []string) error {
	ret := _m.Called(namespace, clusterIds)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []string) error); ok {
		r0 = rf(namespace, clusterIds)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetOffloadDeleteLag provides a mock function with given fields: namespace, timeMs
func (_m *MockNamespaces) SetOffloadDeleteLag(namespace utils.NameSpaceName, timeMs int64) error {
	ret := _m.Called(namespace, timeMs)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, int64) error); ok {
		r0 = rf(namespace, timeMs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetOffloadThreshold provides a mock function with given fields: namespace, threshold
func (_m *MockNamespaces) SetOffloadThreshold(namespace utils.NameSpaceName, threshold int64) error {
	ret := _m.Called(namespace, threshold)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, int64) error); ok {
		r0 = rf(namespace, threshold)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetPersistence provides a mock function with given fields: namespace, persistence
func (_m *MockNamespaces) SetPersistence(namespace string, persistence utils.PersistencePolicies) error {
	ret := _m.Called(namespace, persistence)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, utils.PersistencePolicies) error); ok {
		r0 = rf(namespace, persistence)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetPublishRate provides a mock function with given fields: namespace, pubRate
func (_m *MockNamespaces) SetPublishRate(namespace utils.NameSpaceName, pubRate utils.PublishRate) error {
	ret := _m.Called(namespace, pubRate)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, utils.PublishRate) error); ok {
		r0 = rf(namespace, pubRate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetReplicatorDispatchRate provides a mock function with given fields: namespace, rate
func (_m *MockNamespaces) SetReplicatorDispatchRate(namespace utils.NameSpaceName, rate utils.DispatchRate) error {
	ret := _m.Called(namespace, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, utils.DispatchRate) error); ok {
		r0 = rf(namespace, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetRetention provides a mock function with given fields: namespace, policy
func (_m *MockNamespaces) SetRetention(namespace string, policy utils.RetentionPolicies) error {
	ret := _m.Called(namespace, policy)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, utils.RetentionPolicies) error); ok {
		r0 = rf(namespace, policy)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetSchemaAutoUpdateCompatibilityStrategy provides a mock function with given fields: namespace, strategy
func (_m *MockNamespaces) SetSchemaAutoUpdateCompatibilityStrategy(namespace utils.NameSpaceName, strategy utils.SchemaCompatibilityStrategy) error {
	ret := _m.Called(namespace, strategy)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, utils.SchemaCompatibilityStrategy) error); ok {
		r0 = rf(namespace, strategy)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetSchemaValidationEnforced provides a mock function with given fields: namespace, schemaValidationEnforced
func (_m *MockNamespaces) SetSchemaValidationEnforced(namespace utils.NameSpaceName, schemaValidationEnforced bool) error {
	ret := _m.Called(namespace, schemaValidationEnforced)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, bool) error); ok {
		r0 = rf(namespace, schemaValidationEnforced)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetSubscribeRate provides a mock function with given fields: namespace, rate
func (_m *MockNamespaces) SetSubscribeRate(namespace utils.NameSpaceName, rate utils.SubscribeRate) error {
	ret := _m.Called(namespace, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, utils.SubscribeRate) error); ok {
		r0 = rf(namespace, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetSubscriptionAuthMode provides a mock function with given fields: namespace, mode
func (_m *MockNamespaces) SetSubscriptionAuthMode(namespace utils.NameSpaceName, mode utils.SubscriptionAuthMode) error {
	ret := _m.Called(namespace, mode)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, utils.SubscriptionAuthMode) error); ok {
		r0 = rf(namespace, mode)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetSubscriptionDispatchRate provides a mock function with given fields: namespace, rate
func (_m *MockNamespaces) SetSubscriptionDispatchRate(namespace utils.NameSpaceName, rate utils.DispatchRate) error {
	ret := _m.Called(namespace, rate)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, utils.DispatchRate) error); ok {
		r0 = rf(namespace, rate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTopicAutoCreation provides a mock function with given fields: namespace, config
func (_m *MockNamespaces) SetTopicAutoCreation(namespace utils.NameSpaceName, config utils.TopicAutoCreationConfig) error {
	ret := _m.Called(namespace, config)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, utils.TopicAutoCreationConfig) error); ok {
		r0 = rf(namespace, config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SplitNamespaceBundle provides a mock function with given fields: namespace, bundle, unloadSplitBundles
func (_m *MockNamespaces) SplitNamespaceBundle(namespace string, bundle string, unloadSplitBundles bool) error {
	ret := _m.Called(namespace, bundle, unloadSplitBundles)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, bool) error); ok {
		r0 = rf(namespace, bundle, unloadSplitBundles)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Unload provides a mock function with given fields: namespace
func (_m *MockNamespaces) Unload(namespace string) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UnloadNamespaceBundle provides a mock function with given fields: namespace, bundle
func (_m *MockNamespaces) UnloadNamespaceBundle(namespace string, bundle string) error {
	ret := _m.Called(namespace, bundle)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(namespace, bundle)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UnsubscribeNamespace provides a mock function with given fields: namespace, sName
func (_m *MockNamespaces) UnsubscribeNamespace(namespace utils.NameSpaceName, sName string) error {
	ret := _m.Called(namespace, sName)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, string) error); ok {
		r0 = rf(namespace, sName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UnsubscribeNamespaceBundle provides a mock function with given fields: namespace, bundle, sName
func (_m *MockNamespaces) UnsubscribeNamespaceBundle(namespace utils.NameSpaceName, bundle string, sName string) error {
	ret := _m.Called(namespace, bundle, sName)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.NameSpaceName, string, string) error); ok {
		r0 = rf(namespace, bundle, sName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockNamespaces interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockNamespaces creates a new instance of MockNamespaces. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockNamespaces(t mockConstructorTestingTNewMockNamespaces) *MockNamespaces {
	mock := &MockNamespaces{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
