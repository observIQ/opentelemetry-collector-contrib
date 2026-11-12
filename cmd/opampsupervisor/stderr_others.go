// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !aix

package main

// redirectDiscardedStderr is a no-op outside AIX, where the service manager keeps stderr.
func redirectDiscardedStderr(string) error {
	return nil
}
