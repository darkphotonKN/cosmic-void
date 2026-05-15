---
date: 2026-05-15
topic: gcp-gke
subtopic: 01-foundation
extracted-to-vault:
---

# 01 — GCP 帳號、Billing、Project、Artifact Registry

## 為什麼這樣走

- **$300 / 90 天免費額度**：新 GCP 帳號限定，需要綁信用卡（僅驗證身分，額度到期自動 close 不會扣款）
- **us-central1** region：所有 GCP 服務最早上線最便宜的地區，e2-medium $24.46/月（asia-east1 貴 10-15%）
- **Standard zonal cluster + 一般 on-demand nodes**：用戶選擇穩定性勝過 Spot 折扣

## 工具盤點

```bash
brew install --cask google-cloud-sdk  # gcloud
gcloud components install kubectl gke-gcloud-auth-plugin
brew install helm
# docker, buildx — Docker Desktop 內建
```

⚠ **gke-gcloud-auth-plugin** 裝完不在 PATH 的話補 symlink：
```bash
ln -sf /opt/homebrew/share/google-cloud-sdk/bin/gke-gcloud-auth-plugin \
       /opt/homebrew/bin/gke-gcloud-auth-plugin
```
驗證：`gke-gcloud-auth-plugin --version` 回 `Kubernetes v1.30.0+...`

沒這個 plugin 的話 kubectl 後續對 GKE 認證會失敗。

## 帳號 + Billing

1. https://cloud.google.com/free 用沒用過免費試用的 Google 帳號開
2. Billing → Create account → 填信用卡（只驗證）
3. **必做**：Budgets & alerts 設 $50 / $150 / $250 / $290 四階 alert

## Project 建立踩雷

```bash
gcloud auth login   # 瀏覽器互動，user 親手做
gcloud auth list    # 確認 ACTIVE account
gcloud billing accounts list  # 拿 billing account ID
```

**新帳號常踩的雷**：`gcloud projects create cosmic-void-<timestamp>` 會回 **RESOURCE_EXHAUSTED Quota exceeded**。原因是 cloudresourcemanager 對新帳號的 project creation rate limit 很嚴格。

**解法**：用新帳號自動配的 default project（名字像 `project-b6e8596f-fe1e-4dca-a7a` 醜但能用）。

```bash
gcloud projects list  # 找 default project
PROJECT_ID="project-b6e8596f-fe1e-4dca-a7a"

gcloud config set project "$PROJECT_ID"
gcloud projects update "$PROJECT_ID" --name="Cosmic Void"   # 改 display name
gcloud billing projects link "$PROJECT_ID" --billing-account="01CB86-..."

gcloud services enable \
  container.googleapis.com \
  artifactregistry.googleapis.com \
  compute.googleapis.com \
  dns.googleapis.com \
  cloudbuild.googleapis.com
```

## Artifact Registry（取代已 deprecated 的 Container Registry）

```bash
REGION="us-central1"
REPO="cosmic-void"

gcloud artifacts repositories create "$REPO" \
  --repository-format=docker \
  --location=$REGION \
  --description="cosmic-void container images"

# 讓本機 docker push 認證走 gcloud auth
gcloud auth configure-docker "$REGION-docker.pkg.dev" --quiet
```

URL 格式：`us-central1-docker.pkg.dev/<PROJECT_ID>/<REPO>/<IMAGE>:<TAG>`

## 把環境變數持久化（重要學習）

我們建了 `~/.cosmic-void-gcp.env` 讓後續所有 phase 都能 source：

```bash
cat > ~/.cosmic-void-gcp.env <<EOF
export PROJECT_ID=project-b6e8596f-fe1e-4dca-a7a
export BILLING_ACCOUNT=01CB86-16B54C-97FA21
export REGION=us-central1
export ZONE=us-central1-a
export REPO=cosmic-void
EOF
```

之後每個 bash session：`source ~/.cosmic-void-gcp.env`

## 預設 Compute SA 與 IAM 的隱形雷

**新版 GCP project** 的 default Compute Engine SA `<PROJECT_NUMBER>-compute@developer.gserviceaccount.com` 不會自動配 `roles/editor`。這在後面 GKE node 拉 Artifact Registry image 時會炸（見 [03-cluster.md](03-cluster.md)）。

## 驗證

```bash
gcloud artifacts repositories list --location=$REGION  # 看到 cosmic-void
gcloud services list --enabled | grep -E "(container|artifact|compute)"
```

## 為什麼這個選擇而不是其他

| 替代方案 | 為什麼沒選 |
|---|---|
| Cloud Run（serverless）| 不支援 StatefulSet 跑 Postgres；對微服務 mesh 不友善 |
| GKE Autopilot | 控制平面 $74.40/月（GKE Free Tier 在 Autopilot 是抵 vCPU/RAM 計費）；8 service + middleware 估算 ~$107/月，超過預算 |
| GKE Standard + Spot | Spot 隨時可能被 preempt，學習階段 debug 麻煩 |
| **GKE Standard + on-demand**（採用）| 穩定 + 享 $74.40/月 zonal free tier |
