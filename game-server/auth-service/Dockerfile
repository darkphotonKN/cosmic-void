# =============================================================================
# auth-service Dockerfile
# Build context = game-server/ (whole workspace)
#   cd game-server
#   docker build -f auth-service/Dockerfile -t auth-service:dev .
# =============================================================================

FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Stage 1: copy go.work + every module's go.mod/go.sum so `go mod download`
# can resolve via workspace mode (cross-module imports resolve locally).
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

WORKDIR /workspace/auth-service
RUN go mod download

# Stage 2: copy the rest of the workspace (.dockerignore filters tmp/, k8s/, docs/).
WORKDIR /workspace
COPY . .

WORKDIR /workspace/auth-service
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /workspace/auth-service-bin \
    ./cmd/server

FROM alpine:3.19

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /workspace/auth-service-bin ./auth-service
COPY --from=builder /workspace/auth-service/migrations ./migrations

USER app

EXPOSE 7003 8081

CMD ["./auth-service"]
