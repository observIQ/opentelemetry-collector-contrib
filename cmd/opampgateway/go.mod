module github.com/open-telemetry/opentelemetry-collector-contrib/cmd/opampgateway

go 1.23.6

require (
	github.com/open-telemetry/opamp-go v0.19.0
	go.uber.org/zap v1.27.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	google.golang.org/protobuf v1.36.2 // indirect
)

replace github.com/open-telemetry/opamp-go => github.com/observIQ/opamp-go v0.7.1-0.20250221022807-597ba35629e9
