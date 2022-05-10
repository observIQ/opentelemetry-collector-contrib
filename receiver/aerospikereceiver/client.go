// Copyright 2022, OpenTelemetry Authors
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

package aerospikereceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver"

import (
	"time"

	as "github.com/aerospike/aerospike-client-go/v5"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver/internal/model"
)

type defaultASClient struct {
	conn    *as.Connection
	timeout time.Duration
}

// Aerospike is the interface that provides information about a given node
type Aerospike interface {
	// NamespaceInfo gets information about a specific namespace
	NamespaceInfo(namespace string) (*model.NamespaceInfo, error)
	// Info gets high-level information about the node/system.
	Info() (*model.NodeInfo, error)
	// Close closes the connection to the Aerospike node
	Close()
}

// newASClient creates a new defaultASClient connected to the given host and port
func newASClient(host string, port int, timeout time.Duration) (*defaultASClient, error) {
	policy := as.NewClientPolicy()
	policy.Timeout = timeout

	conn, err := as.NewConnection(policy, as.NewHost(host, port))
	if err != nil {
		return nil, err.Unwrap()
	}

	return &defaultASClient{
		conn:    conn,
		timeout: timeout,
	}, nil
}

func (c *defaultASClient) NamespaceInfo(namespace string) (*model.NamespaceInfo, error) {
	c.conn.SetTimeout(time.Now().Add(c.timeout), c.timeout)
	var response model.InfoResponse
	response, err := c.conn.RequestInfo(model.NamespaceKey(namespace))
	if err != nil {
		return nil, err
	}

	return model.ParseNamespaceInfo(response, namespace), nil
}

func (c *defaultASClient) Info() (*model.NodeInfo, error) {
	c.conn.SetTimeout(time.Now().Add(c.timeout), c.timeout)
	var response model.InfoResponse
	response, err := c.conn.RequestInfo("namespaces", "node", "statistics", "services")
	if err != nil {
		return nil, err.Unwrap()
	}

	return model.ParseInfo(response), nil
}

func (c *defaultASClient) Close() {
	c.conn.Close()
}
