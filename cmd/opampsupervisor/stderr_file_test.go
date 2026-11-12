// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStderrPath(t *testing.T) {
	require.Equal(t, filepath.Join("/opt/my dir", "supervisor.err"), stderrPath("/opt/my dir/supervisor.yaml"))
}

func TestRollStderrFile(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "supervisor.err")
		require.NoError(t, rollStderrFile(path))
		require.NoFileExists(t, path+".1")
	})

	t.Run("empty file is left in place", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "supervisor.err")
		require.NoError(t, os.WriteFile(path, nil, 0o600))
		require.NoError(t, rollStderrFile(path))
		require.FileExists(t, path)
		require.NoFileExists(t, path+".1")
	})

	t.Run("previous output moves to the backup", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "supervisor.err")
		require.NoError(t, os.WriteFile(path, []byte("panic"), 0o600))
		require.NoError(t, os.WriteFile(path+".1", []byte("older"), 0o600))
		require.NoError(t, rollStderrFile(path))
		require.NoFileExists(t, path)
		got, err := os.ReadFile(path + ".1")
		require.NoError(t, err)
		require.Equal(t, "panic", string(got))
	})
}

func TestCapStderrFile(t *testing.T) {
	t.Run("under the threshold", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "supervisor.err")
		require.NoError(t, os.WriteFile(path, []byte("small"), 0o600))
		require.NoError(t, capStderrFile(path, 10))
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, "small", string(got))
		require.NoFileExists(t, path+".1")
	})

	t.Run("at the threshold the contents move to the backup", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "supervisor.err")
		require.NoError(t, os.WriteFile(path, []byte("0123456789"), 0o600))
		require.NoError(t, capStderrFile(path, 10))
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Empty(t, got)
		backup, err := os.ReadFile(path + ".1")
		require.NoError(t, err)
		require.Equal(t, "0123456789", string(backup))
	})
}
