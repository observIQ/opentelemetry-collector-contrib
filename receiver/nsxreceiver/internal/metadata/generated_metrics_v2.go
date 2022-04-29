// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// MetricSettings provides common settings for a particular metric.
type MetricSettings struct {
	Enabled bool `mapstructure:"enabled"`
}

// MetricsSettings provides settings for nsxreceiver metrics.
type MetricsSettings struct {
	NodeCacheMemoryUsage        MetricSettings `mapstructure:"node.cache.memory.usage"`
	NodeCacheNetworkThroughput  MetricSettings `mapstructure:"node.cache.network.throughput"`
	NodeLoadBalancerUtilization MetricSettings `mapstructure:"node.load_balancer.utilization"`
	NodeMemoryUsage             MetricSettings `mapstructure:"node.memory.usage"`
	NsxInterfacePacketCount     MetricSettings `mapstructure:"nsx.interface.packet.count"`
	NsxInterfaceThroughput      MetricSettings `mapstructure:"nsx.interface.throughput"`
	NsxNodeCPUUtilization       MetricSettings `mapstructure:"nsx.node.cpu.utilization"`
	NsxNodeDiskUsage            MetricSettings `mapstructure:"nsx.node.disk.usage"`
	NsxNodeDiskUtilization      MetricSettings `mapstructure:"nsx.node.disk.utilization"`
}

func DefaultMetricsSettings() MetricsSettings {
	return MetricsSettings{
		NodeCacheMemoryUsage: MetricSettings{
			Enabled: true,
		},
		NodeCacheNetworkThroughput: MetricSettings{
			Enabled: true,
		},
		NodeLoadBalancerUtilization: MetricSettings{
			Enabled: true,
		},
		NodeMemoryUsage: MetricSettings{
			Enabled: true,
		},
		NsxInterfacePacketCount: MetricSettings{
			Enabled: true,
		},
		NsxInterfaceThroughput: MetricSettings{
			Enabled: true,
		},
		NsxNodeCPUUtilization: MetricSettings{
			Enabled: true,
		},
		NsxNodeDiskUsage: MetricSettings{
			Enabled: true,
		},
		NsxNodeDiskUtilization: MetricSettings{
			Enabled: true,
		},
	}
}

type metricNodeCacheMemoryUsage struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills node.cache.memory.usage metric with initial data.
func (m *metricNodeCacheMemoryUsage) init() {
	m.data.SetName("node.cache.memory.usage")
	m.data.SetDescription("The memory usage of the node’s cache")
	m.data.SetUnit("By")
	m.data.SetDataType(pmetric.MetricDataTypeSum)
	m.data.Sum().SetIsMonotonic(false)
	m.data.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityCumulative)
}

