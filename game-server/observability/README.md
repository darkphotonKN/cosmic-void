# Observability Stack

Complete local setup for traces, metrics, and logs.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            MICROSERVICES                                    │
│                                                                             │
│   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                 │
│   │ game-service │    │stats-service │    │ auth-service │                 │
│   │              │    │              │    │              │                 │
│   │ telemetry.   │    │ telemetry.   │    │ telemetry.   │                 │
│   │   Init()     │    │   Init()     │    │   Init()     │                 │
│   └──────┬───────┘    └──────┬───────┘    └──────┬───────┘                 │
│          │                   │                   │                          │
│          └───────────────────┼───────────────────┘                          │
│                              │                                              │
│                              ▼                                              │
│                    OTLP gRPC (port 4317)                                   │
└──────────────────────────────┼──────────────────────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────────────────────┐
│                              ▼                                              │
│                  ┌───────────────────────┐                                  │
│                  │    OTEL COLLECTOR     │                                  │
│                  │                       │                                  │
│                  │  Receives all data    │                                  │
│                  │  Routes to backends   │                                  │
│                  └───────────┬───────────┘                                  │
│                              │                                              │
│            ┌─────────────────┼─────────────────┐                           │
│            │                 │                 │                            │
│            ▼                 ▼                 ▼                            │
│     ┌──────────┐      ┌──────────┐      ┌──────────┐                       │
│     │  TEMPO   │      │PROMETHEUS│      │   LOKI   │                       │
│     │ (traces) │      │ (metrics)│      │  (logs)  │                       │
│     │  :3200   │      │  :9090   │      │  :3100   │                       │
│     └────┬─────┘      └────┬─────┘      └────┬─────┘                       │
│          │                 │                 │                              │
│          └─────────────────┼─────────────────┘                              │
│                            │                                                │
│                            ▼                                                │
│                     ┌──────────────┐                                        │
│                     │   GRAFANA    │                                        │
│                     │              │                                        │
│                     │  :3030       │ ◄── You look here                     │
│                     └──────────────┘                                        │
│                                                                             │
│                    OBSERVABILITY STACK                                      │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Quick Start

```bash
# Start the stack
cd observability-stack
docker-compose up -d

# Verify everything is running
docker-compose ps

# View logs if something fails
docker-compose logs otel-collector
```

## Access Points

| Service     | URL                   | Purpose                         |
| ----------- | --------------------- | ------------------------------- |
| **Grafana** | http://localhost:3030 | Main UI - traces, metrics, logs |
| Prometheus  | http://localhost:9090 | Metrics queries (PromQL)        |
| Tempo       | http://localhost:3200 | Trace API (use Grafana instead) |
| Loki        | http://localhost:3100 | Log API (use Grafana instead)   |

## Configure Your Services

In each microservice's main.go:

```go
import "github.com/yourproject/common/telemetry"

func main() {
    ctx := context.Background()

    shutdown, err := telemetry.Init(ctx, telemetry.Config{
        ServiceName:       "stats-service",
        ServiceVersion:    "1.0.0",
        Environment:       "development",
        CollectorEndpoint: "localhost:4317",  // OTel Collector
    })
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(ctx)

    // ... rest of your service
}
```

When running in Docker, use `otel-collector:4317` instead of `localhost:4317`.

## Viewing Data in Grafana

### View Traces

1. Open http://localhost:3030
2. Go to Explore (compass icon)
3. Select "Tempo" datasource
4. Search by trace ID or service name

### View Metrics

1. Go to Explore
2. Select "Prometheus" datasource
3. Use PromQL queries like `rate(http_requests_total[5m])`

### View Logs

1. Go to Explore
2. Select "Loki" datasource
3. Use LogQL like `{service_name="stats-service"}`

### The Magic: Correlation

When viewing a trace in Tempo, click "Logs for this span" to jump directly to related logs in Loki. This is why we set up the datasource links.

## Files

```
observability-stack/
├── docker-compose.yaml          # Runs everything
├── config/
│   ├── otel-collector.yaml      # Collector routing config
│   ├── tempo.yaml               # Trace storage config
│   ├── prometheus.yaml          # Metrics scraping config
│   ├── loki.yaml                # Log storage config
│   └── grafana-datasources.yaml # Auto-configures Grafana
└── common/
    └── telemetry/
        └── telemetry.go         # Copy to your project's common/
```

## Troubleshooting

### No traces appearing?

```bash
# Check collector is receiving data
docker-compose logs otel-collector | grep -i trace

# Verify your service connects to the right endpoint
# Should be: localhost:4317 (local) or otel-collector:4317 (in Docker)
```

### No metrics appearing?

```bash
# Check Prometheus targets
open http://localhost:9090/targets

# Should show otel-collector:8889 as UP
```

### Grafana not loading?

```bash
# Check all services are healthy
docker-compose ps

# Restart if needed
docker-compose restart grafana
```
