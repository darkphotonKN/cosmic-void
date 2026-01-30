package items

import "go.opentelemetry.io/otel"

var itemRepositoryTracer = otel.Tracer("Items.Repo")
