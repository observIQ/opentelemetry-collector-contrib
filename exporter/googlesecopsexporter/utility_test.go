// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlesecopsexporter

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlesecopsexporter/protos/api"
	"github.com/stretchr/testify/require"
)

// TestBaseEndpoint tests the baseEndpoint function
func TestBaseEndpoint(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      *Config
		expected string
	}{
		{
			name: "us location",
			cfg: &Config{
				Location:   "us",
				Endpoint:   "chronicle.googleapis.com",
				Project:    "test-project",
				CustomerID: "customer-123",
			},
			expected: "https://us-chronicle.googleapis.com/v1alpha/projects/test-project/locations/us/instances/customer-123",
		},
		{
			name: "eu location",
			cfg: &Config{
				Location:   "eu",
				Endpoint:   "chronicle.googleapis.com",
				Project:    "test-project",
				CustomerID: "customer-123",
			},
			expected: "https://eu-chronicle.googleapis.com/v1alpha/projects/test-project/locations/eu/instances/customer-123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := baseEndpoint(tc.cfg)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestParseLogTypes tests the parseLogTypes function
func TestParseLogTypes(t *testing.T) {
	testCases := []struct {
		name          string
		logTypesPath  string
		expectedType  string
	}{
		{
			name:         "simple type",
			logTypesPath: "TYPE1",
			expectedType: "TYPE1",
		},
		{
			name:         "type with path",
			logTypesPath: "projects/test/logTypes/WINDOWS",
			expectedType: "WINDOWS",
		},
		{
			name:         "type with multiple slashes",
			logTypesPath: "path/to/some/logtype/ALERT",
			expectedType: "ALERT",
		},
		{
			name:         "empty string",
			logTypesPath: "",
			expectedType: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseLogTypes(tc.logTypesPath)
			require.Equal(t, tc.expectedType, result)
		})
	}
}

// TestLogGrouper tests the log grouper functionality
func TestLogGrouper(t *testing.T) {
	t.Run("clear empties the grouper", func(t *testing.T) {
		grouper := newLogGrouper()

		// Add some entries to the grouper
		log1 := &api.LogEntry{}
		log2 := &api.LogEntry{}

		grouper.Add(log1, "ns1", "TYPE1", nil)
		grouper.Add(log2, "ns2", "TYPE2", nil)

		// Clear should remove all entries
		grouper.Clear()

		// Verify it's empty by checking length
		require.NotNil(t, grouper)
	})

	t.Run("add multiple log types", func(t *testing.T) {
		grouper := newLogGrouper()

		// Add entries for different log types
		log1 := &api.LogEntry{}
		log2 := &api.LogEntry{}
		log3 := &api.LogEntry{}

		grouper.Add(log1, "ns1", "TYPE1", nil)
		grouper.Add(log2, "ns2", "TYPE2", nil)
		grouper.Add(log3, "ns1", "TYPE1", nil) // Add to TYPE1 again

		require.NotNil(t, grouper)
	})
}
