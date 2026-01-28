package member

import "go.opentelemetry.io/otel"

var (
	serviceTracer = otel.Tracer("auth/service")
	repoTracer    = otel.Tracer("auth/repository")
)
