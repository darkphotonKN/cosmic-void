package metrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

/**
* Custom metrics for tracking game loop time.
**/

var (
	meter = otel.Meter("game-service")

	TickDuration metric.Float64Histogram
	EntityCount  metric.Int64Gauge
)

func Init() error {
	var err error

	TickDuration, err = meter.Float64Histogram(
		"game.tick.duration_seconds",
		metric.WithDescription("Time to process one game tick"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	EntityCount, err = meter.Int64Gauge(
		"game.tick.entity_count",
		metric.WithDescription("Entities processed per tick"),
	)
	if err != nil {
		return err
	}

	return nil
}
