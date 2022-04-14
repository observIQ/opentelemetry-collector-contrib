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

package golden

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestWriteMetrics(t *testing.T) {
	metricslice := testMetrics()
	metrics := pmetric.NewMetrics()
	metricslice.CopyTo(metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics())

	tempDir := filepath.Join(t.TempDir(), "metrics.json")
	WriteMetrics(tempDir, metrics)

	actualBytes, err := ioutil.ReadFile(tempDir)
	require.NoError(t, err)

	expectedFile := filepath.Join("testdata", "roundtrip", "expected.json")
	expectedBytes, err := ioutil.ReadFile(expectedFile)
	require.NoError(t, err)

	require.Equal(t, expectedBytes, actualBytes)
}

func TestReadMetrics(t *testing.T) {
	metricslice := testMetrics()
	expectedMetrics := pmetric.NewMetrics()
	metricslice.CopyTo(expectedMetrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics())

	expectedFile := filepath.Join("testdata", "roundtrip", "expected.json")
	actualMetrics, err := ReadMetrics(expectedFile)
	require.NoError(t, err)
	require.Equal(t, expectedMetrics, actualMetrics)
}

func TestRoundTrip(t *testing.T) {
	metricslice := testMetrics()
	expectedMetrics := pmetric.NewMetrics()
	metricslice.CopyTo(expectedMetrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics())

	tempDir := filepath.Join(t.TempDir(), "metrics.json")
	err := WriteMetrics(tempDir, expectedMetrics)
	require.NoError(t, err)

	actualMetrics, err := ReadMetrics(tempDir)
	require.NoError(t, err)
	require.Equal(t, expectedMetrics, actualMetrics)
}

func testMetrics() pmetric.MetricSlice {
	slice := pmetric.NewMetricSlice()

	// Gauge with two double dps
	metric := slice.AppendEmpty()
	initGauge(metric, "test gauge multi", "multi gauge", "1")
	dps := metric.Gauge().DataPoints()

	dp := dps.AppendEmpty()
	attributes := pcommon.NewMap()
	attributes.Insert("testKey1", pcommon.NewValueString("teststringvalue1"))
	attributes.Insert("testKey2", pcommon.NewValueString("testvalue1"))
	setDPDoubleVal(dp, 2, attributes, time.Time{})

	dp = dps.AppendEmpty()
	attributes = pcommon.NewMap()
	attributes.Insert("testKey1", pcommon.NewValueString("teststringvalue2"))
	attributes.Insert("testKey2", pcommon.NewValueString("testvalue2"))
	setDPDoubleVal(dp, 2, attributes, time.Time{})

	// Gauge with one int dp
	metric = slice.AppendEmpty()
	initGauge(metric, "test gauge single", "single gauge", "By")
	dps = metric.Gauge().DataPoints()

	dp = dps.AppendEmpty()
	attributes = pcommon.NewMap()
	attributes.Insert("testKey2", pcommon.NewValueString("teststringvalue2"))
	setDPIntVal(dp, 2, attributes, time.Time{})

	// Delta Sum with two int dps
	metric = slice.AppendEmpty()
	initSum(metric, "test delta sum multi", "multi sum", "s", pmetric.MetricAggregationTemporalityDelta, false)
	dps = metric.Sum().DataPoints()

	dp = dps.AppendEmpty()
	attributes = pcommon.NewMap()
	attributes.Insert("testKey2", pcommon.NewValueString("teststringvalue2"))
	setDPIntVal(dp, 2, attributes, time.Time{})

	dp = dps.AppendEmpty()
	attributes = pcommon.NewMap()
	attributes.Insert("testKey2", pcommon.NewValueString("teststringvalue2"))
	setDPIntVal(dp, 2, attributes, time.Time{})

	// Cumulative Sum with one double dp
	metric = slice.AppendEmpty()
	initSum(metric, "test cumulative sum single", "single sum", "1/s", pmetric.MetricAggregationTemporalityCumulative, true)
	dps = metric.Sum().DataPoints()

	dp = dps.AppendEmpty()
	attributes = pcommon.NewMap()
	setDPDoubleVal(dp, 2, attributes, time.Date(1997, 07, 27, 1, 1, 1, 1, &time.Location{}))
	return slice
}

func setDPDoubleVal(dp pmetric.NumberDataPoint, value float64, attributes pcommon.Map, timeStamp time.Time) {
	dp.SetDoubleVal(value)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(timeStamp))
	attributes.CopyTo(dp.Attributes())
}

func setDPIntVal(dp pmetric.NumberDataPoint, value int64, attributes pcommon.Map, timeStamp time.Time) {
	dp.SetIntVal(value)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(timeStamp))
	attributes.CopyTo(dp.Attributes())
}

func initGauge(metric pmetric.Metric, name, desc, unit string) {
	metric.SetDataType(pmetric.MetricDataTypeGauge)
	metric.SetDescription(desc)
	metric.SetName(name)
	metric.SetUnit(unit)
}

func initSum(metric pmetric.Metric, name, desc, unit string, aggr pmetric.MetricAggregationTemporality, isMonotonic bool) {
	metric.SetDataType(pmetric.MetricDataTypeSum)
	metric.Sum().SetIsMonotonic(isMonotonic)
	metric.Sum().SetAggregationTemporality(aggr)
	metric.SetDescription(desc)
	metric.SetName(name)
	metric.SetUnit(unit)
}
