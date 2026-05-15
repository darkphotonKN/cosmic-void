---
topic: gke-deployment
subtopic: gcp-setup
date: 2026-05-15
extracted-to-vault: ""
---

# Phase 1 — GCP Project + Artifact Registry

## 1.1 本機工具

```bash
# Google Cloud SDK（macOS via brew）
brew install --cask google-cloud-sdk

# kubectl 認 GKE 要的 plugin
gcloud components install gke-gcloud-auth-plugin
# brew 裝的 gcloud 把 plugin 放在 /opt/homebrew/share/google-cloud-sdk/bin/
# 但 /opt/homebrew/bin 沒 symlink → kubectl 找不到，要補：
ln -sf /opt/homebrew/share/google-cloud-sdk/bin/gke-gcloud-auth-plugin /opt/homebrew/bin/

# 其他
brew install kubectl helm
docker --version  # 需要 Docker Desktop + buildx
```

## 1.2 GCP 帳號設定（user 在瀏覽器做）

1. https://cloud.google.com/free → 用沒用過免費試用的 Google 帳號
2. 填信用卡（**只驗證身分**，不會自動扣款）
3. Billing console 建立 Billing account
4. Console → Billing → Budgets & alerts：$50 / $150 / $250 / $290 四階 alert

## 1.3 gcloud auth + project

```bash
gcloud auth login
gcloud auth list  # 確認 active account

# 抓 billing account ID（會是類似 01CB86-16B54C-97FA21 格式）
gcloud billing accounts list
```

### 踩雷：建新 project 撞 quota

```bash
gcloud projects create cosmic-void-$(date +%s) --name="Cosmic Void"
# ERROR: RESOURCE_EXHAUSTED: Quota exceeded for quota metric 'Write requests'
```

新帳號剛開通時 cloudresourcemanager API 的 write quota 很緊（gcloud 默認的 quota project = `32555940559`）。等 1-60 分鐘 quota reset 才能再試。

**繞道方案**：直接用 GCP 自動配的 default project（新帳號開通時自動建一個叫「My First Project」的，project ID 是 `project-XXXX-XXXX-XXX` 格式）。

```bash
gcloud projects list
# PROJECT_ID                      NAME              PROJECT_NUMBER
# project-b6e8596f-fe1e-4dca-a7a  My First Project  450898188368

PROJECT_ID="project-b6e8596f-fe1e-4dca-a7a"

# 改個顯示名稱
gcloud projects update $PROJECT_ID --name="Cosmic Void"

# 設為 active
gcloud config set project $PROJECT_ID

# Link billing
gcloud billing projects link $PROJECT_ID --billing-account=01CB86-16B54C-97FA21
```

**取捨**：用 default project ID 比較醜（`project-b6e8...`），但**功能完全一樣**且省一輪等待。

## 1.4 啟用 5 個必要 API

```bash
gcloud services enable \
  container.googleapis.com \
  artifactregistry.googleapis.com \
  compute.googleapis.com \
  dns.googleapis.com \
  cloudbuild.googleapis.com
```

| API | 為什麼需要 |
|---|---|
| container | GKE cluster 建立 |
| artifactregistry | 推送 docker image |
| compute | Cluster VM、LB、firewall、static IP |
| dns | Cloud DNS（後來改 Cloudflare 沒用到，但啟用無成本） |
| cloudbuild | 雖然 build 用本機 docker buildx，但 gcloud 部分指令需要 |

## 1.5 Artifact Registry（取代 Container Registry）

```bash
REGION="us-central1"
REPO="cosmic-void"

gcloud artifacts repositories create $REPO \
  --repository-format=docker \
  --location=$REGION \
  --description="cosmic-void container images"

# 設 docker daemon 認證
gcloud auth configure-docker $REGION-docker.pkg.dev --quiet
```

image 路徑會是：
```
us-central1-docker.pkg.dev/<PROJECT_ID>/cosmic-void/<svc>:<tag>
```

| 概念 | 對應 AWS |
|---|---|
| Artifact Registry | ECR |
| Container Registry (舊版) | ECR (legacy) |

**為什麼用 Artifact Registry 不是 Container Registry**：Google 2024 後把 Container Registry deprecate 了，新 project 預設 Artifact Registry。

## 1.6 踩雷：default Compute SA 沒 artifactregistry.reader

新 GCP project（2024 後）的 default Compute Engine SA **不再自動有 roles/editor**。GKE node 用這個 SA 拉 image，會被 403 reject。

### 症狀

```
Failed to pull image "us-central1-docker.pkg.dev/.../auth-service:v1": 
failed to authorize: 403 Forbidden
```

### 修法

```bash
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
DEFAULT_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$DEFAULT_SA" \
  --role="roles/artifactregistry.reader" \
  --condition=None
```

### 驗證

```bash
gcloud projects get-iam-policy $PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:$DEFAULT_SA AND bindings.role:roles/artifactregistry.reader"
# 應該回 "roles/artifactregistry.reader"
```

## 1.7 環境變數儲存（後續 phase 都會 source）

```bash
cat > ~/.cosmic-void-gcp.env <<EOF
export PROJECT_ID=$PROJECT_ID
export BILLING_ACCOUNT=$BILLING_ACCOUNT
export REGION=us-central1
export ZONE=us-central1-a
export REPO=cosmic-void
EOF

# 後續 phase 開頭都這樣 source
source ~/.cosmic-void-gcp.env
```
