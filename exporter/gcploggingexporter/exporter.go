// Copyright 2019, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gcploggingexporter

import (
	"context"
	"fmt"

	vkit "cloud.google.com/go/logging/apiv2"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/pdata"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type gcpLoggingExporter struct {
	done chan struct{}

	projectID   string
	credentials *google.Credentials

	client *vkit.Client
}

// Ensure this exporter adheres to required interface
var _ component.LogsExporter = (*gcpLoggingExporter)(nil)

func (e *gcpLoggingExporter) Start(ctx context.Context, host component.Host) error {
	options := []option.ClientOption{option.WithCredentials(e.credentials)}
	client, err := vkit.NewClient(ctx, options...)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	e.client = client
	return nil
}

// Shutdown is invoked during service shutdown
func (e *gcpLoggingExporter) Shutdown(_ context.Context) error {
	close(e.done)
	return nil
}

func (e *gcpLoggingExporter) ConsumeLogs(ctx context.Context, ld pdata.Logs) error {

	return nil
}
