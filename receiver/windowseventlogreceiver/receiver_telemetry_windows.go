// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package windowseventlogreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver"

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver/internal/metadata"
)

type receiverWindowsTelemetry struct {
	tb *metadata.TelemetryBuilder
}

func (r *receiverWindowsTelemetry) RecordEventSize(ctx context.Context, channel string, sizeBytes int) {
	r.tb.ReceiverWindowsEventLogEventSize.Record(ctx, int64(sizeBytes),
		metric.WithAttributes(attribute.String("channel", channel)))
}

func (r *receiverWindowsTelemetry) RecordChannelSize(ctx context.Context, channel string, size int64) {
	r.tb.ReceiverWindowsEventLogChannelSize.Record(ctx, size,
		metric.WithAttributes(attribute.String("channel", channel)))
}

func (r *receiverWindowsTelemetry) RecordMissedEvents(ctx context.Context, channel string, count int64) {
	r.tb.ReceiverWindowsEventLogMissedEvents.Add(ctx, count,
		metric.WithAttributes(attribute.String("channel", channel)))
}

func (r *receiverWindowsTelemetry) RecordBatchSize(ctx context.Context, channel string, count int64) {
	r.tb.ReceiverWindowsEventLogBatchSize.Record(ctx, count,
		metric.WithAttributes(attribute.String("channel", channel)))
}
