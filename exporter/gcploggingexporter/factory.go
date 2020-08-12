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
	"io/ioutil"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"golang.org/x/oauth2/google"
)

const (
	// The value of "type" key in configuration.
	typeStr      = "gcplogging"
	loggingWrite = "https://www.googleapis.com/auth/logging.write"
)

// Factory is the factory for the gcplogging exporter.
type Factory struct {
}

// Ensure this factory adheres to required interface
var _ component.LogsExporterFactory = (*Factory)(nil)

// Type returns "gcplogging"
func (f *Factory) Type() configmodels.Type {
	return configmodels.Type(typeStr)
}

// CustomUnmarshaler returns nil
func (f *Factory) CustomUnmarshaler() component.CustomUnmarshaler {
	return nil
}

// CreateDefaultConfig creates the default configuration for the exporter.
func (f *Factory) CreateDefaultConfig() configmodels.Exporter {
	return &Config{
		ExporterSettings: configmodels.ExporterSettings{
			TypeVal: configmodels.Type(typeStr),
			NameVal: typeStr,
		},
	}
}

// CreateLogsExporter creates a logs exporter based on this config.
func (f *Factory) CreateLogsExporter(
	ctx context.Context,
	params component.ExporterCreateParams,
	cfg configmodels.Exporter,
) (component.LogsExporter, error) {

	eCfg := cfg.(*Config)

	ex := gcpLoggingExporter{
		projectID: eCfg.ProjectID,
	}

	if eCfg.CredentialsFile != "" {
		credentialsBytes, err := ioutil.ReadFile(eCfg.CredentialsFile)
		if err != nil {
			return nil, fmt.Errorf("read credentials file: %s", err)
		}
		ex.credentials, err = google.CredentialsFromJSON(context.Background(), credentialsBytes, loggingWrite)
		if err != nil {
			return nil, fmt.Errorf("parse credentials: %s", err)
		}
	} else {
		var err error
		ex.credentials, err = google.FindDefaultCredentials(ctx, loggingWrite)
		if err != nil {
			return nil, fmt.Errorf("no credentials found: %s", err)
		}
	}

	if ex.projectID == "" {
		if ex.credentials.ProjectID == "" {
			return nil, fmt.Errorf("project id empty")
		}
		ex.projectID = ex.credentials.ProjectID
	}

	return &ex, nil
}
