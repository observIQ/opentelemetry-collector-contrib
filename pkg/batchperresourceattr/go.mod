module github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr

go 1.16

require (
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.29.1-0.20210708235311-fb95c88e72fa
	go.opentelemetry.io/collector/model v0.0.0-00010101000000-000000000000
)

replace go.opentelemetry.io/collector/model => go.opentelemetry.io/collector/model v0.0.0-20210708235311-fb95c88e72fa
