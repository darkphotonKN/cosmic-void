# =============================================================================
# auth-service Dockerfile
#
# Build context 必須在 game-server/（因為 go.work + replace ../common）
# Build 指令：
#   cd game-server
#   docker build -f auth-service/Dockerfile -t auth-service:dev .
# =============================================================================

# ============================================================
# Stage 1: Builder - 編譯 Go binary
# ============================================================
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# 注意：不複製 go.work
# 原因：go.work 列了所有 9 個 service，但這個 image 只 build auth-service。
# 為了避免 go.work 要求其他 service 存在，build 時用單個 module 模式。
# 透過 GOWORK=off 環境變數忽略任何 go.work（即使父目錄有也忽略）。

# 先複製 module 描述檔（Layer cache 友善）
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
COPY common/go.mod common/go.sum ./common/

# 下載依賴（只在 go.mod 改變時才重跑）
WORKDIR /workspace/auth-service
ENV GOWORK=off
RUN go mod download

# 複製原始碼
WORKDIR /workspace
COPY auth-service/ ./auth-service/
COPY common/ ./common/

# 編譯（CGO_ENABLED=0 純靜態，跑 Alpine 必要）
WORKDIR /workspace/auth-service
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /workspace/auth-service-bin \
    ./cmd/server

# ============================================================
# Stage 2: Runtime - 最小執行環境
# ============================================================
FROM alpine:3.19

# Non-root user（資安最佳實踐）
# addgroup / adduser 是 Alpine 內建指令，不需要 apk install
#
# 註：原本還會 apk add ca-certificates + tzdata，但目前環境 apk repo 連線有問題
# 對 cosmic-void 的影響：
#   - ca-certificates：只連內部 ClusterIP（明文），不需要
#                      未來加 Stripe / SendGrid 等 HTTPS API 時要補
#   - tzdata：deployment.yml 已用 hostPath 掛載 Asia/Taipei，已有解
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# 拷 binary 過來
COPY --from=builder /workspace/auth-service-bin ./auth-service

# 拷 migrations 過來（程式啟動時會跑 migrate.Up()）
COPY --from=builder /workspace/auth-service/migrations ./migrations

# 切到 non-root user
USER app

# 開放 port（純文件性質，不影響實際監聽）
EXPOSE 7003 8081

# 啟動
CMD ["./auth-service"]
