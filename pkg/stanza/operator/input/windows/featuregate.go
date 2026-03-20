// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package windows // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/windows"

import "go.opentelemetry.io/collector/featuregate"

// EventDrivenScrapingGate enables event-driven scraping for the Windows Event Log receiver.
// When enabled, the receiver blocks on a Windows API signal rather than polling on a fixed
// interval, reducing latency and avoiding unnecessary wakeups between events.
var EventDrivenScrapingGate = featuregate.GlobalRegistry().MustRegister(
	"receiver.windowseventlog.eventDrivenScraping",
	featuregate.StageAlpha,
	featuregate.WithRegisterDescription("When enabled, the Windows Event Log receiver wakes on API signals instead of polling on a fixed interval."),
)
