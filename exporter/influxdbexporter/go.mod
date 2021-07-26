module github.com/open-telemetry/opentelemetry-collector-contrib/exporter/influxdbexporter

go 1.16

require (
	github.com/influxdata/influxdb-observability/common v0.2.4
	github.com/influxdata/influxdb-observability/otel2influx v0.2.4
	github.com/influxdata/line-protocol/v2 v2.0.0-20210520103755-6551a972d603
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.30.2-0.20210719230137-809cae954ed3
	go.opentelemetry.io/collector/model v0.30.2-0.20210719230137-809cae954ed3
	go.uber.org/zap v1.18.1
)
