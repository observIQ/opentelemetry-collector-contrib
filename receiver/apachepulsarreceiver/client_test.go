// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apachepulsarreceiver

import (
	"errors"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachepulsarreceiver/internal/mocks"
	pulsarctl "github.com/streamnative/pulsarctl/pkg/pulsar"
	"github.com/streamnative/pulsarctl/pkg/pulsar/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	tenantsResponseFile = "get_tenants_response.json"
	URL                 = "http://www.pulsarmock.com"
)

func TestGetTenants(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "Returns an error on Tenants().List() call error",
			testFunc: func(t *testing.T) {
				// mock Tenant and Pulsar package client interfaces
				mockTenant := &mocks.MockTenants{}
				// Return an error when List() is called on the mock tenant
				mockTenant.On("List").Return(nil, errors.New("error occurred while fetching list of tenant names"))
				// Tenants() returns a dummy
				mockClient := &mocks.MockClient{}
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
				mockClient := &mocks.MockClient{}
				mockClient.On("Tenants").Return(mockTenant)

				tc := createTestClient(t, mockClient, URL)
				tenantNames, err := tc.GetTenants()
				require.Nil(t, err)
				require.NotNil(t, tenantNames)
				// TODO: figure out how to best handle empty responses on the client side
			},
		},
		{
			desc: "Success case",
			testFunc: func(t *testing.T) {

				// mock Tenant and Pulsar package client interfaces
				mockTenant := &mocks.MockTenants{}
				// Return list of tenants when List() is called on the mock tenant
				mockTenant.On("List").Return([]string{"public", "pulsar"}, nil)
				mockClient := &mocks.MockClient{}
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
				mockClient := &mocks.MockClient{}
				mockClient.On("Namespaces").Return(mockNamespace)

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
				mockClient := &mocks.MockClient{}
				mockClient.On("Namespaces").Return(mockNamespace)

				tc := createTestClient(t, mockClient, URL)
				namespaces, err := tc.GetNameSpaces([]string{"public"})
				require.Nil(t, err)
				require.NotNil(t, namespaces)
			},
		},
		{
			desc: "Handles the case where tenants parameter is empty",
			testFunc: func(t *testing.T) {
				mockClient := &mocks.MockClient{}

				tc := createTestClient(t, mockClient, URL)
				namespaces, err := tc.GetNameSpaces([]string{})
				require.NotNil(t, err)
				require.EqualError(t, err, `cannot perform this operation on an empty array`)
				require.Nil(t, namespaces)
			},
		},
		{
			desc: "Success case",
			testFunc: func(t *testing.T) {

				returnedNamespaces := []string{"Hello", "World"}
				mockNamespace := &mocks.MockNamespaces{}
				mockNamespace.On("GetNamespaces", mock.Anything).Return(returnedNamespaces, nil)
				mockClient := &mocks.MockClient{}
				mockClient.On("Namespaces").Return(mockNamespace)

				tc := createTestClient(t, mockClient, URL)
				namespaces, err := tc.GetNameSpaces([]string{"public"})
				require.Nil(t, err)
				require.NotNil(t, namespaces)
				require.Equal(t, namespaces, returnedNamespaces)
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
	}{
		{
			desc: "Returns an error when GetTopics() fails",
			testFunc: func(t *testing.T) {
				mockNamespace := &mocks.MockNamespaces{}
				mockNamespace.On("GetTopics", mock.Anything).Return(nil, errors.New(`Failed to get topics for that namespace`))
				mockClient := &mocks.MockClient{}
				mockClient.On("Namespaces").Return(mockNamespace)

				tc := createTestClient(t, mockClient, URL)
				topics, err := tc.GetTopics([]string{"namespace1"})
				require.NotNil(t, err)
				require.Nil(t, topics)
			},
		},
		{
			desc: "Handles the case where topics result is empty",
			testFunc: func(t *testing.T) {
				mockNamespace := &mocks.MockNamespaces{}
				mockNamespace.On("GetTopics", mock.Anything).Return([]string{}, nil)
				mockClient := &mocks.MockClient{}
				mockClient.On("Namespaces").Return(mockNamespace)

				tc := createTestClient(t, mockClient, URL)
				topics, err := tc.GetTopics([]string{"namespace1"})
				require.Nil(t, err)
				require.NotNil(t, topics)
				// TODO: how to handle this case
			},
		},
		{
			desc: "Handles the case where namespace parameter is empty",
			testFunc: func(t *testing.T) {
				mockClient := &mocks.MockClient{}

				tc := createTestClient(t, mockClient, URL)
				topics, err := tc.GetTopics([]string{})
				require.NotNil(t, err)
				require.EqualError(t, err, `cannot perform this operation on an empty array`)
				require.Nil(t, topics)
			},
		},
		{
			desc: "Success case",
			testFunc: func(t *testing.T) {
				mockNamespace := &mocks.MockNamespaces{}
				mockNamespace.On("GetTopics", mock.Anything).Return([]string{"my-topic"}, nil)
				mockClient := &mocks.MockClient{}
				mockClient.On("Namespaces").Return(mockNamespace)

				tc := createTestClient(t, mockClient, URL)
				topics, err := tc.GetTopics([]string{"default"})
				require.Nil(t, err)
				require.NotNil(t, topics)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

func TestGetTopicStats(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "Returns an error when GetStats() fails",
			testFunc: func(t *testing.T) {
				mockTopic := &mocks.MockTopics{}
				mockTopic.On("GetStats", mock.Anything).Return(utils.TopicStats{}, errTopicStatsNotFound)
				mockClient := &mocks.MockClient{}
				mockClient.On("Topics").Return(mockTopic)

				tc := createTestClient(t, mockClient, URL)
				stats, err := tc.GetTopicStats([]string{"namespace1"})
				require.NotNil(t, err)
				require.EqualError(t, err, `error occurred while fetching stats for a topic`)
				require.Nil(t, stats)
			},
		},
		{
			desc: "Returns an error when GetTopicName() takes in a bad value",
			testFunc: func(t *testing.T) {
				mockTopic := &mocks.MockTopics{}

				mockClient := &mocks.MockClient{}
				mockClient.On("Topics").Return(mockTopic)

				tc := createTestClient(t, mockClient, URL)
				stats, err := tc.GetTopicStats([]string{"!fjnaiubgidf;uhgsuihtgsl!!!!!{}{}{}{}{}{}{}fkdgjsghibhusiuht/irgjirgj"})
				require.NotNil(t, err)
				require.Nil(t, stats)
			},
		},
		{
			desc: "Handles empty param case",
			testFunc: func(t *testing.T) {
				mockTopic := &mocks.MockTopics{}
				mockTopic.On("GetStats", mock.Anything).Return(utils.TopicStats{}, errors.New(`error occurred while fetching stats for a topic`))
				mockClient := &mocks.MockClient{}
				mockClient.On("Topics").Return(mockTopic)

				tc := createTestClient(t, mockClient, URL)
				stats, err := tc.GetTopicStats([]string{})
				require.NotNil(t, err)
				require.EqualError(t, err, `cannot perform this operation on an empty array`)
				require.Nil(t, stats)
			},
		},
		{
			desc: "Success case",
			testFunc: func(t *testing.T) {
				mockTopic := &mocks.MockTopics{}
				mockTopic.On("GetStats", mock.Anything).Return(utils.TopicStats{}, nil)
				mockClient := &mocks.MockClient{}
				mockClient.On("Topics").Return(mockTopic)

				tc := createTestClient(t, mockClient, URL)
				stats, err := tc.GetTopicStats([]string{"my-topic"})
				require.Nil(t, err)
				require.NotNil(t, stats)
			},
		},
	}

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
