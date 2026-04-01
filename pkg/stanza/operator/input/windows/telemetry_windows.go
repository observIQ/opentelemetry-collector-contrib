// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package windows // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/windows"

import "context"

func (noopWindowsInputTelemetry) RecordMissedEvents(_ context.Context, _ string, _ int64) {}
