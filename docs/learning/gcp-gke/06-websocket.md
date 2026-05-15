---
date: 2026-05-15
topic: gcp-gke
subtopic: 06-websocket
extracted-to-vault:
---

# 06 — WebSocket 部署：Frontend env、Multi-port Service 拆分、Ingress 路徑

## 問題

部署完前端後遊戲 scene 進不去：
```
WebSocket connection to 'ws://localhost:5555/game/ws?token=...' failed
WebSocket disconnected, code: 1006
```

`code 1006` = abnormal closure（沒收到 close frame 就斷）。

## 根因 1：前端寫死 `ws://localhost:5555`

3 個檔案有寫死：
- `src/scenes/BootScene.ts:57`
- `src/scenes/GameScene.ts:50`
- `src/scenes/CosmicVoidScene.ts:2456`

### 修法：抽 helper + 用 env var

```ts
// src/utils/wsUrl.ts
export function getWsBaseUrl(): string {
  return process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:5555";
}
```

3 個檔案改用：
```ts
import { getWsBaseUrl } from "@/utils/wsUrl";
// ...
socketManager.connect(`${getWsBaseUrl()}/game/ws?token=${token}&name=${name}`);
```

### Dockerfile 必須宣告 ARG

```dockerfile
ARG NEXT_PUBLIC_WS_URL=wss://cosmicvoid.uk
ENV NEXT_PUBLIC_WS_URL=$NEXT_PUBLIC_WS_URL
```

少了 `ARG` 那行，`docker buildx build --build-arg NEXT_PUBLIC_WS_URL=wss://...` 看似有傳但 build 過程不會使用 → 最終 JS bundle 還是 fallback `ws://localhost:5555`。**踩過 1 次**。

驗證：
```bash
HTML=$(curl -sS https://cosmicvoid.uk/game)
JS=$(echo "$HTML" | grep -oE '/_next/static/chunks/[a-zA-Z0-9-]+\.js' | head -1)
curl -sS "https://cosmicvoid.uk$JS" | grep -o "wss://cosmicvoid.uk"  # 應 match
curl -sS "https://cosmicvoid.uk$JS" | grep -o "localhost:5555"       # 應 empty
```

## 根因 2：Ingress 沒路徑路由到 game-service

game-service 跑在 cluster 內 5555 port，但 Ingress 只有：
- `api.cosmicvoid.uk/*` → api-gateway
- `cosmicvoid.uk/*` → game-client

→ `cosmicvoid.uk/game/ws` 被當成 game-client 的路徑 → Next.js 回 404 HTML。

### 修法：Ingress 加 `/game/ws` 路徑

```yaml
- host: cosmicvoid.uk
  http:
    paths:
    - path: /game/ws        # 更具體的 path 必須擺前面（雖然 GCE LB 用 longest-prefix-match 自動處理，順序為了清晰）
      pathType: Prefix
      backend:
        service:
          name: game-service-ws
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

## 根因 3：game-service 是 multi-port Service → GCE Ingress controller 拒絕

game-service Service 同時 expose：
- `grpc` port 7004（intra-cluster，service discovery）
- `http` port 5555（WS endpoint）

加完 NEG annotation `{"ingress": true}` 後 controller 還是回報：
```
Translation failed: invalid ingress spec: service "cosmic-void/game-service"
is type "ClusterIP", expected "NodePort" or "LoadBalancer" when not using NEGs
```

雖然 5555 有 NEG，但 controller **檢查全部 ports**，看到 7004 沒 NEG 就罷工。`{"ingress": true}` 只給 Ingress 引用的 port 建 NEG，沒被 Ingress 引用的 port 不會自動有 NEG。

### 修法：拆 Service

```yaml
# game-service (留給 gRPC，cluster 內服務發現用)
apiVersion: v1
kind: Service
metadata:
  name: game-service
spec:
  type: ClusterIP
  selector:
    component: game-service
  ports:
  - name: grpc
    port: 7004
    targetPort: grpc

---

# game-service-ws (給 LB 用，有 NEG)
apiVersion: v1
kind: Service
metadata:
  name: game-service-ws
  annotations:
    cloud.google.com/neg: '{"ingress": true}'
    cloud.google.com/backend-config: '{"default": "game-service-ws-backend-config"}'
spec:
  type: ClusterIP
  selector:
    component: game-service   # 兩個 Service 都選同一個 Pod
  ports:
  - name: http
    port: 5555
    targetPort: http
```

更新 Ingress 引用 `game-service-ws`（不是 game-service）。

```bash
service/game-service configured
service/game-service-ws created
ingress.networking.k8s.io/cosmic-void-ingress configured
```

之後 backends 顯示：
```
"k8s1-...-cosmic-void-game-service-5555-...": "HEALTHY"
```

## BackendConfig 必須給長 timeout（WS 長連線）

GCE LB BackendService **預設 timeoutSec=30 秒**，會把長 idle WS connection 砍掉。

```yaml
apiVersion: cloud.google.com/v1
kind: BackendConfig
metadata:
  name: game-service-ws-backend-config
spec:
  timeoutSec: 86400              # 24 小時
  connectionDraining:
    drainingTimeoutSec: 60
```

## Cloudflare WebSocket 注意事項

- **Free tier 支援 WebSocket**（不用付費）✓
- **100 秒 idle timeout**：CF→Origin 那段，沒收發訊息 100s 會斷。遊戲循環頻繁送資料就無感
- **Flexible SSL** 模式：User 拿 wss://（Cloudflare cert），CF→Origin 是 HTTP/ws://（明文但走 CF 私網）

## 最終流量路徑

```
Browser
   │ wss://cosmicvoid.uk/game/ws?token=...
   ↓
Cloudflare edge (TLS termination)
   │ HTTP Upgrade: websocket
   ↓
GCP HTTPS LB :80
   │ URL Map: host=cosmicvoid.uk, path=/game/ws → game-service-ws backend
   ↓
NEG endpoint = Pod IP:5555
   │
   ↓
game-service Pod (Go gin server 在 :5555)
   │ HTTP 101 Switching Protocols
   ↓
WebSocket established
```

## 驗證

```bash
# Backends HEALTHY？
kubectl -n cosmic-void describe ingress cosmic-void-ingress | grep -A 1 "backends:"
# 應該看到 game-service-5555 backend HEALTHY

# URL Map 有 /game/ws？
gcloud compute url-maps describe k8s2-um-... \
  --format="yaml(pathMatchers[].pathRules[].paths)"
# 應該看到 /game/ws 路徑

# 手動 WS handshake test（這個會失敗回 400，正常，因為沒帶真 token）
curl -H "Connection: Upgrade" -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Key: $(echo -n test | base64)" \
  -H "Sec-WebSocket-Version: 13" \
  https://cosmicvoid.uk/game/ws
# HTTP 400 = 通到 game-service 了（不是 404 = 路由錯）
```

## 學習點

| 點 | 教訓 |
|---|---|
| Hardcode URLs in frontend | 從一開始就抽 helper，別在 3 個地方 copy paste |
| `NEXT_PUBLIC_*` env 注入 | Build-time inline，必須 `ARG` 在 Dockerfile，不是 runtime |
| Multi-port Service + Ingress | 簡單的辦法是拆 Service，比塞 annotation hacks 清楚 |
| GCE LB timeoutSec 預設 30s | 任何長連線（WS / SSE / long polling）都記得設 BackendConfig |
| Cloudflare proxy + WS | 免費 plan 就支援，但有 100s idle timeout |
