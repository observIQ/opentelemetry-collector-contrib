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
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/scrapererror"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver/internal/model"
)

// aerospikeReceiver is a metrics receiver using the Aerospike interface to collect
type aerospikeReceiver struct {
	config        *Config
	consumer      consumer.Metrics
	host          string // host/IP of configured Aerospike node
	port          int    // port of configured Aerospike node
	clientFactory clientFactoryFunc
	mb            *metadata.MetricsBuilder
	logger        *zap.Logger
}

type clientFactoryFunc func(host string, port int, username, password string, timeout time.Duration) (aerospike, error)

// newAerospikeReceiver creates a new aerospikeReceiver connected to the endpoint provided in cfg
//
// If the host or port can't be parsed from endpoint, an error is returned.
func newAerospikeReceiver(params component.ReceiverCreateSettings, cfg *Config, consumer consumer.Metrics) (*aerospikeReceiver, error) {
	host, portStr, err := net.SplitHostPort(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errBadEndpoint, err)
	}

	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errBadPort, err)
	}

	return &aerospikeReceiver{
		logger:   params.Logger,
		config:   cfg,
		consumer: consumer,
		clientFactory: func(host string, port int, username, password string, timeout time.Duration) (aerospike, error) {
			return newASClient(host, port, username, password, timeout)
		},
		host: host,
		port: int(port),
		mb:   metadata.NewMetricsBuilder(cfg.Metrics, params.BuildInfo),
	}, nil
}

// scrape scrapes both Node and Namespace metrics from the provided Aerospike node.
// If CollectClusterMetrics is true, it then scrapes every discovered node
func (r *aerospikeReceiver) scrape(ctx context.Context) (pmetric.Metrics, error) {
	errs := &scrapererror.ScrapeErrors{}
	now := pcommon.NewTimestampFromTime(time.Now().UTC())
	client, err := r.clientFactory(r.host, r.port, r.config.Username, r.config.Password, r.config.Timeout)
	if err != nil {
		return pmetric.NewMetrics(), fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	info, err := client.Info()
	if err != nil {
		r.logger.Warn(fmt.Sprintf("failed to get INFO: %s", err.Error()))
		return r.mb.Emit(), err
	}
	r.emitNode(info, now)
	r.scrapeNamespaces(info, client, now, errs)

	if r.config.CollectClusterMetrics {
		r.logger.Debug("Collecting peer nodes")
		for _, n := range info.Services {
			r.scrapeDiscoveredNode(n, now, errs)
		}
	}

	return r.mb.Emit(), errs.Combine()
}

// scrapeNode collects metrics from a single Aerospike node
func (r *aerospikeReceiver) scrapeNode(client aerospike, now pcommon.Timestamp, errs *scrapererror.ScrapeErrors) {
	info, err := client.Info()
	if err != nil {
		errs.AddPartial(0, err)
		return
	}

	r.emitNode(info, now)
	r.scrapeNamespaces(info, client, now, errs)
}

// scrapeDiscoveredNode connects to a discovered Aerospike node and scrapes it using that connection
//
// If unable to parse the endpoint or connect, that error is logged and we return early
func (r *aerospikeReceiver) scrapeDiscoveredNode(endpoint string, now pcommon.Timestamp, errs *scrapererror.ScrapeErrors) {
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		r.logger.Warn(fmt.Sprintf("%s: %s", errBadEndpoint, err))
		errs.Add(err)
		return
	}
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		r.logger.Warn(fmt.Sprintf("%s: %s", errBadPort, err))
		errs.Add(err)
		return
	}

	nClient, err := r.clientFactory(host, int(port), r.config.Username, r.config.Password, r.config.Timeout)
	if err != nil {
		r.logger.Warn(err.Error())
		errs.Add(err)
		return
	}
	defer nClient.Close()

	r.scrapeNode(nClient, now, errs)
}

// emitNode records node metrics and emits the resource. If statistics are missing in INFO, nothing is recorded
func (r *aerospikeReceiver) emitNode(info *model.NodeInfo, now pcommon.Timestamp) {
	stats := info.Statistics
	if stats == nil {
		return
	}

	if stats.ClientConnections != nil {
		r.mb.RecordAerospikeNodeConnectionOpenDataPoint(now, *stats.ClientConnections, metadata.AttributeConnectionTypeClient)
	}
	if stats.FabricConnections != nil {
		r.mb.RecordAerospikeNodeConnectionOpenDataPoint(now, *stats.FabricConnections, metadata.AttributeConnectionTypeFabric)
	}
	if stats.HeartbeatConnections != nil {
		r.mb.RecordAerospikeNodeConnectionOpenDataPoint(now, *stats.HeartbeatConnections, metadata.AttributeConnectionTypeHeartbeat)
	}

	if stats.ClientConnectionsClosed != nil {
		r.mb.RecordAerospikeNodeConnectionCountDataPoint(now, *stats.ClientConnectionsClosed, metadata.AttributeConnectionTypeClient, metadata.AttributeConnectionOpClose)
	}
	if stats.ClientConnectionsOpened != nil {
		r.mb.RecordAerospikeNodeConnectionCountDataPoint(now, *stats.ClientConnectionsOpened, metadata.AttributeConnectionTypeClient, metadata.AttributeConnectionOpOpen)
	}
	if stats.FabricConnectionsClosed != nil {
		r.mb.RecordAerospikeNodeConnectionCountDataPoint(now, *stats.FabricConnectionsClosed, metadata.AttributeConnectionTypeFabric, metadata.AttributeConnectionOpClose)
	}
	if stats.FabricConnectionsOpened != nil {
		r.mb.RecordAerospikeNodeConnectionCountDataPoint(now, *stats.FabricConnectionsOpened, metadata.AttributeConnectionTypeFabric, metadata.AttributeConnectionOpOpen)
	}
	if stats.HeartbeatConnectionsClosed != nil {
		r.mb.RecordAerospikeNodeConnectionCountDataPoint(now, *stats.HeartbeatConnectionsClosed, metadata.AttributeConnectionTypeHeartbeat, metadata.AttributeConnectionOpClose)
	}
	if stats.HeartbeatConnectionsOpened != nil {
		r.mb.RecordAerospikeNodeConnectionCountDataPoint(now, *stats.HeartbeatConnectionsOpened, metadata.AttributeConnectionTypeHeartbeat, metadata.AttributeConnectionOpOpen)
	}

	r.mb.EmitForResource(metadata.WithNodeName(info.Name))
}

