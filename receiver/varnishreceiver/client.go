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
	Execute(command string, args ...string) ([]byte, error)
}

type client interface {
	GetStats() (*Stats, error)
}

// type client interface {
// 	GetStats() (*Stats, error)
// }

var _ client = (*varnishClient)(nil)

type varnishClient struct {
	cfg    *Config
	logger *zap.Logger
}

func newVarnishClient(cfg *Config, host component.Host, settings component.TelemetrySettings) client {
	// check version here
	return &varnishClient{
		cfg:    cfg,
		logger: settings.Logger,
	}
}

// GetStats executes the varnish command to collect json formated stats.
func (v *varnishClient) GetStats() (*Stats, error) {

	output, err := exec.Command("varnishstat", "-j").Output()
	if err != nil {
		v.logger.Error(err.Error(), zap.String("try executeing with elevated privileges", "sudo ./bin/otelcontrib_OS_VERSION --config config.yaml"))
		return nil, err
	}

	return parseStats(output)
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

func parseStats(rawStats []byte) (*Stats, error) {

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
