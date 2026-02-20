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
		metric.WithExplicitBucketBoundaries(
			0.001, // 1ms
			0.002, // 2ms
			0.005, // 5ms
			0.01,  // 10ms
			0.015, // 15ms
			0.02,  // 20ms
			0.025, // 25ms
			0.03,  // 30ms
			0.033, // 33ms  NOTE: budget limit when runnign at 30 ticks a second
			0.05,  // 50ms  WARN: over budget
			0.1,   // 100ms WARN: over budget
		),
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
