---
topic: gke-deployment
subtopic: overview
date: 2026-05-15
extracted-to-vault: ""
---

# Overview — 整體決策與成本規劃

## 起點需求

- 部署到 **GCP GKE**
- 用 **新帳號 $300 / 90 天免費額度**
- 目標：能跑滿 3 個月不爆預算

## 第一輪重大決策（部署計畫前置）

每個都用 AskUserQuestion 跟 user 確認過。

### 1. GKE 模式：Standard + on-demand 節點

| 選項 | 月費估算 | 為什麼選/不選 |
|---|---|---|
| **Standard + on-demand** ✓ | $73/月 (3 node) | user 選；穩定可控 |
| Standard + Spot | $30/月 | 60-91% 折扣但 Pod 隨時被搶 |
| Autopilot | $107/月 | 含 $74.4/月 control plane 費（free tier 不抵 Autopilot 控制面） |

**節點：3 × e2-medium**（2 vCPU、4 GB RAM、shared core）
- 8 微服務 × 100m = 800m CPU requests
- 加中介軟體 + frontend + ingress ≈ 1.3 vCPU 總 requests
- 4.5 vCPU allocatable（3 × 940m）足夠

### 2. 資料庫策略：全部跑在 cluster 內

| 選項 | 月費 | 為什麼選/不選 |
|---|---|---|
| **全跑 cluster (PostgreSQL StatefulSet)** ✓ | ~$0.04/GB-月 | user 選；學習用 |
| Cloud SQL db-f1-micro for Postgres | +$10/月 | 自動備份 + HA，但吃掉 1/3 學習額度 |
| 全部用託管（含 Memorystore Redis） | +$35/月 | 預算炸裂 |

### 3. Region：us-central1（Iowa）

- list price 最便宜（其他 region 多 10-30%）
- 從台灣 ping ~130-160ms（學習用無感）
- asia-east1 (彰化) 快但每月多 ~$15

### 4. Cluster 必須 **zonal** 不能 regional

- **GKE free tier**：每個 billing account 每月 $74.40 抵免，**只能抵 zonal Standard 或 Autopilot**
- regional cluster 直接被收 $73/月（3 個 zone 的 master 費）
- 對學習而言 zonal 失效率沒差別

### 5. Boot disk 一定要 `--disk-size=20`

- 預設 100 GB × 3 nodes × pd-balanced ($0.10/GB-月) = **多花 $30/月**
- 20 GB 對 GKE node OS + container image 來說綽綽有餘

## 90 天最終成本（含實際運作後調整）

| 項目 | 月費 USD | 備註 |
|---|---:|---|
| GKE 管理費（1 zonal cluster） | $0 | $74.40 free tier 全額抵 |
| Compute：3 × e2-medium | $73.38 | 偶爾 autoscaler 升到 4 nodes ~$98 |
| External HTTPS LB | $18.25 | 一個 LB 服 api + frontend |
| pd-balanced ~30 GB | $3.00 | PG PVC + node boot |
| Artifact Registry ~5 GB | $0.45 | 9 個 image × ~80 MB |
| Egress 10 GB/月 | $1.20 | Cloudflare proxy 後可能更低 |
| Cloud DNS 1 zone | $0.20 | 沒實際用，DNS 在 Cloudflare |
| Logging / Monitoring | $0 | 50 GiB Logging 免費 |
| **每月合計** | **~$96** | |
| **90 天總計** | **~$289** | 落在 $300 內，餘裕 ~$11 |

## 規劃時驗證過的 $300 條款（2026 確認）

1. **額度 $300 / 90 天** — 仍然存在
2. **僅需信用卡驗證身分** — 額度用完帳號自動 close，不會自動扣款（除非主動升級）
3. **升級成付費後 Always Free tier 仍有效**

## 為什麼分這麼多 phase

| Phase | 內容 | 為什麼分這裡 |
|---|---|---|
| 0 | 修代碼 + 補 manifests | 不修這些光部署也起不來，要先做完才碰 cloud |
| 1 | GCP project + Artifact Registry | 不能跳，後續所有事的前置 |
| 2 | Build & push 8 image | 純自動化但跑 15-25 分鐘要等 |
| 3 | 建 GKE cluster | 一指令但慢（5-8 分鐘） |
| 4 | Ingress + TLS 工具 | 跨多個架構決策（後來改 Cloudflare） |
| 5-6 | DNS + cert | user 動作 + 等 propagation |
| 7 | Apply manifests | 最容易踩雷（image pull 權限、migration 撞車） |
| 8 | 前端 | 跟 backend 解耦，可單獨 reroll |
