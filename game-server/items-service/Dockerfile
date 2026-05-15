# =============================================================================
# items-service Dockerfile
# Build context = game-server/
#   cd game-server
#   docker build -f items-service/Dockerfile -t items-service:dev .
# =============================================================================

FROM golang:1.24-alpine AS builder

WORKDIR /workspace

COPY go.work go.work.sum ./
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
COPY common/go.mod common/go.sum ./common/
COPY example-service/go.mod example-service/go.sum ./example-service/
COPY game-service/go.mod game-service/go.sum ./game-service/
COPY items-service/go.mod items-service/go.sum ./items-service/
COPY notification-service/go.mod notification-service/go.sum ./notification-service/
COPY payment-service/go.mod payment-service/go.sum ./payment-service/
COPY stats-service/go.mod stats-service/go.sum ./stats-service/

WORKDIR /workspace/items-service
RUN go mod download

WORKDIR /workspace
COPY . .

WORKDIR /workspace/items-service
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /workspace/items-service-bin \
    ./cmd/server

FROM alpine:3.19

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /workspace/items-service-bin ./items-service
COPY --from=builder /workspace/items-service/migrations ./migrations

USER app

EXPOSE 7013

CMD ["./items-service"]
