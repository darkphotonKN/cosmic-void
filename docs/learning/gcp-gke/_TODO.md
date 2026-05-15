---
date: 2026-05-15
topic: gcp-gke
subtopic: _TODO
extracted-to-vault:
---

# 未完成任務清單

## 對話結束時剩下的 2 個任務

### 1. Cloud Armor + Cloudflare IP allowlist（用戶已批准 +$6/月）

**狀態**：策略明確、用戶同意，**還沒執行**。

**做的步驟**：
```bash
# 抓 Cloudflare 最新 IPv4 CIDR 清單
CF_IPS=$(curl -s https://www.cloudflare.com/ips-v4)

# 建 security policy
gcloud compute security-policies create cosmic-void-cf-only \
  --description="Allow only Cloudflare edge IPs to reach the LB"

# 加 allow 規則（CF IPs）
CF_IPS_CSV=$(echo "$CF_IPS" | paste -sd, -)
gcloud compute security-policies rules create 1000 \
  --security-policy=cosmic-void-cf-only \
  --src-ip-ranges="$CF_IPS_CSV" \
  --action=allow \
  --description="Cloudflare edge IPs"

# default rule (priority 2147483647) 自動是 deny
# 但要把它從預設的 "allow" 改成 "deny":
gcloud compute security-policies rules update 2147483647 \
  --security-policy=cosmic-void-cf-only \
  --action=deny-403

# 掛到所有 backend services
for bs in $(gcloud compute backend-services list --global \
              --filter="name~k8s1-89c4df67" --format="value(name)"); do
  gcloud compute backend-services update "$bs" --global \
    --security-policy=cosmic-void-cf-only
done

# 驗證
curl -sS -m 10 -w "HTTP %{http_code}\n" -H "Host: api.cosmicvoid.uk" \
  http://34.111.74.79/healthz                                 # 應 403
curl -sS -m 10 -w "HTTP %{http_code}\n" https://api.cosmicvoid.uk/healthz  # 應 200
```

**注意**：
- 也要把 IPv6 加入（`curl -s https://www.cloudflare.com/ips-v6`）— Cloudflare 默認雙堆疊
- Cloudflare 更新 IP 段時要記得 sync 進 policy（一年一次左右）

### 2. Cluster max-nodes 從 4 降回 3（省 ~$24/月）

**狀態**：分析過、用戶沒明確指示，**還沒執行**。

**做的步驟**：
```bash
gcloud container clusters update cosmic-void --zone=us-central1-a \
  --enable-autoscaling --min-nodes=2 --max-nodes=3
```

**為什麼**：
- autoscaler 在 rolling restart 時短暫 surge 到 4 nodes
- 學習用 13 pod 在 3 nodes 充裕
- 4 nodes 持續存在 = +$24/月

**風險**：
- 緊密 rolling update 期間可能短暫 Pending
- 規避辦法：用 maxUnavailable=1 策略（api-gateway 已改）

## Phase 8 後續完善工作（用戶未批准的）

### 3. WebSocket 透過 Cloudflare 的 idle timeout（free plan 100 秒）

當前限制：CF Free plan WS idle timeout = 100s。遊戲循環頻繁送資料就無感，但「進 lobby 後閒置」會掉線。

修法選項：
- **client 側 heartbeat**（推薦）：每 30s 送個 ping message 保活
- **upgrade Cloudflare plan**：Pro $20/月，WS idle 提高到 5 分鐘
- **改用 WebSocket via raw GCP LB**（DNS-only 模式）：失去 Cloudflare 邊緣防護

### 4. Cloudflare → Origin 還是 HTTP（Flexible SSL）

當前：User HTTPS → CF → Origin **HTTP**（明文，但走 CF 私網）。

升級成 Full SSL：
1. 在 GKE 端用 cert-manager 簽自簽 cert 或申請 Origin Cert（Cloudflare 提供 15 年免費）
2. GCE Ingress 加 HTTPS port 443 + cert
3. Cloudflare SSL 模式從 Flexible 改 Full

**成本**：~30 分鐘工 + 0 額外金額。

