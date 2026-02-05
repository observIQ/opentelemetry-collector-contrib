// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlesecopsexporter

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidV4Address(t *testing.T) {
	testCases := []struct {
		name     string
		address  net.IP
		expected bool
	}{
		{
			name:     "valid IPv4",
			address:  net.IPv4(192, 168, 1, 1),
			expected: true,
		},
		{
			name:     "valid public IPv4",
			address:  net.IPv4(1, 1, 1, 1),
			expected: true,
		},
		{
			name:     "loopback IPv4",
			address:  net.IPv4(127, 0, 0, 1),
			expected: false,
		},
		{
			name:     "IPv6",
			address:  net.IPv6loopback,
			expected: false,
		},
		{
			name:     "unspecified",
			address:  net.IPv4zero,
			expected: false,
		},
		{
			name:     "nil",
			address:  nil,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid := isValidV4Address(tc.address)
			require.Equal(t, tc.expected, valid)
		})
	}
}

func TestFindMACAddress(t *testing.T) {
	testCases := []struct {
		name       string
		interfaces func() ([]net.Interface, error)
		expected   string
	}{
		{
			name: "Failed to get interfaces",
			interfaces: func() ([]net.Interface, error) {
				return nil, errors.New("failure")
			},
			expected: unknownValue,
		},
		{
			name: "No interfaces",
			interfaces: func() ([]net.Interface, error) {
				return []net.Interface{}, nil
			},
			expected: unknownValue,
		},
		{
			name: "Interface without MAC",
			interfaces: func() ([]net.Interface, error) {
				return []net.Interface{
					{
						Name: "lo",
					},
				}, nil
			},
			expected: unknownValue,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mac := findMACAddress(tc.interfaces)
			require.Equal(t, tc.expected, mac)
		})
	}
}

func TestGetMACAddress(t *testing.T) {
	// This test verifies that getMACAddress returns a value (either a valid MAC or "unknown")
	// We don't assert a specific value since it depends on the host's network configuration
	mac := getMACAddress()
	require.NotEmpty(t, mac)
}
