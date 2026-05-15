---
topic: gke-deployment
subtopic: websocket-routing
date: 2026-05-15
extracted-to-vault: ""
---

# WebSocket 經過 LB 完整路徑

User 打開 `cosmicvoid.uk/game`，console 跳：

```
WebSocket connection to 'ws://localhost:5555/game/ws?token=...' failed
WebSocket disconnected, code: 1006
```

前端寫死 `ws://localhost:5555` — 開發時跟 game-service 本機跑沒問題，prod 顯然 broken。

## 1. 前端：集中 WS URL helper

把 hardcoded 散落 3 個檔案：
- `src/scenes/BootScene.ts:57`
- `src/scenes/GameScene.ts:50`
- `src/scenes/CosmicVoidScene.ts:2456`

集中到一個 helper：

```typescript
// src/utils/wsUrl.ts
export function getWsBaseUrl(): string {
  return process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:5555";
}
```

3 個檔案改：

```diff
- socketManager.connect(`ws://localhost:5555/game/ws?token=${token}&name=${name}`);
+ socketManager.connect(`${getWsBaseUrl()}/game/ws?token=${token}&name=${name}`);
```

且每個檔加 `import { getWsBaseUrl } from "@/utils/wsUrl";`。

## 2. Dockerfile：宣告 build arg

`NEXT_PUBLIC_*` env vars 是 **build time inline** 到 JS bundle（不是 runtime）。所以要在 build 時注入：

```dockerfile
ARG NEXT_PUBLIC_API_URL=https://api.cosmicvoid.uk
ARG NEXT_PUBLIC_WS_URL=wss://cosmicvoid.uk          # ← 加這行
ARG NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=""
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_WS_URL=$NEXT_PUBLIC_WS_URL          # ← 加這行
ENV NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=$NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
```

Build：

```bash
docker buildx build \
  --platform linux/amd64 \
  --build-arg NEXT_PUBLIC_API_URL=https://api.cosmicvoid.uk \
  --build-arg NEXT_PUBLIC_WS_URL=wss://cosmicvoid.uk \
  --build-arg NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY="" \
  -t $IMG_BASE/game-client:v2 \
  --push .