// scrapeNamespaces records metrics for all namespaces on a node
// The given client is used to collect namespace metrics, which is connected to a single node
func (r *aerospikeReceiver) scrapeNamespaces(info *model.NodeInfo, client aerospike, now pcommon.Timestamp, errs *scrapererror.ScrapeErrors) {
	if info.Namespaces == nil {
		return
	}

	for _, n := range info.Namespaces {
		nInfo, err := client.NamespaceInfo(n)
		if err != nil {
			r.logger.Warn(fmt.Sprintf("failed getting namespace %s: %s", n, err.Error()))
			errs.AddPartial(0, err)
			continue
		}
		nInfo.Node = info.Name
		r.emitNamespace(nInfo, now)
	}
}

// emitNamespace emits a namespace resource with its name as resource attribute
func (r *aerospikeReceiver) emitNamespace(info *model.NamespaceInfo, now pcommon.Timestamp) {
	if info.DeviceAvailablePct != nil {
		r.mb.RecordAerospikeNamespaceDiskAvailableDataPoint(now, *info.DeviceAvailablePct)
	}
	if info.MemoryFreePct != nil {
		r.mb.RecordAerospikeNamespaceMemoryFreeDataPoint(now, *info.MemoryFreePct)
	}
	if info.MemoryUsedDataBytes != nil {
		r.mb.RecordAerospikeNamespaceMemoryUsageDataPoint(now, *info.MemoryUsedDataBytes, metadata.AttributeNamespaceComponentData)
	}
	if info.MemoryUsedIndexBytes != nil {
		r.mb.RecordAerospikeNamespaceMemoryUsageDataPoint(now, *info.MemoryUsedIndexBytes, metadata.AttributeNamespaceComponentIndex)
	}
	if info.MemoryUsedSIndexBytes != nil {
		r.mb.RecordAerospikeNamespaceMemoryUsageDataPoint(now, *info.MemoryUsedSIndexBytes, metadata.AttributeNamespaceComponentSindex)
	}
	if info.MemoryUsedSetIndexBytes != nil {
		r.mb.RecordAerospikeNamespaceMemoryUsageDataPoint(now, *info.MemoryUsedSetIndexBytes, metadata.AttributeNamespaceComponentSetIndex)
	}

	if info.ScanAggrAbort != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanAggrAbort, metadata.AttributeScanTypeAggr, metadata.AttributeScanResultAbort)
	}
	if info.ScanAggrComplete != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanAggrComplete, metadata.AttributeScanTypeAggr, metadata.AttributeScanResultComplete)
	}
	if info.ScanAggrError != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanAggrError, metadata.AttributeScanTypeAggr, metadata.AttributeScanResultError)
	}

	if info.ScanBasicAbort != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanBasicAbort, metadata.AttributeScanTypeBasic, metadata.AttributeScanResultAbort)
	}
	if info.ScanBasicComplete != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanBasicComplete, metadata.AttributeScanTypeBasic, metadata.AttributeScanResultComplete)
	}
	if info.ScanBasicError != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanBasicError, metadata.AttributeScanTypeBasic, metadata.AttributeScanResultError)
	}

	if info.ScanOpsBgAbort != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanOpsBgAbort, metadata.AttributeScanTypeOpsBg, metadata.AttributeScanResultAbort)
	}
	if info.ScanOpsBgComplete != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanOpsBgComplete, metadata.AttributeScanTypeOpsBg, metadata.AttributeScanResultComplete)
	}
	if info.ScanOpsBgError != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanOpsBgError, metadata.AttributeScanTypeOpsBg, metadata.AttributeScanResultError)
	}

	if info.ScanUdfBgAbort != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanUdfBgAbort, metadata.AttributeScanTypeUdfBg, metadata.AttributeScanResultAbort)
	}
	if info.ScanUdfBgComplete != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanUdfBgComplete, metadata.AttributeScanTypeUdfBg, metadata.AttributeScanResultComplete)
	}
	if info.ScanUdfBgError != nil {
		r.mb.RecordAerospikeNamespaceScanCountDataPoint(now, *info.ScanUdfBgError, metadata.AttributeScanTypeUdfBg, metadata.AttributeScanResultError)
	}
	r.mb.EmitForResource(metadata.WithAerospikeNamespace(info.Name), metadata.WithNodeName(info.Node))
}
