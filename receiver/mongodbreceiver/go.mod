module github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver

go 1.16

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.mongodb.org/mongo-driver v1.6.0
	go.opentelemetry.io/collector v0.29.1-0.20210708235311-fb95c88e72fa
	go.opentelemetry.io/collector/model v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.18.1
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/common => ../../internal/common

replace go.opentelemetry.io/collector/model => go.opentelemetry.io/collector/model v0.0.0-20210708235311-fb95c88e72fa
