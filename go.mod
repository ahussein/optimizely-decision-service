module github.com/ahussein/optimizely-decision-service

go 1.14

require (
	contrib.go.opencensus.io/exporter/jaeger v0.2.1
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/optimizely/go-sdk v1.6.1
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.6.1
	go.opencensus.io v0.22.4
	go.opentelemetry.io/otel v0.9.0
	go.uber.org/zap v1.15.0
)
