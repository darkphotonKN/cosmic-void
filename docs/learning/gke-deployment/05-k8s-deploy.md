---
topic: gke-deployment
subtopic: k8s-deploy
date: 2026-05-15
extracted-to-vault: ""
---

# Phase 7 — 套用 k8s manifests

8 個 service + middleware + ingress 套上 cluster。

## 7.1 建 namespace + 切 context

```bash
kubectl create namespace cosmic-void --dry-run=client -o yaml | kubectl apply -f -
kubectl config set-context --current --namespace=cosmic-void
```

**用 dry-run + apply 取代 kubectl create**：可重複跑不會報錯（idempotent）。

## 7.2 用 openssl 產生隨機密碼

```bash
DB_PASS=$(openssl rand -base64 24 | tr -d '/+=' | head -c 24)
RABBIT_PASS=$(openssl rand -base64 24 | tr -d '/+=' | head -c 24)
JWT_SECRET=$(openssl rand -base64 48 | tr -d '/+=' | head -c 48)

# 存本機 0600 給之後用
cat > ~/.cosmic-void-secrets.env <<EOF
export DB_PASS='$DB_PASS'
export RABBIT_PASS='$RABBIT_PASS'
export JWT_SECRET='$JWT_SECRET'
EOF
chmod 600 ~/.cosmic-void-secrets.env
```

`tr -d '/+='` 把 base64 裡的特殊字元去掉（避免 URL escape 麻煩），`head -c 24` 切 24 字元。

## 7.3 為什麼用 `kubectl create secret --dry-run | kubectl apply`

```bash
kubectl create secret generic auth-service-secrets -n cosmic-void \
  --from-literal=DB_PASSWORD="$DB_PASS" \
  --from-literal=RABBITMQ_PASS="$RABBIT_PASS" \
  --from-literal=JWT_SECRET="$JWT_SECRET" \
  --dry-run=client -o yaml | kubectl apply -f -
```

| 寫法 | 重跑 |
|---|---|
| `kubectl create secret` | 第二次跑會報「already exists」 |
| `kubectl create --dry-run=client -o yaml \| kubectl apply` | ✅ idempotent，相同值 = unchanged |

## 7.4 批次 patch 8 個 deployment 的 image 路徑

manifests 寫 `image: cosmic-void/auth-service:dev`（dev 環境），要改成 Artifact Registry path：

```bash
IMG_BASE="us-central1-docker.pkg.dev/$PROJECT_ID/$REPO"

find . -path '*/k8s/deployment.yml' -print0 | while IFS= read -r -d '' f; do
  if grep -q "image: cosmic-void/" "$f"; then
    sed -i '' "s|image: cosmic-void/\\([a-z-]*\\):dev|image: $IMG_BASE/\\1:v1|g" "$f"
  fi
done
```

正則 `cosmic-void/\\([a-z-]*\\):dev` 抓住 service 名稱再代換到完整路徑。

## 7.5 部署順序：先中介軟體再 service

```bash
# 1. PostgreSQL StatefulSet 必須先起，否則 service migration 會卡
kubectl apply -n cosmic-void -f auth-service/k8s/postgres.yml
kubectl apply -n cosmic-void -f auth-service/k8s/redis.yml
kubectl apply -n cosmic-void -f auth-service/k8s/rabbitmq.yml

# 等 postgres ready（最關鍵）
kubectl -n cosmic-void wait --for=condition=ready pod -l component=auth-service-db --timeout=240s
kubectl -n cosmic-void wait --for=condition=ready pod -l component=redis --timeout=60s
kubectl -n cosmic-void wait --for=condition=ready pod -l component=rabbitmq --timeout=180s

# 2. ConfigMaps + Deployments
for svc in auth-service api-gateway items-service game-service \
           notification-service payment-service stats-service example-service; do
  kubectl apply -n cosmic-void \
    -f $svc/k8s/deployment.yml \
    -f $svc/k8s/service.yml \
    -f $svc/k8s/configmap.yml
done

# 3. Ingress + Cert
kubectl apply -n cosmic-void -f api-gateway/k8s/managedcertificate.yml
kubectl apply -n cosmic-void -f api-gateway/k8s/ingress.yml
```

**Service 之間不嚴格依賴啟動順序**（k8s native DNS 解析會 retry），但 PG 沒起來業務 service 會 crashloop migration。所以 PG 必須先 Ready。

## 7.6 第一次 deploy 全炸：ImagePullBackOff

```
Failed to pull image "us-central1-docker.pkg.dev/.../auth-service:v1": 
failed to authorize: 403 Forbidden
```

→ 直接看 [02-gcp-setup.md](02-gcp-setup.md) 的 1.6 節：default Compute SA 沒 artifactregistry.reader。

## 7.7 第二次：CrashLoopBackOff（migration 撞車）

→ 看 [06-db-per-service.md](06-db-per-service.md)。

## 7.8 工具：force restart deployment

修了 Service / ConfigMap 不會自動 rollout，要明確觸發：

```bash
kubectl -n cosmic-void rollout restart deployment <name>
# 或所有
kubectl -n cosmic-void rollout restart deployment
```

**為什麼不能改 image tag 後直接 apply？** 同 tag 同 image，`imagePullPolicy: IfNotPresent` 不會重拉。所以一般要嘛換 tag（推薦），要嘛改 imagePullPolicy: Always + rollout restart。
