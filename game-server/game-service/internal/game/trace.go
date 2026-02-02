package game

import "go.opentelemetry.io/otel"

var gameItemPoolTracer = otel.Tracer("Game.InitItemPool")
