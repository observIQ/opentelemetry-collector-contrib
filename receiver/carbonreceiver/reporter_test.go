// Copyright 2019 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// nolint:errcheck
package carbonreceiver

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/obsreport/obsreporttest"
)

func TestReporterObservability(t *testing.T) {
	tt, err := obsreporttest.SetupTelemetry()
	require.NoError(t, err)
	defer tt.Shutdown(context.Background())

	receiverID := config.NewComponentIDWithName(typeStr, "fake_receiver")
	reporter := newReporter(receiverID, tt.ToReceiverCreateSettings())

	ctx := reporter.OnDataReceived(context.Background())

	reporter.OnMetricsProcessed(ctx, 17, nil)

	require.NoError(t, obsreporttest.CheckReceiverMetrics(tt, receiverID, "tcp", 17, 0))

	// Below just exercise the error paths.
	err = errors.New("fake error for tests")
	reporter.OnTranslationError(ctx, err)
	reporter.OnMetricsProcessed(ctx, 10, err)

	require.NoError(t, obsreporttest.CheckReceiverMetrics(tt, receiverID, "tcp", 17, 10))
}
