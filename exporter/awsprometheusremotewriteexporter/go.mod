module github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsprometheusremotewriteexporter

go 1.16

require (
	github.com/aws/aws-sdk-go v1.40.2
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.29.1-0.20210708235311-fb95c88e72fa
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
)

replace go.opentelemetry.io/collector/model => go.opentelemetry.io/collector/model v0.0.0-20210708235311-fb95c88e72fa