### 5. Per-service secret 真正獨立（不是 default password）

當前：所有 service 共用同一個 DB_PASS / RABBIT_PASS / JWT_SECRET（為了部署簡單）。

正規做法：每個 service 自己一組密碼，via Secret Manager + Workload Identity。每次部署用 GitOps 同步。

### 5b. Secret 進 git 的稽查結果（2026-05-15）

**重大發現**：commit `c178a5b`（branch `feature/fixRetry_outbox`）曾經 commit `auth-service/k8s/secret.yml` 含：
```
DB_PASSWORD: "password"
RABBITMQ_PASS: "cosmicvoid"
JWT_SECRET: "dev-jwt-secret-change-in-prod"
```
都是 **dev placeholder**，不是 production 密碼（production 由 `openssl rand` 產生，存在 k8s Secrets + `~/.cosmic-void-secrets.env`）。

**決定**：private repo + dev 值 → 不清歷史。

**未來規則**：
1. 永遠用 `kubectl create secret --from-literal=...`，**不要落地** `secret.yml` 進 working tree
2. 真值用 `openssl rand -base64 24` 隨機產生，存 `~/.<project>-secrets.env`（mode 600，repo 外）
3. `secret.yml.example` 永遠 `REPLACE_ME` placeholder，可放 git
4. `.gitignore` 必須有 `**/k8s/secret.yml`（不擋 `.example`）

**已驗證的 `.gitignore` 完整度**：
- ✅ `**/k8s/secret.yml` 阻擋真值
- ✅ `*.pem` `*.key` `*.crt` `*.p12` `*.pfx`
- ✅ `.env` `.env.local` 等 dotenv 變體
- ✅ `secrets/` `.secrets/` `jwt-secret*`
- ✅ `.aws/` `ssl/` `certs/`

### 6. Postgres Cloud SQL 替換（production-grade）

當前：StatefulSet + 1Gi PVC 學習用。Node 重啟 PVC 保留資料，但備份/HA 全靠自己。

升級：Cloud SQL db-f1-micro $10/月，自動備份 + HA。Migration 步驟：
1. 建 Cloud SQL instance
2. `pg_dumpall` 從 cluster Postgres → 灌進 Cloud SQL
3. 改 configmap `DB_HOST` 指 Cloud SQL private IP（或 Cloud SQL Auth Proxy sidecar）
4. 砍 cluster Postgres StatefulSet

### 7. Observability 從「丟掉」升級成「真用」

當前：otel-collector debug exporter 把所有 traces/metrics 丟掉。

升級：exporter 改 otlphttp 指到 Grafana Cloud / Honeycomb / SigNoz 等 SaaS。多半有免費 tier 給開發者。

## 程式碼層級未完工項目（CONSUL_TO_K8S_MIGRATION.md Section 5）

### 5.3 跨 module 怪異 import

`common/constants/types/item.go` import `game-service/grpc/items` — 反向依賴方向，靠 go.work 繞過。長期該把 grpcitems client interface 抽到 common，或 item types 改成不依賴 grpcitems。

### 5.4 dial-per-call gRPC 模式

所有 gateway client / 內部 client 每次都新建 conn → close。改成長連線 + `dns:///` resolver + round_robin 客戶端 LB。等效能不夠再做。

## 程式碼層級已知問題（不卡部署但要清）

### TypeScript / ESLint 警告（next.config.ts ignoreBuildErrors）

`game-client/next.config.ts` 為了部署通先 disable 嚴格檢查。實際 unused vars / missing deps / `<img>` 警告該逐個修。當前壓在 TODO comment 裡。

清單：
- `src/app/portal/page.tsx` showHint state unused
- `src/app/subscription/page.tsx` router unused
- `src/app/leaderboard/page.tsx` missing useEffect dep `fetchLeaderboard`
- `src/app/profile/page.tsx` `<img>` 警告
- `src/components/NotificationBell.tsx` missing useEffect dep
- `src/components/UserMenu.tsx` `<img>` 警告
- Phaser default export warnings (4 處)
