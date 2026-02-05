// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlesecopsexporter

import (
	"net"
	"strings"
)

const unknownValue = "unknown"

// getMACAddress returns the MAC address for the host.
// It attempts to find the first valid network interface with an IPv4 address.
func getMACAddress() string {
	return findMACAddress(net.Interfaces)
}

// findMACAddress does its best to find the MAC address for the local network interface.
func findMACAddress(interfaces func() ([]net.Interface, error)) string {
	iFaces, err := interfaces()
	if err != nil {
		return unknownValue
	}

	for _, iFace := range iFaces {
		if iFace.HardwareAddr.String() != "" {
			addrs, _ := iFace.Addrs()
			for _, addr := range addrs {
				address := addr.String()
				if strings.Contains(address, "/") {
					address = address[:strings.Index(address, "/")]
				}
				ipAddress := net.ParseIP(address)
				if isValidV4Address(ipAddress) {
					return iFace.HardwareAddr.String()
				}
			}
		}
	}

	return unknownValue
}

// isValidV4Address checks that the IP address is not nil, loopback, or unspecified. It also checks that it is IPv4
func isValidV4Address(address net.IP) bool {
	return address != nil && !address.IsLoopback() && !address.IsUnspecified() && address.To4() != nil
}
