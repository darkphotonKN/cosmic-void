---
date: 2026-05-15
topic: gcp-gke
subtopic: _index
extracted-to-vault:
---

# GCP GKE 部署 cosmic-void 完整紀錄

把 8 個 Go 微服務 + Next.js 前端從本地 docker-compose 部署到 GCP GKE Standard cluster，TLS 由 Cloudflare proxy 處理。$300 / 90 天免費額度。

## 學習路徑（按順序讀）

1. [01-foundation.md](01-foundation.md) — GCP 帳號、Billing、Project、Artifact Registry
2. [02-build-images.md](02-build-images.md) — Docker buildx 跨平台、Dockerfile pattern、.dockerignore 踩坑
3. [03-cluster.md](03-cluster.md) — GKE Standard zonal cluster、IAM 為什麼要手動加 artifactregistry.reader
4. [04-k8s-deploy.md](04-k8s-deploy.md) — Manifest 結構、Secrets 自動產生、Per-service DB 隔離、Postgres extensions
5. [05-ingress-tls.md](05-ingress-tls.md) — GCE Ingress + ManagedCert → 切換 Cloudflare proxy + Flexible SSL
6. [06-websocket.md](06-websocket.md) — Frontend 環境變數注入、Multi-port Service 拆分、WebSocket 路由
7. [07-debugging.md](07-debugging.md) — 過程中所有 incident 的因果鏈
8. [08-firewall-security.md](08-firewall-security.md) — Firewall 稽查、Cloud Armor 為何必要、剩下的安全強化選項
9. [_TODO.md](_TODO.md) — 還沒做完的任務

## 高層架構圖（最終狀態）

```
                    User Browser
                         │ HTTPS
                         ↓
              Cloudflare (Free Plan, Proxied 橘雲)
              ├ TLS termination (Universal cert)
              ├ DDoS 邊緣防護
              └ Free CDN
                         │ HTTP (Flexible mode)
                         ↓
              GCP HTTPS Load Balancer (34.111.74.79)
              ├ URL Map (host + path routing)
              └ Backend Services (with NEG)
                         │ HTTP
                         ↓
              GKE Standard zonal cluster (us-central1-a)
              ├ 3 × e2-medium on-demand nodes
              ├ ingress-nginx? 沒裝 — GCE Ingress 取代
              └ cert-manager? 沒裝 — Cloudflare TLS 取代

              cosmic-void namespace:
                api-gateway (HTTP 7001) ──┐
                game-client (HTTP 3000)   │
                game-service ─────────────┤
                  ├ gRPC 7004 (intra)     │
                  └ HTTP 5555 (WS) ───────┤  ← ws/game/ws path
                auth × 2, items, stats,   │
                payment, notif, example   │
                otel-collector (4317)     │
                postgres (StatefulSet)    │
                redis, rabbitmq           │
                                          ↓
              SERVICE DISCOVERY: k8s native DNS
              <svc>.cosmic-void.svc.cluster.local
              (Consul 已移除，見 CONSUL_TO_K8S_MIGRATION.md)
```

## 域名

- `cosmicvoid.uk` → game-client（前端）
- `www.cosmicvoid.uk` → game-client
- `api.cosmicvoid.uk` → api-gateway（HTTP API）
- `cosmicvoid.uk/game/ws` → game-service:5555（WebSocket）

## 成本實況（90 天估算）

| 項目 | 月費 |
|---|---:|
| GKE 管理費（1 zonal cluster） | $0（free tier 抵 $74.40） |
| 3× e2-medium on-demand | $73.38 |
| GCP HTTPS LB | $18.25 |
| PVC + boot disks ~30GB | ~$3 |
| Artifact Registry ~7GB | ~$0.65 |
| Egress（部份走 Cloudflare 省了） | ~$1 |
| Cloud DNS | $0（用 Cloudflare 不開 Cloud DNS）|
| OTEL Collector | 含在 cluster 內 |
| **預計月費** | **~$96** |
| **預計 90 天** | **~$289**（含 ~$10 buffer） |

待做：Cloud Armor + Cloudflare IP allowlist 會多 ~$6/月。
