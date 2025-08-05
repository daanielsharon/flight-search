module provider-service

go 1.24.1

require github.com/redis/go-redis/v9 v9.11.0

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
)

require (
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0
	go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	shared v0.0.0
)

replace shared => ../shared
