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

// Program otelcontribcol is an extension to the OpenTelemetry Collector
// that includes additional components, some vendor-specific, contributed
// from the wider community.

//go:build !testbed
// +build !testbed

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/components"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelcontribcore"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/tracer"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	fmt.Println(projectID)
	exporter, err := trace.New(trace.WithProjectID(projectID))
	if err != nil {
		log.Fatalf("texporter.NewExporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	defer tp.ForceFlush(ctx)
	otel.SetTracerProvider(tp)

	tracer.Tracer = otel.GetTracerProvider().Tracer("otel-collector")
	spanCtx, span := tracer.Tracer.Start(ctx, "main")
	tracer.Context = spanCtx
	defer span.End()

	otelcontribcore.RunWithComponents(components.Components)
}
