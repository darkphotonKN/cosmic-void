---
date: 2026-05-15
topic: gcp-gke
subtopic: 04-k8s-deploy
extracted-to-vault:
---

# 04 — K8s Manifest、Secrets、Per-service DB、Postgres extensions

## Phase 0：部署前必修 Blocker

[CONSUL_TO_K8S_MIGRATION.md](../../../game-server/docs/CONSUL_TO_K8S_MIGRATION.md) 把 Consul 換成 k8s native DNS。但 Section 5 列了「部署 GKE 前必修」的兩件事：

### Blocker 1：所有 main.go 的 listener bind 在 `localhost`

```go
// Before — k8s Service 流量進不來
listener, err := net.Listen("tcp", "localhost:"+grpcAddr)

// After
listener, err := net.Listen("tcp", ":"+grpcAddr)
```

7 個 service 改（api-gateway 用 gin `router.Run(":port")` 已 bind all interfaces，不用改）：

```
auth-service/cmd/server/main.go
items-service/cmd/server/main.go
game-service/cmd/server/main.go
stats-service/cmd/server/main.go
notification-service/cmd/server/main.go
payment-service/cmd/server/main.go
example-service/cmd/server/main.go
```

### Blocker 2：補 6 個 service 的 k8s manifests

migration 後只有 `auth-service/k8s/` 完整（含 postgres、redis、rabbitmq）、`api-gateway/k8s/` 只有 service + ingress。其餘 6 個 service 完全沒 k8s manifest。

每個 service 補：
- `deployment.yml`（image、resource requests/limits、envFrom configmap+secret、livenessProbe）
- `service.yml`（ClusterIP，port 對齊 serviceMap）
- `configmap.yml`（DB host、RABBITMQ、REDIS、K8S_NAMESPACE）
- `secret.yml.example`（範本，密碼填 REPLACE_ME）

### 為什麼 Service name + port 必須跟 serviceMap 對齊

`game-server/common/discovery/k8s/k8s.go`：
```go
var serviceMap = map[string]serviceEntry{
    "auth":         {"auth-service", 7003},
    "payments":     {"payment-service", 7021},
    "items":        {"items-service", 7013},
    "stats":        {"stats-service", 7011},
    "notification": {"notification-service", 7077},
    "examples":     {"example-service", 7010},
    "game":         {"game-service", 7004},
    "api-gateway":  {"api-gateway", 7001},
}
```

`Discover()` 回 `<service>.<namespace>.svc.cluster.local:<port>`。如果 k8s Service 的 metadata.name 或 port 跟這對不上，gRPC dial 失敗。

## Namespace 策略

```bash
kubectl create namespace cosmic-void
kubectl config set-context --current --namespace=cosmic-void
```

**所有 manifest 不要寫 `namespace: default`**。原本 auth-service 的舊 manifest 有寫，會跟 `kubectl apply -n cosmic-void` 衝突。批次清掉：

```bash
find game-server -path '*/k8s/*.yml' -exec sed -i '' '/^  namespace: default$/d' {} \;
find game-server -type f \( -name '*.yml.example' -o -name '*.yml.template' \) \
  -path '*/k8s/*' -exec sed -i '' '/^  namespace: default$/d' {} \;
```

## Secrets — 不要手填，自動產

```bash
DB_PASS=$(openssl rand -base64 24 | tr -d '/+=' | head -c 24)
RABBIT_PASS=$(openssl rand -base64 24 | tr -d '/+=' | head -c 24)
JWT_SECRET=$(openssl rand -base64 48 | tr -d '/+=' | head -c 48)

# 存本機方便之後對照（mode 600 防外人看）
cat > ~/.cosmic-void-secrets.env <<EOF
export DB_PASS='$DB_PASS'
export RABBIT_PASS='$RABBIT_PASS'
export JWT_SECRET='$JWT_SECRET'
EOF
chmod 600 ~/.cosmic-void-secrets.env

# 每個 service 一個 Secret，dry-run + apply 確保 idempotent
kubectl create secret generic auth-service-secrets -n cosmic-void \
  --from-literal=DB_PASSWORD="$DB_PASS" \
  --from-literal=RABBITMQ_PASS="$RABBIT_PASS" \
  --from-literal=JWT_SECRET="$JWT_SECRET" \
  --dry-run=client -o yaml | kubectl apply -f -
# ... 對其他 service 重複 ...
```

Stripe 用 placeholder（payment-service 程式碼檢查 `STRIPE_SECRET_KEY != ""`，填空字串會 panic exit）：
```bash
--from-literal=STRIPE_SECRET_KEY="sk_test_PLACEHOLDER_REPLACE_ME"
--from-literal=STRIPE_WEBHOOK_SECRET="whsec_PLACEHOLDER_REPLACE_ME"
```

## Image refs 批次 patch

8 個 deployment.yml 的 image 從 `cosmic-void/<svc>:dev` 改成 Artifact Registry 路徑：

```bash
IMG_BASE="$REGION-docker.pkg.dev/$PROJECT_ID/$REPO"

find . -path '*/k8s/deployment.yml' -print0 | while IFS= read -r -d '' f; do
  if grep -q "image: cosmic-void/" "$f"; then
    sed -i '' "s|image: cosmic-void/\\([a-z-]*\\):dev|image: $IMG_BASE/\\1:v1|g" "$f"
  fi
done
```

## 部署順序（很重要）

