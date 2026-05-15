---
topic: gke-deployment
subtopic: troubleshooting
date: 2026-05-15
extracted-to-vault: ""
---

# Troubleshooting — 通用 debug 步驟集

按症狀分類，每個附 reproduce 指令。

## 症狀 1：Pod ImagePullBackOff / ErrImagePull

### Debug

```bash
kubectl -n cosmic-void describe pod <pod-name> 2>&1 | tail -20
# 看 Events section 找 Pulling / Failed
```

### 常見原因

| 訊息 | 原因 | 修法 |
|---|---|---|
| `403 Forbidden` | node SA 沒 artifactregistry.reader | grant role |
| `manifest unknown` | tag 錯（v1 vs v2） | 確認 `gcloud artifacts docker images list` 有 |
| `exec format error` | 沒 cross-compile | `--platform linux/amd64` |
| `not found` | image 從來沒 push | re-run buildx with `--push` |

## 症狀 2：Pod CrashLoopBackOff

### Debug

```bash
kubectl -n cosmic-void logs <pod-name> --tail=30
kubectl -n cosmic-void logs <pod-name> --previous --tail=30  # 看上次 crash 的 log

# 多 replica 一次看
kubectl -n cosmic-void logs -l component=<svc> --tail=20
```

### 常見原因

| log 樣態 | 原因 |
|---|---|
| `dial tcp <ip>:5432: connect: connection refused` | DB pod 沒起 / DB host 設錯 |
| `Failed to run migrations: no migration found for version N` | 共用 DB 撞 migration（看 06-db-per-service） |
| `Dirty database version N` | 上次 migration 跑到一半失敗（drop & recreate DB） |
| `pq: function uuid_generate_v4() does not exist` | DB 沒裝 uuid-ossp extension |
| `STRIPE_SECRET_KEY is required` | env var 沒設（payment-service） |
| `name resolver error: produced zero addresses` | Service DNS 解析失敗（OTEL collector 沒部署）|

## 症狀 3：Service Endpoint 沒有 Pod IP

```bash
kubectl -n cosmic-void get endpoints <svc-name>
# 如果是 <none>，Pod selector 沒 match 上
```

```bash
# 比對 Service selector vs Pod label
kubectl -n cosmic-void describe svc <svc> | grep Selector
kubectl -n cosmic-void get pods --show-labels | grep <expected-label>
```

## 症狀 4：LB Backend UNHEALTHY

```bash
kubectl -n cosmic-void describe ingress cosmic-void-ingress | grep -A 2 backends
# 看到 UNHEALTHY 後 ↓

# 抓 backend service 名稱
BSNAME=$(gcloud compute backend-services list --global --format="value(name)" --filter="name~<svc>")

# 看 health check 設定
HC_URL=$(gcloud compute backend-services describe "$BSNAME" --global --format="value(healthChecks[0])")
gcloud compute health-checks describe $(basename "$HC_URL") --global

# 看實際 health 結果
gcloud compute backend-services get-health "$BSNAME" --global
```

### 常見原因

| 現象 | 修法 |
|---|---|
| Pod /healthz 內部 OK，但 LB 仍 UNHEALTHY | BackendConfig 沒套上；annotate Service 觸發 re-sync |
| firewall 沒開 GCP HC IP 段（35.191/130.211）→ port | 看 `k8s-fw-l7--*` 有沒有自動建 |
| readinessProbe 路徑 404 | Service 沒這個路徑，加 `/healthz` 或設 BackendConfig 自訂 path |
| 第一次 deploy 慢 | 等 1-2 分鐘 NEG controller sync |

## 症狀 5：Ingress 一直沒 ADDRESS

```bash
kubectl -n cosmic-void get ingress
# ADDRESS 欄一直空
```

```bash
kubectl -n cosmic-void describe ingress cosmic-void-ingress | grep -i "error\|warning\|translate"
```

### 常見原因

| Translation event | 原因 |
|---|---|
| `service "X" is type "ClusterIP", expected NodePort or LoadBalancer when not using NEGs` | NEG annotation 缺 / 寫錯 / multi-port 部分缺 |
| `referenced secret X not found` | TLS secret 沒建 |
| `BackendConfig X not found` | BackendConfig CRD 沒 apply |

## 症狀 6：DNS 解析錯（從 cluster 外）

```bash
dig +short api.cosmicvoid.uk @8.8.8.8
# 應該回 34.111.74.79（或 Cloudflare proxy IP）
```

