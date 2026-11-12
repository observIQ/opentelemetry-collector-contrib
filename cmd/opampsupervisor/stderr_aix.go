// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build aix

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

// The stderr file catches output that skips the logger, like runtime panics and
// goroutine dumps. These bound it to stderrMaxBytes plus one backup.
const (
	stderrMaxBytes      = 10 * 1024 * 1024
	stderrCheckInterval = 10 * time.Second
)

// redirectDiscardedStderr sends stderr to supervisor.err next to the config when
// AIX SRC leaves it on /dev/console. SRC's own mkssys -e is no substitute: SRC
// truncates that file on every start, including the restart after a crash.
func redirectDiscardedStderr(configPath string) error {
	if configPath == "" || !stderrDiscarded() {
		return nil
	}
	path := stderrPath(configPath)
	if err := rollStderrFile(path); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Clean(path), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("open stderr file: %w", err)
	}
	// dup2 points fd 2 at the file, so runtime panics written straight to fd 2 land in it too
	if err := unix.Dup2(int(f.Fd()), 2); err != nil {
		return fmt.Errorf("redirect stderr: %w (close err: %w)", err, f.Close())
	}
	os.Stderr = f
	go watchStderrFile(path)
	return nil
}

// stderrDiscarded reports whether stderr is /dev/console or /dev/null, where nothing keeps it.
func stderrDiscarded() bool {
	stderrInfo, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	for _, dev := range []string{"/dev/console", "/dev/null"} {
		if info, err := os.Stat(dev); err == nil && os.SameFile(info, stderrInfo) {
			return true
		}
	}
	return false
}

// watchStderrFile caps the stderr file for the life of the process. A failed
// check is retried on the next tick, since stderr is the file being managed.
func watchStderrFile(path string) {
	ticker := time.NewTicker(stderrCheckInterval)
	defer ticker.Stop()
	for range ticker.C {
		_ = capStderrFile(path, stderrMaxBytes)
	}
}