```bash
# 1. middleware 先（PostgreSQL/Redis/RabbitMQ）
kubectl apply -n cosmic-void -f auth-service/k8s/postgres.yml
kubectl apply -n cosmic-void -f auth-service/k8s/redis.yml
kubectl apply -n cosmic-void -f auth-service/k8s/rabbitmq.yml

# 等 Ready
kubectl -n cosmic-void wait --for=condition=ready pod -l component=auth-service-db --timeout=240s

# 2. 8 個 service
for svc in auth-service api-gateway items-service game-service \
           notification-service payment-service stats-service example-service; do
  kubectl apply -n cosmic-void \
    -f "$svc/k8s/deployment.yml" \
    -f "$svc/k8s/service.yml" \
    -f "$svc/k8s/configmap.yml"
done

# 3. Ingress 最後（要先有 Service backend）
kubectl apply -n cosmic-void -f api-gateway/k8s/ingress.yml
```

## 第一次部署遇到的 DB Migration 問題

**症狀**：例如 example-service log：
```
Failed to run migrations: could not run migrations:
no migration found for version 15: read down for version 15: file does not exist
```

**根因**：起初所有 service configmap 都指向 `cosmic_void_auth_service_db` 同一個 DB。每個 service 都有自己的 migration 集合，跑起來時 migrate 工具讀 `schema_migrations` table，看到「v15」就要找對應 file，但其他 service 的 migrations 沒 v15 → 炸。

### 修法：Per-service Database

7 個 service（不含 api-gateway 它無狀態）各自一個 DB：

```bash
for db in cosmic_void_items_service_db cosmic_void_stats_service_db \
          cosmic_void_notification_service_db cosmic_void_payment_service_db \
          cosmic_void_example_service_db cosmic_void_game_service_db; do
  kubectl -n cosmic-void exec auth-service-db-0 -- \
    psql -U user -d cosmic_void_auth_service_db -c "CREATE DATABASE $db;"
done
```

然後 patch 對應 configmap：
```bash
sed -i '' 's|DB_NAME: "cosmic_void_auth_service_db"|DB_NAME: "cosmic_void_items_service_db"|' \
  items-service/k8s/configmap.yml
# ... 其他 5 個比照 ...
```

## 第二雷：v1 migration 卡 "dirty version 1"

某個 service v1 migration 用了 `uuid_generate_v4()` 但 PostgreSQL 17 預設**沒裝 `uuid-ossp` extension**。第一次跑 partial table 建好就掛 → `schema_migrations` table 留下 `(version=1, dirty=t)`。下次起 service 看到 dirty 就放棄。

### 修法：先裝 extensions，再 drop + recreate dirty DB

```bash
# 在所有 service DB 都裝 extension（pre-emptive）
for db in cosmic_void_auth_service_db cosmic_void_items_service_db ...; do
  kubectl -n cosmic-void exec auth-service-db-0 -- \
    psql -U user -d "$db" -c "
      CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";
      CREATE EXTENSION IF NOT EXISTS pgcrypto;
    "
done

# Drop + recreate stats DB（清掉 dirty state）
kubectl -n cosmic-void exec auth-service-db-0 -- psql -U user -d postgres -c "
  SELECT pg_terminate_backend(pid) FROM pg_stat_activity
  WHERE datname = 'cosmic_void_stats_service_db' AND pid <> pg_backend_pid();
"
kubectl -n cosmic-void exec auth-service-db-0 -- \
  psql -U user -d postgres -c "DROP DATABASE cosmic_void_stats_service_db;"
kubectl -n cosmic-void exec auth-service-db-0 -- \
  psql -U user -d postgres -c "CREATE DATABASE cosmic_void_stats_service_db;"
# 重新裝 extension
kubectl -n cosmic-void exec auth-service-db-0 -- \
  psql -U user -d cosmic_void_stats_service_db -c "
    CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";
    CREATE EXTENSION IF NOT EXISTS pgcrypto;
"
kubectl -n cosmic-void rollout restart deployment stats-service
```

### 第三雷：game-service 用不同 env var 名

game-service main.go 用 `config.InitStatsServiceDB()` 而不是 `config.InitDB()`，這函式讀 `STATS_DB_USER / STATS_DB_PASSWORD / STATS_DB_HOST / STATS_DB_PORT / STATS_DB_NAME`，不是標準 `DB_*`。

```yaml
# game-service/k8s/configmap.yml
data:
  DB_HOST: "auth-service-db"
  ...
  # 額外加 STATS_DB_* 因為 game-service 連 stats 的 DB
  STATS_DB_HOST: "auth-service-db"
  STATS_DB_PORT: "5432"
  STATS_DB_NAME: "cosmic_void_stats_service_db"
  STATS_DB_USER: "user"
```

Secret 也加 STATS_DB_PASSWORD（值同 DB_PASS）。

## 知識點：kubectl exec websocket 噪訊

執行 kubectl exec 時偶爾看到：
```
E0515 03:22:31.043084   65743 websocket.go:296] Unknown stream id 1, discarding message
```

這是 kubectl 與 kubelet 透過 SPDY/WebSocket 溝通的雜訊，不影響 exec 結果。Ignore。

## 為什麼 manifest 結構這樣 vs 其他

| 替代 | 為何沒選 |
|---|---|
| Helm chart | 學習階段直接看 manifest 更直觀，Helm 是後期優化 |
| Kustomize | 同上，base + overlay 之後可加 |
| Per-service postgres StatefulSet | 7 個 StatefulSet × 1Gi PVC = 多花 disk + 維運成本，學習用單 instance 多 DB 即可 |
| Cloud SQL | $10/月（micro），90 天額外 $30，但 production-grade。可後期遷 |
| **單 Postgres + 多 DB**（採用）| 學習階段最便宜、夠用 |
