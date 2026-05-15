---
date: 2026-05-15
topic: gcp-gke
subtopic: 07-debugging
extracted-to-vault:
---

# 07 — 部署過程所有 incident 的因果鏈

按發生順序，每個 incident 列「症狀 → 根因 → 修法 → 教訓」。

## 1. 8 個 Go service `GOWORK=off` build 全 fail

**症狀**：本地 + Docker build 都報 `missing go.sum entry for go.mod`

**根因**：用戶開發都用 go.work workspace 模式，每個單一 module 的 go.sum 從來沒被 `go mod tidy` 過 → 缺 entry。Docker build 設 `GOWORK=off` 用單模組模式就炸。

**修法**：Dockerfile 保留 go.work，copy 全部 9 個 module 的 go.mod/go.sum 進 builder。

**教訓**：workspace 模式優點是開發省事，但每個 module 的 go.sum 仍應該完整自包含才好部署。下次該定期 `cd <svc> && GOWORK=off go mod tidy` 一次保持 go.sum 同步。

## 2. `.dockerignore` 排掉了業務 code

**症狀**：build 報 `no required module provides package github.com/.../common/discovery/k8s`

**根因**：`.dockerignore` 寫 `**/k8s/`，遞迴匹配所有名為 k8s 的目錄，包括 `common/discovery/k8s/`（Go service discovery 的 package）。

**修法**：改成顯式列每個 service 的 k8s/ 目錄。

**教訓**：monorepo 裡 `**/<name>` 通配很危險。盡量用 `*/<name>` 或 explicit path list。

## 3. Background bash 報 exit 0 但 build 其實 fail

**症狀**：notification 說「completed (exit code 0)」，registry 卻只有 7 個 image，少 game-service。

**根因**：我寫的 wrapper 是：
```bash
if docker buildx build ... | tail -8; then
  echo OK
else
  echo FAIL
fi
```
`| tail -8` 的退出碼是 tail 的（永遠 0），不是 docker 的。失敗看不出來。

**修法**：改成把全 log 存檔，最後再 tail 顯示：
```bash
docker buildx build ... > /tmp/log 2>&1
RESULT=$?
tail -8 /tmp/log
exit $RESULT
```

**教訓**：bash pipeline exit code 預設取最後一個 command 的。要看全管線 fail 用 `set -o pipefail` 或別 pipe。

## 4. GKE node 拉 image 全 ImagePullBackOff

**症狀**：
```
failed to pull and unpack image "...": failed to authorize:
failed to fetch oauth token: ... 403 Forbidden
```

**根因**：新版 GCP project 的 default Compute SA `<NUM>-compute@developer.gserviceaccount.com` 不再自動配 `roles/editor`。GKE node 用這個 SA 拉 Artifact Registry，沒權限。

**修法**：
```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$DEFAULT_SA" \
  --role="roles/artifactregistry.reader" --condition=None
kubectl -n cosmic-void rollout restart deployment
```

**教訓**：新版 GCP 預設 SA 是「最小權限」起點，要明確賦 role。舊文章常假設 default SA 有 editor，不可全信。

## 5. 服務一起來 panic：DB migration「no migration found for v15」

**症狀**：example-service / stats-service / 等 CrashLoopBackOff，log 顯示 schema_migrations 表已有 v15 但找不到對應 file。

**根因**：起初所有 service 都連到同一個 DB `cosmic_void_auth_service_db`。每個 service 都跑自己的 migrations，把對方的 schema_migrations table 弄亂。

**修法**：每個 service 自己一個 DB（`cosmic_void_<svc>_db`），patch 各 configmap 的 DB_NAME。

**教訓**：multi-tenant DB 跑 migration 工具會撞車。要嘛各自 DB，要嘛各自 schema，最差也得用不同的 `schema_migrations` table 名。

## 6. stats-service「Dirty database version 1」

**症狀**：stats-service log 反覆 `Could not run migrations: Dirty database version 1. Fix and force version.`

