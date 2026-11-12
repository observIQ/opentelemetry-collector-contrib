// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// stderrFileName is the supervisor's stderr capture file, written next to its config.
const stderrFileName = "supervisor.err"

// stderrPath returns the stderr capture file for the supervisor config at configPath.
func stderrPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), stderrFileName)
}

// rollStderrFile moves a non-empty stderr file to "<path>.1", so the previous
// run's output survives a restart.
func rollStderrFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat stderr file: %w", err)
	}
	if info.Size() == 0 {
		return nil
	}
	if err := os.Rename(path, path+".1"); err != nil {
		return fmt.Errorf("roll stderr file: %w", err)
	}
	return nil
}

// capStderrFile copies the stderr file to "<path>.1" and truncates it in place
// once it reaches threshold bytes. Truncating keeps the open stderr handle
// valid, since it appends at EOF.
func capStderrFile(path string, threshold int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat stderr file: %w", err)
	}
	if info.Size() < threshold {
		return nil
	}

	src, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("open stderr file: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(filepath.Clean(path+".1"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open stderr backup file: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return fmt.Errorf("copy stderr file to backup: %w", err)
	}
	if err := dst.Close(); err != nil {
		return fmt.Errorf("close stderr backup file: %w", err)
	}
	if err := os.Truncate(path, 0); err != nil {
		return fmt.Errorf("truncate stderr file: %w", err)
	}
	return nil
}
