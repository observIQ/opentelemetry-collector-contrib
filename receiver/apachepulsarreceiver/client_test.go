package apachepulsarreceiver

import (
	"errors"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachepulsarreceiver/internal/mocks"
	pulsarctl "github.com/streamnative/pulsarctl/pkg/pulsar"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	tenantsResponseFile = "get_tenants_response.json"
	URL                 = "http://www.pulsarmock.com"
)

var (
	mockClient = &mocks.MockClient{}
)

func TestGetTenants(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		// TODO: make small iterations on error test case for other cases
		{
			desc: "Returns an error on Tenants().List() call error",
			testFunc: func(t *testing.T) {
				// mock Tenant and Pulsar package client interfaces
				mockTenant := &mocks.MockTenants{}
				// Return an error when List() is called on the mock tenant
				mockTenant.On("List").Return(nil, errors.New("error occurred while fetching list of tenant names"))
				// Tenants() returns a dummy
				mockClient.On("Tenants").Return(mockTenant)
				tc := createTestClient(t, mockClient, URL)
				tenantNames, err := tc.GetTenants()
				require.NotNil(t, err)
				require.Nil(t, tenantNames)
			},
		},
		{
			desc: "Handles case where result is empty",
			testFunc: func(t *testing.T) {

				// mock Tenant and Pulsar package client interfaces
				mockTenant := &mocks.MockTenants{}
				// Return an error when List() is called on the mock tenant
				mockTenant.On("List").Return([]string{}, nil)
				mockClient.On("Tenants").Return(mockTenant)

				tc := createTestClient(t, mockClient, URL)
				tenantNames, err := tc.GetTenants()
				require.Nil(t, err)
				require.NotNil(t, tenantNames)
				// TODO: figure out how to best handle empty responses
			},
		},
		{
			desc: "Success case",
			testFunc: func(t *testing.T) {

				// mock Tenant and Pulsar package client interfaces
				mockTenant := &mocks.MockTenants{}
				// Return list of tenants when List() is called on the mock tenant
				mockTenant.On("List").Return([]string{"public", "pulsar"}, nil)
				mockClient.On("Tenants").Return(mockTenant)

				var expected = []string{"public", "pulsar"}
				tc := createTestClient(t, mockClient, URL)
				tenantNames, err := tc.GetTenants()
				require.NoError(t, err)
				require.Nil(t, err)
				require.Equal(t, expected, tenantNames)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

func TestGetNamespaces(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "Returns an error when GetNamespaces() fails",
			testFunc: func(t *testing.T) {
				mockNamespace := &mocks.MockNamespaces{}
				mockNamespace.On("GetNamespaces", mock.Anything).Return(nil, errors.New(`Failed to get namespaces for that tenant`))

				tc := createTestClient(t, mockClient, URL)
				namespaces, err := tc.GetNameSpaces([]string{"tenant1"})
				require.NotNil(t, err)
				require.Nil(t, namespaces)
			},
		},
		{
			desc: "Handles the case where namespace result is empty",
			testFunc: func(t *testing.T) {
				mockNamespace := &mocks.MockNamespaces{}
				mockNamespace.On("GetNamespaces", mock.Anything).Return([]string{}, nil)

				tc := createTestClient(t, mockClient, URL)
				namespaces, err := tc.GetNameSpaces([]string{"tenant1"})
				require.Nil(t, err)
				require.NotNil(t, namespaces)
			},
		},
		{
			desc: "Handles the case where tenants parameter is empty",
			// TODO: implement this case on client side
			testFunc: func(t *testing.T) {

				tc := createTestClient(t, mockClient, URL)
				namespaces, err := tc.GetNameSpaces([]string{})
				require.NotNil(t, err)
				require.EqualError(t, errEmptyParam, `cannot perform this operation on an empty array`)
				require.Nil(t, namespaces)
			},
		},
		{
			desc: "Returns a number of namespaces >= the number of tenants passed in",
			// TODO: finish this test case, if it's worth testing
			// not a perfect metric for the function's success
			testFunc: func(t *testing.T) {
				mockNamespace := &mocks.MockNamespaces{}
				mockNamespace.On("GetNamespaces", mock.AnythingOfType("slice")).Return([]string{}, nil)

				tc := createTestClient(t, mockClient, URL)
				namespaces, err := tc.GetNameSpaces([]string{"tenant1"})
				require.Nil(t, err)
				require.NotNil(t, namespaces)
			},
		},
		{
			desc: "Success case",
			// TODO: start/finish this test case
			testFunc: func(t *testing.T) {
				mockNamespace := &mocks.MockNamespaces{}
				mockNamespace.On("GetNamespaces", mock.AnythingOfType("slice")).Return([]string{}, nil)

				tc := createTestClient(t, mockClient, URL)
				namespaces, err := tc.GetNameSpaces([]string{"tenant1"})
				require.Nil(t, err)
				require.NotNil(t, namespaces)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

func TestGetTopics(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{}

	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

func TestGetStats(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{}

	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

func createTestClient(t *testing.T, client pulsarctl.Client, baseEndpoint string) client {
	t.Helper()
	testClient := &apachePulsarClient{
		client:       client,
		hostEndpoint: baseEndpoint,
		logger:       zap.NewNop(),
	}
	return testClient
}
