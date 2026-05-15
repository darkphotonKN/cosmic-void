---
topic: gke-deployment
date: 2026-05-15
extracted-to-vault: ""
---

# GKE 部署全程學習筆記

把 cosmic-void（8 個 Go 微服務 + Next.js 前端 + PostgreSQL/Redis/RabbitMQ）部署到 GCP GKE，使用 $300 / 90 天免費額度。本系列記錄完整流程、每個決策的理由、踩過的雷、以及指令對照。

## 最終架構

```
User ──HTTPS──► Cloudflare proxy ──HTTP──► GCP HTTPS LB (34.111.74.79)
                  (orange cloud)             │
                  (Flexible SSL)             │
                                             ↓
                                       GCE Ingress URL Map
                                       ├─ api.cosmicvoid.uk/*  → api-gateway:80
                                       ├─ cosmicvoid.uk/game/ws → game-service-ws:5555
                                       └─ cosmicvoid.uk/*      → game-client:80
                                             │
                                             ↓
                                       GKE Standard zonal cluster
                                       (us-central1-a, 3-4 × e2-medium)
                                       ├─ 8 Go microservices
                                       ├─ Next.js frontend
                                       ├─ PostgreSQL / Redis / RabbitMQ
                                       └─ OTEL Collector
```

## 章節索引

| 編號 | 主題 | 重點 |
|---|---|---|
| [00-overview](00-overview.md) | 架構 / 成本 / 整體決策 | $300 額度規劃、3 種 LB 方案比較 |
| [01-phase0-blockers](01-phase0-blockers.md) | Phase 0 必修 blockers | listener bind、補 k8s manifests |
| [02-gcp-setup](02-gcp-setup.md) | GCP project + Artifact Registry | project quota、IAM、SA roles |
| [03-docker-build-gotchas](03-docker-build-gotchas.md) | Docker buildx + go.work 踩雷 | dockerignore、tail -8 吃 exit code |
| [04-gke-cluster](04-gke-cluster.md) | GKE Standard zonal | autoscaler、boot disk、free tier |
| [05-k8s-deploy](05-k8s-deploy.md) | apply manifests | secrets 隨機產、image patching |
| [06-db-per-service](06-db-per-service.md) | DB 隔離 + pg 擴充 | uuid-ossp、dirty migration |
| [07-ingress-vs-cloudflare](07-ingress-vs-cloudflare.md) | TLS 路徑切換 | ManagedCert → Cloudflare Flexible |
| [08-websocket-routing](08-websocket-routing.md) | WS 經過 LB | Service split、NEG 格式坑 |
| [09-firewall-cloud-armor](09-firewall-cloud-armor.md) | 安全強化 | VPC firewall vs Cloud Armor |
| [10-troubleshooting](10-troubleshooting.md) | 踩雷集錦 + 通用 debug | image pull、cert provisioning |

## 主要學到的事

1. **VPC firewall 無法擋 L7 LB frontend IP** — Cloud Armor 是唯一辦法（Cloud Armor === GCP 版 AWS WAF）
2. **`tail -N` 吞掉 docker buildx 的 non-zero exit code** — script 看起來成功實際失敗
3. **go.work 模式下單 module build 必須複製所有 service 的 go.mod**（GOWORK=off 會踩 go.sum 缺 entry）
4. **dockerignore 用 `**/k8s/` 會誤刪 `common/discovery/k8s/`** — 共用代碼跟 manifests 目錄名衝突要小心
5. **GCE Ingress 對 multi-port ClusterIP Service + 部分 port 沒 NEG 會卡 Translation failed** — 拆 Service 是最乾淨解
6. **新 GCP project 的 default compute SA 不再自動有 roles/editor** — 要手動 grant artifactregistry.reader 給 GKE node
7. **L7 LB 是 proxy 不是 passthrough** — 源 IP 在 X-Forwarded-For，不在封包 source 上，VPC firewall 看不到
8. **Cloudflare orange-cloud (proxy) DNS 回 Cloudflare IP 不是 origin IP** — 不是「DNS 不通」，是設計使然
9. **Per-service Postgres DB 不分隔會 migration 撞車** — 共用 DB 雖然省事但會出現 dirty migration 鬼故事
10. **Cloud Armor === AWS WAF 的 GCP 對應**，~$6/月，1 policy + 1 rule

## 未完成任務（待我繼續做）

1. **Cloud Armor + Cloudflare IP 允許清單** — 鎖死「公網直接打 34.111.74.79」攻擊面（user 已 approve，預估 5-10 分鐘）
2. **Cluster max-nodes 4 → 3** — 已沒有 Pending pod，可調降省 ~$24/月

## 已知未做（user 之後可自行處理）

- **GKE Master Authorized Networks** — user 選擇保留 0.0.0.0/0（依賴 IAM auth）
- **WebSocket Cloudflare 100s idle timeout** — Free plan 限制，遊戲 active 時不會碰到
- **Stripe key 是 placeholder** — payment-service 起來但 API 一打就失敗，要正式用要填真 key
- **TypeScript 嚴格檢查暫時關掉** — `ignoreBuildErrors: true` in next.config.ts，等代碼乾淨再開回來
- **`common/constants/types/item.go` 反向 import `game-service/grpc/items`** — 靠 go.work 繞過，長期該重構
- **dial-per-call 模式** — 每個跨服務 gRPC 呼叫都建新連線，效能不夠時要改 `dns:///` + round_robin

## 相關文件

- 部署計畫源頭：`~/.claude/plans/gcp-gke-google-shimmying-dragonfly.md`
- 環境設定：`~/.cosmic-void-gcp.env`（PROJECT_ID、REGION、ZONE 等）
- 密碼：`~/.cosmic-void-secrets.env`（DB_PASS、RABBIT_PASS、JWT_SECRET，本機 0600）
- 之前的 Consul → k8s migration：`game-server/docs/CONSUL_TO_K8S_MIGRATION.md`
