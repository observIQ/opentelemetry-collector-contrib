// Copyright  The OpenTelemetry Authors
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

package varnishreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/varnishreceiver"

import (
	"encoding/json"
	"os/exec"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

// executer executes commands.
type executer interface {
	Execute(command string, args string) ([]byte, error)
}

// varnishExecuter executes varnish commands.
type varnishExecuter struct{}

func newExecuter() executer {
	return &varnishExecuter{}
}

// Execute executes commands with args flag.
func (e *varnishExecuter) Execute(command string, args string) ([]byte, error) {
	return exec.Command(command, args).Output()
}

type version string

// validate validates the version is supported.
func (ve *version) validate() bool {
	return true
}

// client is an interface to get stats and version using an executer.
type client interface {
	GetStats() (*Stats, error)
	GetVersion() (version, error)
}

var _ client = (*varnishClient)(nil)

type varnishClient struct {
	version string
	exec    executer
	cfg     *Config
	logger  *zap.Logger
}

// newVarnishClient creates a client and does a health check via version validation.
// If version is not supported, a log warning is sent.
func newVarnishClient(cfg *Config, host component.Host, settings component.TelemetrySettings) (client, error) {
	vc := varnishClient{
		exec:   newExecuter(),
		cfg:    cfg,
		logger: settings.Logger,
	}

	version, err := vc.GetVersion()
	if err != nil {
		return nil, err
	}

	if !version.validate() {
		vc.logger.Warn("unsupported version", zap.String("version", vc.version))
	}

	return &vc, nil
}

// GetVersion executes and parses the varnish version.
func (v *varnishClient) GetVersion() (version, error) {
	output, err := exec.Command("varnishd", "-V").Output()
	if err != nil {
		v.logger.Error(err.Error())
		return "", err
	}

	return parseVersion(output)
}

// parseVersion parses the varnish version from a varnishd -V response.
func parseVersion(rawVersion []byte) (version, error) {
	return "", nil
}

// GetStats executes and parses the varnish stats.
func (v *varnishClient) GetStats() (*Stats, error) {
	output, err := v.exec.Execute("varnishstat", "-j")
	if err != nil {
		v.logger.Error(err.Error())
		return nil, err
	}

	return parseStats(v.version, output)
}

// parseStats parses varnishStats json response into a Stats struct.
// Version < 6.5 does not have a json response with a nested in a "counter"
// field and therefore is therefore converted to gain parity to 6.5+ Stats.
func parseStats(version string, rawStats []byte) (*Stats, error) {

	// jsonBody := `{
	// 	"version": 1,
	// 	"timestamp": "2022-03-07T09:54:30",
	// 	"counters": {
	// 	   "MAIN.cache_hit": {
	// 		"description": "Cache hits",
	// 		"flag": "c",
	// 		"format": "i",
	// 		"value": 97
	// 	  },
	// 	  "MAIN.cache_hit_grace": {
	// 		"description": "Cache grace hits",
	// 		"flag": "c",
	// 		"format": "i",
	// 		"value": 2
	// 	  },
	// 	  "MAIN.cache_hitpass": {
	// 		"description": "Cache hits for pass.",
	// 		"flag": "c",
	// 		"format": "i",
	// 		"value": 0
	// 	  },
	// 	  "MAIN.cache_hitmiss": {
	// 		"description": "Cache hits for miss.",
	// 		"flag": "c",
	// 		"format": "i",
	// 		"value": 0
	// 	  },
	// 	  "MAIN.cache_miss": {
	// 		"description": "Cache misses",
	// 		"flag": "c",
	// 		"format": "i",
	// 		"value": 60
	// 	  }
	// 	}
	//   }`
	var jsonParsed Stats
	// err := json.Unmarshal([]byte(jsonBody), &jsonParsed)
	err := json.Unmarshal(rawStats, &jsonParsed)
	if err != nil {
		return nil, err
	}
	return &jsonParsed, nil
}
