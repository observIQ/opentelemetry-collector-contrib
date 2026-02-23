// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build aix

package commander

import (
	"os"
	"syscall"
	"time"
)

func oneShotTimeout() time.Duration {
	return 30 * time.Second
}

func sendShutdownSignal(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}

func sysProcAttrs() *syscall.SysProcAttr {
	// On non-windows systems, no extra attributes are needed.
	return nil
}
