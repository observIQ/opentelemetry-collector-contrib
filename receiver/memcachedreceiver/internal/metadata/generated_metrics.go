// Copyright The OpenTelemetry Authors
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

// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/model/pdata"
)

// Type is the component type name.
const Type config.Type = "memcachedreceiver"

// MetricIntf is an interface to generically interact with generated metric.
type MetricIntf interface {
	Name() string
	New() pdata.Metric
	Init(metric pdata.Metric)
}

// Intentionally not exposing this so that it is opaque and can change freely.
type metricImpl struct {
	name     string
	initFunc func(pdata.Metric)
}

// Name returns the metric name.
func (m *metricImpl) Name() string {
	return m.name
}

// New creates a metric object preinitialized.
func (m *metricImpl) New() pdata.Metric {
	metric := pdata.NewMetric()
	m.Init(metric)
	return metric
}

// Init initializes the provided metric object.
func (m *metricImpl) Init(metric pdata.Metric) {
	m.initFunc(metric)
}

type metricStruct struct {
	MemcachedBytes              MetricIntf
	MemcachedCommandCount       MetricIntf
	MemcachedCurrentConnections MetricIntf
	MemcachedCurrentItems       MetricIntf
	MemcachedEvictionCount      MetricIntf
	MemcachedNetwork            MetricIntf
	MemcachedOperationCount     MetricIntf
	MemcachedOperationHitRatio  MetricIntf
	MemcachedRusage             MetricIntf
	MemcachedThreads            MetricIntf
	MemcachedTotalConnections   MetricIntf
}

// Names returns a list of all the metric name strings.
func (m *metricStruct) Names() []string {
	return []string{
		"memcached.bytes",
		"memcached.command_count",
		"memcached.current_connections",
		"memcached.current_items",
		"memcached.eviction_count",
		"memcached.network",
		"memcached.operation_count",
		"memcached.operation_hit_ratio",
		"memcached.rusage",
		"memcached.threads",
		"memcached.total_connections",
	}
}

var metricsByName = map[string]MetricIntf{
	"memcached.bytes":               Metrics.MemcachedBytes,
	"memcached.command_count":       Metrics.MemcachedCommandCount,
	"memcached.current_connections": Metrics.MemcachedCurrentConnections,
	"memcached.current_items":       Metrics.MemcachedCurrentItems,
	"memcached.eviction_count":      Metrics.MemcachedEvictionCount,
	"memcached.network":             Metrics.MemcachedNetwork,
	"memcached.operation_count":     Metrics.MemcachedOperationCount,
	"memcached.operation_hit_ratio": Metrics.MemcachedOperationHitRatio,
	"memcached.rusage":              Metrics.MemcachedRusage,
	"memcached.threads":             Metrics.MemcachedThreads,
	"memcached.total_connections":   Metrics.MemcachedTotalConnections,
}

func (m *metricStruct) ByName(n string) MetricIntf {
	return metricsByName[n]
}

func (m *metricStruct) FactoriesByName() map[string]func(pdata.Metric) {
	return map[string]func(pdata.Metric){
		Metrics.MemcachedBytes.Name():              Metrics.MemcachedBytes.Init,
		Metrics.MemcachedCommandCount.Name():       Metrics.MemcachedCommandCount.Init,
		Metrics.MemcachedCurrentConnections.Name(): Metrics.MemcachedCurrentConnections.Init,
		Metrics.MemcachedCurrentItems.Name():       Metrics.MemcachedCurrentItems.Init,
		Metrics.MemcachedEvictionCount.Name():      Metrics.MemcachedEvictionCount.Init,
		Metrics.MemcachedNetwork.Name():            Metrics.MemcachedNetwork.Init,
		Metrics.MemcachedOperationCount.Name():     Metrics.MemcachedOperationCount.Init,
		Metrics.MemcachedOperationHitRatio.Name():  Metrics.MemcachedOperationHitRatio.Init,
		Metrics.MemcachedRusage.Name():             Metrics.MemcachedRusage.Init,
		Metrics.MemcachedThreads.Name():            Metrics.MemcachedThreads.Init,
		Metrics.MemcachedTotalConnections.Name():   Metrics.MemcachedTotalConnections.Init,
	}
}

