// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !aix

package config

import "time"

func defaultOrphanDetectionInterval() time.Duration {
	return 5 * time.Second
}

func defaultConfigApplyTimeout() time.Duration {
	return 5 * time.Second
}

func defaultBootstrapTimeout() time.Duration {
	return 3 * time.Second
}