**根因**：v1 migration 用 `uuid_generate_v4()` 需要 `uuid-ossp` extension。Postgres 17 預設沒裝。第一次 migration 部分 table 建好就 fail → `schema_migrations` 留下 `(version=1, dirty=t)`。下次起服務看到 dirty 直接放棄。

**修法**：
1. 在所有 service DB 裝 extension：
   ```sql
   CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
   CREATE EXTENSION IF NOT EXISTS pgcrypto;
   ```
2. DROP + CREATE stats DB 清掉 dirty state。

**教訓**：Postgres 13+ 內建 `gen_random_uuid()`，不需要 extension。`uuid_generate_v4()` 是 uuid-ossp 的，能用就用 `gen_random_uuid()` 省事。

## 7. game-service env var 名與其他 service 不同

**症狀**：DSN `postgres://:JL0W7r...@:/?sslmode=disable` user/host/db 全空。

**根因**：game-service main.go 用 `config.InitStatsServiceDB()`，這個 func 讀 `STATS_DB_*` 而不是 `DB_*`。

**修法**：configmap 加 `STATS_DB_HOST` / `STATS_DB_PORT` / `STATS_DB_NAME` / `STATS_DB_USER`，Secret 加 `STATS_DB_PASSWORD`。

**教訓**：跨 service 的 env naming convention 該統一。`config/db.go` 寫不同名字 = 隱形坑。

## 8. OTEL collector log spam 每 10 秒一次

**症狀**：每個 service log 噴：
```
failed to upload metrics: exporter export timeout: rpc error: code = Unavailable
desc = name resolver error: produced zero addresses
```

**根因**：configmap 設 `COLLECTOR_ENDPOINT: "otel-collector:4317"`，但沒部署這個 Service。CoreDNS 解不到 → SDK 重試 every 10s。

**修法**：部署極簡 OTEL Collector（receivers: otlp + exporters: debug verbosity=basic）→ DNS 通了 + 收到 metrics 就丟掉。

**教訓**：configmap 寫了某個依賴的位置就要負責部署它，不然 SDK 會無止盡重試。

## 9. OTEL spam 部署 collector 後**仍未停止**

**症狀**：collector 部署完 + DNS 確認可解析，service log 還是繼續噴錯。

**根因**：Go gRPC client 有 internal DNS cache + connection state。Service 在 collector 上線前已啟動，client 進入 transient failure 狀態，後續 retry 也不重新 DNS resolve。

**修法**：`kubectl rollout restart deployment` 重啟所有 service → 新 pod 從乾淨狀態 DNS resolve 成功。

**教訓**：DNS 變更後 gRPC client 不會自動恢復。任何 gRPC dial 失敗的 service 該 restart。

## 10. ManagedCertificate「FailedNotVisible」

**症狀**：
```
Status: Provisioning
Domain: api.cosmicvoid.uk
Status: FailedNotVisible
```

**根因**：用戶 DNS 在 Cloudflare 且預設**橘雲 proxied**，dig 回 Cloudflare IP（104.21.x.x）而不是 GCP LB IP（34.111.74.79）。Google 的 ACME challenge probe 經過 Cloudflare 失敗。

**修法**：完全跳過 ManagedCert，改用 Cloudflare proxy + Flexible SSL（見 [05-ingress-tls.md](05-ingress-tls.md)）。

**教訓**：DNS 走 proxy（任何 CDN）+ Google-managed cert 不相容。要嘛 DNS-only 走 Google，要嘛全給 Cloudflare 處理 TLS。

## 11. Backend UNHEALTHY 但 /healthz from pod OK

**症狀**：`kubectl exec deploy/api-gateway -- wget -qO- localhost:7001/healthz` 回 200 OK，但 `kubectl describe ingress` 顯示 backend UNHEALTHY。

**根因**：GCE LB 的 BackendService 早期用 default health check（GET / 期望 200），後加的 readinessProbe `/healthz` 不會回頭去更新 BackendService 設定。

**修法**：建 BackendConfig 明確指定 `requestPath: /healthz`，Service 加 annotation 引用：
```yaml
cloud.google.com/backend-config: '{"default": "api-gateway-backend-config"}'
```

