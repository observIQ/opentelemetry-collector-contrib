module github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver

go 1.16

require (
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/prometheus/common v0.29.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/testcontainers/testcontainers-go v0.11.1 // indirect
	go.opentelemetry.io/collector v0.29.1-0.20210702192737-aaa6d7d6b859
	go.opentelemetry.io/collector/model v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.18.1
	k8s.io/client-go v0.21.1 // indirect
)

replace go.opentelemetry.io/collector/model => go.opentelemetry.io/collector/model v0.0.0-20210702192737-aaa6d7d6b859
