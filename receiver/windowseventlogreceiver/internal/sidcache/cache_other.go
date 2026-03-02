// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package sidcache

import "errors"

// New returns an error on non-Windows platforms since SID resolution
// requires the Windows Local Security Authority API.
func New(_ Config) (Cache, error) {
	return nil, errors.New("SID cache is only supported on Windows")
}