func (m *metricNodeCacheMemoryUsage) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Sum().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntVal(val)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNodeCacheMemoryUsage) updateCapacity() {
	if m.data.Sum().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Sum().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNodeCacheMemoryUsage) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Sum().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNodeCacheMemoryUsage(settings MetricSettings) metricNodeCacheMemoryUsage {
	m := metricNodeCacheMemoryUsage{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNodeCacheNetworkThroughput struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills node.cache.network.throughput metric with initial data.
func (m *metricNodeCacheNetworkThroughput) init() {
	m.data.SetName("node.cache.network.throughput")
	m.data.SetDescription("The memory usage of the node’s cache")
	m.data.SetUnit("By")
	m.data.SetDataType(pmetric.MetricDataTypeSum)
	m.data.Sum().SetIsMonotonic(false)
	m.data.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityCumulative)
}

func (m *metricNodeCacheNetworkThroughput) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Sum().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntVal(val)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNodeCacheNetworkThroughput) updateCapacity() {
	if m.data.Sum().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Sum().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNodeCacheNetworkThroughput) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Sum().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNodeCacheNetworkThroughput(settings MetricSettings) metricNodeCacheNetworkThroughput {
	m := metricNodeCacheNetworkThroughput{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNodeLoadBalancerUtilization struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills node.load_balancer.utilization metric with initial data.
func (m *metricNodeLoadBalancerUtilization) init() {
	m.data.SetName("node.load_balancer.utilization")
	m.data.SetDescription("The utilization of load balancers by the node")
	m.data.SetUnit("%")
	m.data.SetDataType(pmetric.MetricDataTypeGauge)
	m.data.Gauge().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricNodeLoadBalancerUtilization) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val float64, loadBalancerAttributeValue string) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetDoubleVal(val)
	dp.Attributes().Insert(A.LoadBalancer, pcommon.NewValueString(loadBalancerAttributeValue))
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNodeLoadBalancerUtilization) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNodeLoadBalancerUtilization) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNodeLoadBalancerUtilization(settings MetricSettings) metricNodeLoadBalancerUtilization {
	m := metricNodeLoadBalancerUtilization{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNodeMemoryUsage struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills node.memory.usage metric with initial data.
func (m *metricNodeMemoryUsage) init() {
	m.data.SetName("node.memory.usage")
	m.data.SetDescription("The memory usage of the node")
	m.data.SetUnit("By")
	m.data.SetDataType(pmetric.MetricDataTypeSum)
	m.data.Sum().SetIsMonotonic(false)
	m.data.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityCumulative)
}

func (m *metricNodeMemoryUsage) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Sum().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntVal(val)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNodeMemoryUsage) updateCapacity() {
	if m.data.Sum().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Sum().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNodeMemoryUsage) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Sum().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNodeMemoryUsage(settings MetricSettings) metricNodeMemoryUsage {
	m := metricNodeMemoryUsage{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNsxInterfacePacketCount struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills nsx.interface.packet.count metric with initial data.
func (m *metricNsxInterfacePacketCount) init() {
	m.data.SetName("nsx.interface.packet.count")
	m.data.SetDescription("The number of packets flowing through the network interface.")
	m.data.SetUnit("{messages}")
	m.data.SetDataType(pmetric.MetricDataTypeSum)
	m.data.Sum().SetIsMonotonic(true)
	m.data.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityCumulative)
	m.data.Sum().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricNsxInterfacePacketCount) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64, directionAttributeValue string) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Sum().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntVal(val)
	dp.Attributes().Insert(A.Direction, pcommon.NewValueString(directionAttributeValue))
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNsxInterfacePacketCount) updateCapacity() {
	if m.data.Sum().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Sum().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNsxInterfacePacketCount) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Sum().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNsxInterfacePacketCount(settings MetricSettings) metricNsxInterfacePacketCount {
	m := metricNsxInterfacePacketCount{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNsxInterfaceThroughput struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills nsx.interface.throughput metric with initial data.
func (m *metricNsxInterfaceThroughput) init() {
	m.data.SetName("nsx.interface.throughput")
	m.data.SetDescription("The number of messages published to a queue.")
	m.data.SetUnit("{messages}")
	m.data.SetDataType(pmetric.MetricDataTypeSum)
	m.data.Sum().SetIsMonotonic(true)
	m.data.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityCumulative)
	m.data.Sum().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricNsxInterfaceThroughput) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64, directionAttributeValue string) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Sum().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntVal(val)
	dp.Attributes().Insert(A.Direction, pcommon.NewValueString(directionAttributeValue))
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNsxInterfaceThroughput) updateCapacity() {
	if m.data.Sum().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Sum().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNsxInterfaceThroughput) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Sum().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNsxInterfaceThroughput(settings MetricSettings) metricNsxInterfaceThroughput {
	m := metricNsxInterfaceThroughput{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNsxNodeCPUUtilization struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills nsx.node.cpu.utilization metric with initial data.
func (m *metricNsxNodeCPUUtilization) init() {
	m.data.SetName("nsx.node.cpu.utilization")
	m.data.SetDescription("The average amount of CPU being used by the node.")
	m.data.SetUnit("%")
	m.data.SetDataType(pmetric.MetricDataTypeGauge)
	m.data.Gauge().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricNsxNodeCPUUtilization) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val float64, cpuProcessClassAttributeValue string) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetDoubleVal(val)
	dp.Attributes().Insert(A.CPUProcessClass, pcommon.NewValueString(cpuProcessClassAttributeValue))
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNsxNodeCPUUtilization) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNsxNodeCPUUtilization) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNsxNodeCPUUtilization(settings MetricSettings) metricNsxNodeCPUUtilization {
	m := metricNsxNodeCPUUtilization{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNsxNodeDiskUsage struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills nsx.node.disk.usage metric with initial data.
func (m *metricNsxNodeDiskUsage) init() {
	m.data.SetName("nsx.node.disk.usage")
	m.data.SetDescription("The amount of storage space used by the node.")
	m.data.SetUnit("By")
	m.data.SetDataType(pmetric.MetricDataTypeSum)
	m.data.Sum().SetIsMonotonic(false)
	m.data.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityCumulative)
	m.data.Sum().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricNsxNodeDiskUsage) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64, diskAttributeValue string) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Sum().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntVal(val)
	dp.Attributes().Insert(A.Disk, pcommon.NewValueString(diskAttributeValue))
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNsxNodeDiskUsage) updateCapacity() {
	if m.data.Sum().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Sum().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNsxNodeDiskUsage) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Sum().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNsxNodeDiskUsage(settings MetricSettings) metricNsxNodeDiskUsage {
	m := metricNsxNodeDiskUsage{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNsxNodeDiskUtilization struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills nsx.node.disk.utilization metric with initial data.
func (m *metricNsxNodeDiskUtilization) init() {
	m.data.SetName("nsx.node.disk.utilization")
	m.data.SetDescription("The percentage of storage space utilized.")
	m.data.SetUnit("%")
	m.data.SetDataType(pmetric.MetricDataTypeGauge)
	m.data.Gauge().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricNsxNodeDiskUtilization) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val float64, diskAttributeValue string) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetDoubleVal(val)
	dp.Attributes().Insert(A.Disk, pcommon.NewValueString(diskAttributeValue))
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNsxNodeDiskUtilization) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNsxNodeDiskUtilization) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNsxNodeDiskUtilization(settings MetricSettings) metricNsxNodeDiskUtilization {
	m := metricNsxNodeDiskUtilization{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

// MetricsBuilder provides an interface for scrapers to report metrics while taking care of all the transformations
// required to produce metric representation defined in metadata and user settings.
type MetricsBuilder struct {
	startTime                         pcommon.Timestamp // start time that will be applied to all recorded data points.
	metricsCapacity                   int               // maximum observed number of metrics per resource.
	resourceCapacity                  int               // maximum observed number of resource attributes.
	metricsBuffer                     pmetric.Metrics   // accumulates metrics data before emitting.
	metricNodeCacheMemoryUsage        metricNodeCacheMemoryUsage
	metricNodeCacheNetworkThroughput  metricNodeCacheNetworkThroughput
	metricNodeLoadBalancerUtilization metricNodeLoadBalancerUtilization
	metricNodeMemoryUsage             metricNodeMemoryUsage
	metricNsxInterfacePacketCount     metricNsxInterfacePacketCount
	metricNsxInterfaceThroughput      metricNsxInterfaceThroughput
	metricNsxNodeCPUUtilization       metricNsxNodeCPUUtilization
	metricNsxNodeDiskUsage            metricNsxNodeDiskUsage
	metricNsxNodeDiskUtilization      metricNsxNodeDiskUtilization
}

// metricBuilderOption applies changes to default metrics builder.
type metricBuilderOption func(*MetricsBuilder)

// WithStartTime sets startTime on the metrics builder.
func WithStartTime(startTime pcommon.Timestamp) metricBuilderOption {
	return func(mb *MetricsBuilder) {
		mb.startTime = startTime
	}
}

func NewMetricsBuilder(settings MetricsSettings, options ...metricBuilderOption) *MetricsBuilder {
	mb := &MetricsBuilder{
		startTime:                         pcommon.NewTimestampFromTime(time.Now()),
		metricsBuffer:                     pmetric.NewMetrics(),
		metricNodeCacheMemoryUsage:        newMetricNodeCacheMemoryUsage(settings.NodeCacheMemoryUsage),
		metricNodeCacheNetworkThroughput:  newMetricNodeCacheNetworkThroughput(settings.NodeCacheNetworkThroughput),
		metricNodeLoadBalancerUtilization: newMetricNodeLoadBalancerUtilization(settings.NodeLoadBalancerUtilization),
		metricNodeMemoryUsage:             newMetricNodeMemoryUsage(settings.NodeMemoryUsage),
		metricNsxInterfacePacketCount:     newMetricNsxInterfacePacketCount(settings.NsxInterfacePacketCount),
		metricNsxInterfaceThroughput:      newMetricNsxInterfaceThroughput(settings.NsxInterfaceThroughput),
		metricNsxNodeCPUUtilization:       newMetricNsxNodeCPUUtilization(settings.NsxNodeCPUUtilization),
		metricNsxNodeDiskUsage:            newMetricNsxNodeDiskUsage(settings.NsxNodeDiskUsage),
		metricNsxNodeDiskUtilization:      newMetricNsxNodeDiskUtilization(settings.NsxNodeDiskUtilization),
	}
	for _, op := range options {
		op(mb)
	}
	return mb
}

// updateCapacity updates max length of metrics and resource attributes that will be used for the slice capacity.
func (mb *MetricsBuilder) updateCapacity(rm pmetric.ResourceMetrics) {
	if mb.metricsCapacity < rm.ScopeMetrics().At(0).Metrics().Len() {
		mb.metricsCapacity = rm.ScopeMetrics().At(0).Metrics().Len()
	}
	if mb.resourceCapacity < rm.Resource().Attributes().Len() {
		mb.resourceCapacity = rm.Resource().Attributes().Len()
	}
}

// ResourceOption applies changes to provided resource.
type ResourceOption func(pcommon.Resource)

// WithNsxInterfaceID sets provided value as "nsx.interface.id" attribute for current resource.
func WithNsxInterfaceID(val string) ResourceOption {
	return func(r pcommon.Resource) {
		r.Attributes().UpsertString("nsx.interface.id", val)
	}
}

// WithNsxNodeName sets provided value as "nsx.node.name" attribute for current resource.
func WithNsxNodeName(val string) ResourceOption {
	return func(r pcommon.Resource) {
		r.Attributes().UpsertString("nsx.node.name", val)
	}
}

// WithNsxNodeType sets provided value as "nsx.node.type" attribute for current resource.
func WithNsxNodeType(val string) ResourceOption {
	return func(r pcommon.Resource) {
		r.Attributes().UpsertString("nsx.node.type", val)
	}
}

// EmitForResource saves all the generated metrics under a new resource and updates the internal state to be ready for
// recording another set of data points as part of another resource. This function can be helpful when one scraper
// needs to emit metrics from several resources. Otherwise calling this function is not required,
// just `Emit` function can be called instead. Resource attributes should be provided as ResourceOption arguments.
func (mb *MetricsBuilder) EmitForResource(ro ...ResourceOption) {
	rm := pmetric.NewResourceMetrics()
	rm.Resource().Attributes().EnsureCapacity(mb.resourceCapacity)
	for _, op := range ro {
		op(rm.Resource())
	}
	ils := rm.ScopeMetrics().AppendEmpty()
	ils.Scope().SetName("otelcol/nsxreceiver")
	ils.Metrics().EnsureCapacity(mb.metricsCapacity)
	mb.metricNodeCacheMemoryUsage.emit(ils.Metrics())
	mb.metricNodeCacheNetworkThroughput.emit(ils.Metrics())
	mb.metricNodeLoadBalancerUtilization.emit(ils.Metrics())
	mb.metricNodeMemoryUsage.emit(ils.Metrics())
	mb.metricNsxInterfacePacketCount.emit(ils.Metrics())
	mb.metricNsxInterfaceThroughput.emit(ils.Metrics())
	mb.metricNsxNodeCPUUtilization.emit(ils.Metrics())
	mb.metricNsxNodeDiskUsage.emit(ils.Metrics())
	mb.metricNsxNodeDiskUtilization.emit(ils.Metrics())
	if ils.Metrics().Len() > 0 {
		mb.updateCapacity(rm)
		rm.MoveTo(mb.metricsBuffer.ResourceMetrics().AppendEmpty())
	}
}

// Emit returns all the metrics accumulated by the metrics builder and updates the internal state to be ready for
// recording another set of metrics. This function will be responsible for applying all the transformations required to
// produce metric representation defined in metadata and user settings, e.g. delta or cumulative.
func (mb *MetricsBuilder) Emit(ro ...ResourceOption) pmetric.Metrics {
	mb.EmitForResource(ro...)
	metrics := pmetric.NewMetrics()
	mb.metricsBuffer.MoveTo(metrics)
	return metrics
}

// RecordNodeCacheMemoryUsageDataPoint adds a data point to node.cache.memory.usage metric.
func (mb *MetricsBuilder) RecordNodeCacheMemoryUsageDataPoint(ts pcommon.Timestamp, val int64) {
	mb.metricNodeCacheMemoryUsage.recordDataPoint(mb.startTime, ts, val)
}

// RecordNodeCacheNetworkThroughputDataPoint adds a data point to node.cache.network.throughput metric.
func (mb *MetricsBuilder) RecordNodeCacheNetworkThroughputDataPoint(ts pcommon.Timestamp, val int64) {
	mb.metricNodeCacheNetworkThroughput.recordDataPoint(mb.startTime, ts, val)
}

// RecordNodeLoadBalancerUtilizationDataPoint adds a data point to node.load_balancer.utilization metric.
func (mb *MetricsBuilder) RecordNodeLoadBalancerUtilizationDataPoint(ts pcommon.Timestamp, val float64, loadBalancerAttributeValue string) {
	mb.metricNodeLoadBalancerUtilization.recordDataPoint(mb.startTime, ts, val, loadBalancerAttributeValue)
}

// RecordNodeMemoryUsageDataPoint adds a data point to node.memory.usage metric.
func (mb *MetricsBuilder) RecordNodeMemoryUsageDataPoint(ts pcommon.Timestamp, val int64) {
	mb.metricNodeMemoryUsage.recordDataPoint(mb.startTime, ts, val)
}

// RecordNsxInterfacePacketCountDataPoint adds a data point to nsx.interface.packet.count metric.
func (mb *MetricsBuilder) RecordNsxInterfacePacketCountDataPoint(ts pcommon.Timestamp, val int64, directionAttributeValue string) {
	mb.metricNsxInterfacePacketCount.recordDataPoint(mb.startTime, ts, val, directionAttributeValue)
}

// RecordNsxInterfaceThroughputDataPoint adds a data point to nsx.interface.throughput metric.
func (mb *MetricsBuilder) RecordNsxInterfaceThroughputDataPoint(ts pcommon.Timestamp, val int64, directionAttributeValue string) {
	mb.metricNsxInterfaceThroughput.recordDataPoint(mb.startTime, ts, val, directionAttributeValue)
}

// RecordNsxNodeCPUUtilizationDataPoint adds a data point to nsx.node.cpu.utilization metric.
func (mb *MetricsBuilder) RecordNsxNodeCPUUtilizationDataPoint(ts pcommon.Timestamp, val float64, cpuProcessClassAttributeValue string) {
	mb.metricNsxNodeCPUUtilization.recordDataPoint(mb.startTime, ts, val, cpuProcessClassAttributeValue)
}

// RecordNsxNodeDiskUsageDataPoint adds a data point to nsx.node.disk.usage metric.
func (mb *MetricsBuilder) RecordNsxNodeDiskUsageDataPoint(ts pcommon.Timestamp, val int64, diskAttributeValue string) {
	mb.metricNsxNodeDiskUsage.recordDataPoint(mb.startTime, ts, val, diskAttributeValue)
}

// RecordNsxNodeDiskUtilizationDataPoint adds a data point to nsx.node.disk.utilization metric.
func (mb *MetricsBuilder) RecordNsxNodeDiskUtilizationDataPoint(ts pcommon.Timestamp, val float64, diskAttributeValue string) {
	mb.metricNsxNodeDiskUtilization.recordDataPoint(mb.startTime, ts, val, diskAttributeValue)
}

// Reset resets metrics builder to its initial state. It should be used when external metrics source is restarted,
// and metrics builder should update its startTime and reset it's internal state accordingly.
func (mb *MetricsBuilder) Reset(options ...metricBuilderOption) {
	mb.startTime = pcommon.NewTimestampFromTime(time.Now())
	for _, op := range options {
		op(mb)
	}
}

// Attributes contains the possible metric attributes that can be used.
var Attributes = struct {
	// CPUProcessClass (The CPU usage of the architecture allocated for either DPDK (datapath) or non-DPDK (services) processes)
	CPUProcessClass string
	// Direction (The direction of network flow)
	Direction string
	// Disk (The name of the mounted storage)
	Disk string
	// LoadBalancer (The name of the load balancer being utilized)
	LoadBalancer string
}{
	"cpu.process.class",
	"direction",
	"disk",
	"load_balancer",
}

// A is an alias for Attributes.
var A = Attributes

// AttributeCPUProcessClass are the possible values that the attribute "cpu.process.class" can have.
var AttributeCPUProcessClass = struct {
	Datapath string
	Services string
}{
	"datapath",
	"services",
}

// AttributeDirection are the possible values that the attribute "direction" can have.
var AttributeDirection = struct {
	Received    string
	Transmitted string
}{
	"received",
	"transmitted",
}