| 回 | 意義 |
|---|---|
| 34.111.74.79 | DNS-only 模式（Cloudflare 灰雲） |
| 104.21.x.x / 172.67.x.x | Cloudflare proxy 模式（橘雲）— 正常 |
| (empty) | DNS record 沒設 / 還沒 propagate |

## 症狀 7：DNS 在 cluster 內解析錯

```bash
# 從業務 pod 內試 service DNS
kubectl -n cosmic-void exec deploy/api-gateway -- nslookup auth-service.cosmic-void.svc.cluster.local
# 或更短形：
kubectl -n cosmic-void exec deploy/api-gateway -- nslookup auth-service
```

### Alpine 特殊行為

```bash
# Alpine 的 nslookup 行為怪，可能說 NXDOMAIN，但 nc 可以連
kubectl exec ... -- nc -z -w 3 <service> <port>  # nc OK = service 通
```

不要全信 nslookup，**改用 `nc -z` 確認 TCP 可達**。

## 症狀 8：GCE Ingress controller log

```bash
# 看 controller event
kubectl -n cosmic-void get events --sort-by=.lastTimestamp | tail -20

# 或集中看 Ingress 的 events
kubectl -n cosmic-void describe ingress cosmic-void-ingress | tail -30
```

## 症狀 9：Cloudflare → Origin 通不過

```bash
# 直接打 origin LB（繞 Cloudflare）
curl -H "Host: api.cosmicvoid.uk" http://34.111.74.79/healthz
# 期望 200。如果 200 → CF 那段壞了；如果非 200 → origin 壞了

# 經 Cloudflare 打
curl https://api.cosmicvoid.uk/healthz
```

### 兩端結果對比

| origin curl | CF curl | 結論 |
|---|---|---|
| 200 | 200 | 全部 OK |
| 200 | 5xx | CF 設定問題（SSL mode、page rule）|
| 非 200 | 5xx | LB / backend 壞 |
| empty | empty | LB rebuild 中 |

## 症狀 10：Docker build 失敗

```bash
# DON'T use tail in pipeline (吞 exit code)
docker buildx build ... 2>&1 | tail -8        # BAD

# DO save full log
docker buildx build ... > /tmp/build.log 2>&1
RESULT=$?
echo "Exit: $RESULT"
tail -30 /tmp/build.log
```

### Buildx 常見錯誤

| 訊息 | 原因 |
|---|---|
| `missing go.sum entry` | go.work 模式必須 copy 全部 9 個 go.mod |
| `no required module provides package` | `.dockerignore` 誤刪需要的目錄（看 03-docker-build） |
| `Type error: 'X' is declared but its value is never read` | TS strict + WIP 代碼，next.config 加 `ignoreBuildErrors: true` |
| `Cannot find module 'phaser'` | npm ci 沒跑或 cache invalidate 失敗 |

## 通用診斷工具箱

```bash
# Cluster 全貌
kubectl -n cosmic-void get pods,svc,ingress,endpoints

# 特定 service 完整資訊
kubectl -n cosmic-void get pod,svc -l component=<svc> -o wide

# 看誰 OOM 過
kubectl -n cosmic-void get events --field-selector reason=OOMKilling

# 看 cluster 整體資源
kubectl top nodes
kubectl top pods --all-namespaces --sort-by=cpu

# GCP 端 LB 全貌
gcloud compute forwarding-rules list
gcloud compute backend-services list --global
gcloud compute url-maps list

# Artifact Registry 狀態
gcloud artifacts docker images list us-central1-docker.pkg.dev/$PROJECT_ID/cosmic-void

# Firewall 規則
gcloud compute firewall-rules list --format="table(name,sourceRanges,allowed,targetTags)"
```

## 萬用 reset 招式

```bash
# 1. 刪 Ingress 強制 rebuild LB
kubectl -n cosmic-void delete ingress cosmic-void-ingress
sleep 5
kubectl apply -n cosmic-void -f api-gateway/k8s/ingress.yml

# 2. 強制 controller 重看 Service spec
kubectl -n cosmic-void annotate svc <svc> "refresh=$(date +%s)" --overwrite

# 3. 強制 Pod 重建（換 image / 改 secret 後）
kubectl -n cosmic-void rollout restart deployment <name>

# 4. Drop & recreate DB（migration dirty）
kubectl -n cosmic-void exec auth-service-db-0 -- psql -U user -d postgres -c "DROP DATABASE <db>;"
kubectl -n cosmic-void exec auth-service-db-0 -- psql -U user -d postgres -c "CREATE DATABASE <db>;"
```