**教訓**：GCE Ingress 的 BackendService 建好後設定**不會自動更新**。要嘛刪 ingress 重建，要嘛用 BackendConfig 明確設定（推薦）。

## 12. Rollout 死鎖：NEG readiness gate vs LB

**症狀**：新 api-gateway pod Pending：
```
0/4 nodes are available: 1 node(s) were unschedulable, 3 Insufficient cpu
no new claims to deallocate
LoadBalancerNegNotReady: Waiting for pod to become healthy in at least one of the NEG(s)
NotTriggerScaleUp: Pod didn't trigger scale-up: 1 max node group size reached
```

**根因**：deployment strategy `maxSurge: 1, maxUnavailable: 0` 要求新 pod Ready 才砍舊。新 pod 帶 NEG readiness gate，gate 要 backend HEALTHY 才滿足。但 backend 還是舊 pod 服務（沒 `/healthz`）→ UNHEALTHY → 新 pod 永遠不 Ready。

**修法**：改 strategy `maxSurge: 0, maxUnavailable: 1`，舊的先砍。

**教訓**：NEG mode + LB health check 跟 deployment rolling strategy 互動微妙。資源緊時更明顯。

## 13. GCE Ingress `Translation failed`（multi-port Service）

**症狀**：
```
Translation failed: invalid ingress spec: service "cosmic-void/game-service"
is type "ClusterIP", expected "NodePort" or "LoadBalancer" when not using NEGs
```

**根因**：game-service 有兩個 port（grpc 7004 + http 5555）。NEG annotation `{"ingress": true}` 只給 Ingress 引用的 port 建 NEG，7004 沒 NEG。Controller 檢查所有 ports，看到 7004 沒 NEG 罷工。

**修法**：拆 Service — `game-service`（只 7004）+ `game-service-ws`（只 5555，有 NEG）。

**教訓**：GCE Ingress 對 multi-port Service 不友善，能拆就拆。

## 14. WebSocket bundle 沒注入 NEXT_PUBLIC_WS_URL

**症狀**：rebuild + redeploy 後 bundle 仍 `ws://localhost:5555`，不是 `wss://cosmicvoid.uk`。

**根因**：Dockerfile 沒宣告 `ARG NEXT_PUBLIC_WS_URL`，build 環境變數沒被 Next.js inline。

**修法**：Dockerfile 加 `ARG NEXT_PUBLIC_WS_URL=wss://cosmicvoid.uk` + `ENV NEXT_PUBLIC_WS_URL=$NEXT_PUBLIC_WS_URL`。

**教訓**：Dockerfile ARG 必須先宣告才能用 `--build-arg` 注入。CLI 不報錯但靜默忽略。

## 通用 debug checklist

```bash
# 服務 log
kubectl -n cosmic-void logs -l component=<svc> --tail=30
kubectl -n cosmic-void logs -l component=<svc> --previous --tail=20

# 服務內部測試
kubectl -n cosmic-void exec deploy/<svc> -- wget -qO- http://localhost:<port>/health

# 跨 service test
POD_IP=$(kubectl -n cosmic-void get pod -l component=<src> -o jsonpath='{.items[0].status.podIP}')
kubectl -n cosmic-void run probe --rm -i --restart=Never --image=alpine -- \
  sh -c "wget -qO- http://$POD_IP:<port>/path"

# Endpoints（Service→Pod 對應）
kubectl -n cosmic-void get endpoints

# LB backend
kubectl -n cosmic-void describe ingress cosmic-void-ingress | grep -A 1 backends
kubectl -n cosmic-void describe ingress cosmic-void-ingress | grep -E "Translate|Sync|Warning"

# GCP LB
gcloud compute backend-services list --global
gcloud compute backend-services get-health <BS_NAME> --global

# Firewall
gcloud compute firewall-rules list

# DNS 解析
dig +short <host> @8.8.8.8
dig +short <host> @1.1.1.1
```
