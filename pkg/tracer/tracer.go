package tracer

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// Tracer is the main tracer
var Tracer trace.Tracer

// Context is the main tracer context
var Context context.Context