// Metrics contains a set of methods for each metric that help with
// manipulating those metrics.
var Metrics = &metricStruct{
	&metricImpl{
		"memcached.bytes",
		func(metric pdata.Metric) {
			metric.SetName("memcached.bytes")
			metric.SetDescription("Current number of bytes used by this server to store items")
			metric.SetUnit("By")
			metric.SetDataType(pdata.MetricDataTypeIntGauge)
		},
	},
	&metricImpl{
		"memcached.command_count",
		func(metric pdata.Metric) {
			metric.SetName("memcached.command_count")
			metric.SetDescription("Commands executed")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"memcached.current_connections",
		func(metric pdata.Metric) {
			metric.SetName("memcached.current_connections")
			metric.SetDescription("The current number of open connections")
			metric.SetUnit("connections")
			metric.SetDataType(pdata.MetricDataTypeIntGauge)
		},
	},
	&metricImpl{
		"memcached.current_items",
		func(metric pdata.Metric) {
			metric.SetName("memcached.current_items")
			metric.SetDescription("Number of items currently stored in the cache")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeDoubleGauge)
		},
	},
	&metricImpl{
		"memcached.eviction_count",
		func(metric pdata.Metric) {
			metric.SetName("memcached.eviction_count")
			metric.SetDescription("Cache item evictions")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"memcached.network",
		func(metric pdata.Metric) {
			metric.SetName("memcached.network")
			metric.SetDescription("Bytes transferred over the network.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"memcached.operation_count",
		func(metric pdata.Metric) {
			metric.SetName("memcached.operation_count")
			metric.SetDescription("Memcached operation hit/miss counts.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"memcached.operation_hit_ratio",
		func(metric pdata.Metric) {
			metric.SetName("memcached.operation_hit_ratio")
			metric.SetDescription("Hit ratio for memcached operations, expressed as a percentage value between 0.0 and 100.0.")
			metric.SetUnit("%")
			metric.SetDataType(pdata.MetricDataTypeDoubleGauge)
		},
	},
	&metricImpl{
		"memcached.rusage",
		func(metric pdata.Metric) {
			metric.SetName("memcached.rusage")
			metric.SetDescription("Accumulated user and system time.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeDoubleGauge)
		},
	},
	&metricImpl{
		"memcached.threads",
		func(metric pdata.Metric) {
			metric.SetName("memcached.threads")
			metric.SetDescription("Number of threads used by the memcached instance")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeDoubleGauge)
		},
	},
	&metricImpl{
		"memcached.total_connections",
		func(metric pdata.Metric) {
			metric.SetName("memcached.total_connections")
			metric.SetDescription("Total number of connections opened since the server started running")
			metric.SetUnit("connections")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
}

// M contains a set of methods for each metric that help with
// manipulating those metrics. M is an alias for Metrics
var M = Metrics

// Labels contains the possible metric labels that can be used.
var Labels = struct {
	// Command (The type of command)
	Command string
	// Direction (direction of data flow)
	Direction string
	// Operation (the type of operation)
	Operation string
	// Type (hit/miss)
	Type string
	// UsageType (type of CPU usage)
	UsageType string
}{
	"command",
	"direction",
	"operation",
	"type",
	"usage_type",
}

// L contains the possible metric labels that can be used. L is an alias for
// Labels.
var L = Labels

// LabelCommand are the possible values that the label "command" can have.
var LabelCommand = struct {
	Get   string
	Set   string
	Flush string
	Touch string
	Meta  string
}{
	"get",
	"set",
	"flush",
	"touch",
	"meta",
}

// LabelDirection are the possible values that the label "direction" can have.
var LabelDirection = struct {
	Sent     string
	Received string
}{
	"sent",
	"received",
}

// LabelOperation are the possible values that the label "operation" can have.
var LabelOperation = struct {
	Increment string
	Decrement string
	Get       string
	Delete    string
	Cas       string
	Touch     string
}{
	"increment",
	"decrement",
	"get",
	"delete",
	"cas",
	"touch",
}

// LabelType are the possible values that the label "type" can have.
var LabelType = struct {
	Hit  string
	Miss string
}{
	"hit",
	"miss",
}

// LabelUsageType are the possible values that the label "usage_type" can have.
var LabelUsageType = struct {
	System string
	User   string
}{
	"system",
	"user",
}