```

驗證：

```bash
HTML=$(curl -sS https://cosmicvoid.uk/game)
JS_PATH=$(echo "$HTML" | grep -oE '/_next/static/chunks/[a-zA-Z0-9-]+\.js' | head -1)
curl -sS "https://cosmicvoid.uk$JS_PATH" | grep "wss://cosmicvoid.uk" && echo "✓"
curl -sS "https://cosmicvoid.uk$JS_PATH" | grep "localhost:5555" && echo "✗ 還有 hardcoded"
```

## 3. K8s 端：game-service 拆成兩個 Service

### 問題：multi-port + NEG 不一致 → GCE Ingress refuse

game-service 一開始：

```yaml
ports:
- name: grpc
  port: 7004  # intra-cluster
- name: http
  port: 5555  # WebSocket
```

加 NEG annotation 只給 5555：

```yaml
annotations:
  cloud.google.com/neg: '{"exposed_ports": {"5555": {}}}'  # 也試過 {"ingress": true}
```

GCE Ingress controller log：

```
Translation failed: invalid ingress spec: service "cosmic-void/game-service" is type
"ClusterIP", expected "NodePort" or "LoadBalancer" when not using NEGs
```

**這個錯誤訊息誤導**：5555 port 確實有 NEG。但 controller 因為 7004 port 沒 NEG 就拒絕整個 Service。

### 解法：拆 Service

```yaml
# game-service —— 純 gRPC（intra-cluster service discovery 用）
apiVersion: v1
kind: Service
metadata:
  name: game-service
spec:
  ports:
  - name: grpc
    port: 7004
---
# game-service-ws —— WebSocket（給 Ingress 用，有 NEG）
apiVersion: v1
kind: Service
metadata:
  name: game-service-ws
  annotations:
    cloud.google.com/neg: '{"ingress": true}'
    cloud.google.com/backend-config: '{"default": "game-service-ws-backend-config"}'
spec:
  ports:
  - name: http
    port: 5555
```

兩個 Service 共用同一個 Deployment（同一個 Pod 有兩個 port）。selector 相同。

**重要**：`serviceMap` in [common/discovery/k8s/k8s.go](../../game-server/common/discovery/k8s/k8s.go) 不需要改，它仍指向 `game-service:7004`（gRPC）。Ingress 指向 `game-service-ws:5555`（WS）。

## 4. BackendConfig：長 timeout 給 WS

GCE LB 預設 backend timeout 30s — WebSocket 長連線會被殺。

```yaml
apiVersion: cloud.google.com/v1
kind: BackendConfig
metadata:
  name: game-service-ws-backend-config
spec:
  timeoutSec: 86400          # 24 小時（GCE 上限）
  connectionDraining:
    drainingTimeoutSec: 60   # graceful shutdown 給 60s
```

## 5. Ingress：加 /game/ws path

```yaml
- host: cosmicvoid.uk
  http:
    paths:
    - path: /game/ws
      pathType: Prefix
      backend:
        service:
          name: game-service-ws    # ← 新 service
          port:
            name: http
    - path: /
      pathType: Prefix
      backend:
        service:
          name: game-client
          port:
            name: http
```

## 6. Cloudflare：自動支援 WS

WS 經過 Cloudflare proxy：

```
Browser
  │ 1. WSS handshake: GET / HTTP/1.1 + Upgrade: websocket
  │
  ↓ wss://cosmicvoid.uk/game/ws
Cloudflare edge (terminate TLS)
  │
  │ 2. HTTP request to origin (Flexible mode = HTTP)
  ↓ http://34.111.74.79/game/ws
GCP HTTPS LB
  │
  │ 3. URL Map: /game/ws → game-service-ws backend
  ↓
NEG endpoint → Pod IP:5555
  │
  │ 4. game-service 回 101 Switching Protocols
  │ 5. 升級成 WS tunnel
  ↓
Bidirectional WS frames
```

**Cloudflare Free plan 自動支援 WebSocket**，不需要額外設定。

### 限制：Cloudflare 100s idle timeout

WS 連線 100 秒沒收發 frame 會被 CF 強制斷。Active 遊戲循環不會碰到（持續送 player position），但 lobby 閒置會。

繞道：client 加 ping/pong heartbeat（30s 一次）。

## 7. 驗證

```bash
# WS handshake 從外面打（模擬 client）
curl -sS \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Key: dGVzdA==" \
  -H "Sec-WebSocket-Version: 13" \
  -o /dev/null -w "HTTP %{http_code}\n" \
  https://cosmicvoid.uk/game/ws

# 400 = 通了（game-service 看到 handshake，因為我 fake 的 key 不合法所以拒絕，
#         但 LB 路由是對的）
# 404 = path 沒在 LB URL Map → 看 Ingress
# 502 = backend UNHEALTHY → 看 BackendConfig
# 503 = LB 還在 provision → 等
```

從**真實 browser** 帶有效 token 跑，應該看到 101 Switching Protocols → connected。

## 教訓

1. **Multi-port Service + 部分 port 有 NEG** 在 GCE Ingress 是地雷，**拆 Service 是最乾淨解**
2. **NEXT_PUBLIC_* 是 build-time 不是 runtime**，所有 env 注入要在 docker build 階段
3. **GCE LB 預設 30s backend timeout** 對 WS 致命，要 BackendConfig timeoutSec=86400
4. **Cloudflare WS 100s idle** 要 heartbeat，否則閒置 client 會掉線
5. **`Sec-WebSocket-Key` 驗證**：curl 簡單測 HTTP 400 = 路由對的證據
