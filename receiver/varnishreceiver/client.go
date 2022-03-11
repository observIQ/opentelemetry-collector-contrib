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

type executer interface {
	Execute(command string, args string) ([]byte, error)
}

type varnishExecuter struct{}

func newExecuter() executer {
	return &varnishExecuter{}
}

// Execute implementation.
func (e *varnishExecuter) Execute(command string, args string) ([]byte, error) {
	return exec.Command(command, args).Output()
}

type version string

// validate implementation.
func (ve *version) validate() bool {
	return true
}

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
		vc.logger.Warn("not supported version")
	}

	return &vc, nil
}

// need to create a get function interface to mock

// GetVersion implementation.
func (v *varnishClient) GetVersion() (version, error) {
	// output, err := exec.Command("varnishd", "-V").Output()
	// if err != nil {
	// 	v.logger.Error(err.Error(), zap.String("try executeing with elevated privileges", "sudo ./bin/otelcontrib_OS_VERSION --config config.yaml"))
	// 	return "", err
	// }

	// version, err := v.parseVersion(output)
	// if err != nil {
	// 	return err
	// }

	// if v.validateVersion(version) {
	// 	v.logger.Warn(zap.String("warning", ""))
	// }
	return "", nil
}

// GetStats executes the varnish command to collect json formated stats.
func (v *varnishClient) GetStats() (*Stats, error) {
	output, err := v.exec.Execute("varnishstat", "-j")
	if err != nil {
		v.logger.Error(err.Error())
		return nil, err
	}

	return parseStats(v.version, output)
}

// // GetStats executes the varnish command to collect json formated stats.
// func (v *varnishClient) GetStats() (*Stats, error) {

// 	output, err := exec.Command("varnishstat", "-j").Output()
// 	if err != nil {
// 		v.logger.Error(err.Error(), zap.String("try executeing with elevated privileges", "sudo ./bin/otelcontrib_OS_VERSION --config config.yaml"))
// 		return nil, err
// 	}

// 	return parseStats(output)
// }

// if version < 6.5 put into lower version then parse it into new version

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
