---
name: K8s 部署計劃 — auth-service + api-gateway
description: 本機 Docker Desktop K8s → 補齊依賴 → Ingress → AWS EKS 的四階段學習路線
status: in-progress
started: 2026-05-08
target: auth-service + api-gateway
extracted-to-vault: null
---

# K8s 部署學習計劃

> 起點：剛讀完 [[ingress]]，auth-service Dockerfile 已有，k8s manifests（deployment / service / configmap / secret.example）已有。
> 終點：auth-service + api-gateway 部署到 AWS EKS，有真實 HTTPS 域名。
> 風格：使用者自己寫，Claude 只 review。每一步完成後在這份文件打勾。

---

## 進度總覽

- [ ] **Phase 1** — 本機 Docker Desktop K8s 跑通 auth-service（含 in-cluster Postgres）
- [ ] **Phase 2** — 補全依賴（RabbitMQ / Redis / Consul 全部 in-cluster）
- [ ] **Phase 3** — 加上 api-gateway + Ingress（本機 self-signed TLS）
- [ ] **Phase 4** — 上 AWS EKS（ECR / RDS / AmazonMQ / ALB / cert-manager）

> 完成一個 phase 就把 `[ ]` 改成 `[x]` 並補上完成日期。

---

## Phase 1：本機 Docker Desktop K8s — auth-service 跑通

**目標**：`kubectl get pods` 看到 auth-service 2/2 Running，`curl localhost:8081/_status/healthz` 回 200。

### 1.1 啟用 Docker Desktop K8s
- [ ] Docker Desktop → Settings → Kubernetes → Enable Kubernetes
- [ ] 驗證：`kubectl config current-context` 應為 `docker-desktop`
- [ ] 驗證：`kubectl get nodes` 看到一個 Ready 的 node

### 1.2 Build auth-service image（本機）
- [ ] 切到 `game-server/` 目錄
- [ ] `docker build -f auth-service/Dockerfile -t cosmic-void/auth-service:dev .`
- [ ] 驗證：`docker images | grep auth-service` 看到 image
- [ ] 不需要 push registry — Docker Desktop 的 K8s 直接看本機 image

### 1.3 寫 in-cluster Postgres manifests
**新檔位置**：`game-server/auth-service/k8s/postgres.yml`（暫時放這裡，Phase 2 會搬去 infra/）

需要寫的物件：
- [ ] `Service`（headless，name = `auth-service-db`，對齊 ConfigMap 的 `DB_HOST`）
- [ ] `StatefulSet`（image: `postgres:17`，env 用 POSTGRES_USER/PASSWORD/DB）
- [ ] `PersistentVolumeClaim`（透過 StatefulSet 的 `volumeClaimTemplates`）

關鍵點（review 時會檢查）：
- Service 的 selector 要和 StatefulSet 的 labels 對齊
- PVC 不要用 hostPath（學 PV/PVC 的時機點到了）
- POSTGRES_PASSWORD 從 Secret 拿（不要寫在 yml 裡）

### 1.4 處理 Secret
- [ ] `cp secret.yml.example secret.yml`（已被 .gitignore 擋住，安全）
- [ ] 填入真實 `DB_PASSWORD`、`RABBITMQ_PASS`、`JWT_SECRET`
- [ ] 同時把 Postgres 的密碼也用同一個 Secret（DB_PASSWORD），避免兩邊不同步

### 1.5 修現有 deployment.yml 的兩個本機問題
- [ ] 加 `imagePullPolicy: IfNotPresent`（避免 K8s 跑去 registry 找 `:dev` tag）
- [ ] 把 `volumes.localtime` 的 hostPath 拔掉（Docker Desktop VM 內可能沒這檔），改用 env `TZ: Asia/Taipei`
  - 或者 Phase 1 階段先全部拔掉，Phase 4 再補
- [ ] readinessProbe 的 path `/_status/healthz` 確認程式真的有這個 route

### 1.6 Apply 順序與驗證
順序很重要，一次只 apply 一個並觀察結果：
- [x] `kubectl apply -f auth-service/k8s/configmap.yml` ✅ 2026-05-08
- [x] `kubectl apply -f auth-service/k8s/secret.yml` ✅ 2026-05-08
- [x] `kubectl apply -f auth-service/k8s/postgres.yml` → postgres Running ✅ 2026-05-08
  - 額外驗證過：PVC Bound 1Gi、pg_isready 通、database `cosmic_void_auth_service_db` 存在
- [ ] `kubectl apply -f auth-service/k8s/deployment.yml`
- [ ] `kubectl apply -f auth-service/k8s/service.yml`
- [ ] 驗證：`kubectl get pods -l component=auth-service` → 2/2 Running
- [ ] 驗證：`kubectl logs -l component=auth-service` 沒有連 DB 失敗的錯
- [ ] 驗證：`kubectl port-forward svc/auth-service 8081:8081` → `curl localhost:8081/_status/healthz` 回 200

### Phase 1 完成標記
- [ ] 全部勾完
- [ ] 在 `_index.md` 註記「Phase 1 完成於 YYYY-MM-DD」
- [ ] 把這次踩到的坑寫成一篇 learning note：`docs/learning/k8s/local-cluster-debugging.md`

---

## Phase 2：補全依賴（RabbitMQ / Redis / Consul）

**目標**：auth-service 啟動時連 DB / MQ / Consul 全部成功，沒有 retry log。

### 2.1 整理目錄結構
新建 `game-server/infra/k8s/`，把基礎設施放這裡：
```
game-server/
├── infra/k8s/
│   ├── postgres-auth.yml      ← 從 auth-service/k8s/ 搬過來
│   ├── rabbitmq.yml
│   ├── redis.yml
│   ├── consul.yml
│   └── README.md              ← apply 順序文件
├── auth-service/k8s/
│   ├── configmap.yml
│   ├── secret.yml (gitignored)
│   ├── deployment.yml
│   └── service.yml
```

- [ ] 建 `infra/k8s/` 目錄
- [ ] 把 `auth-service/k8s/postgres.yml` 搬成 `infra/k8s/postgres-auth.yml`

### 2.2 寫 RabbitMQ manifests
**檔案**：`infra/k8s/rabbitmq.yml`
- [ ] StatefulSet（image: `rabbitmq:3-management`）
- [ ] Service（name: `rabbitmq`，對齊 ConfigMap 的 `RABBITMQ_HOST`）
- [ ] 帳密用 Secret（共用 auth-service-secrets 或新建一個）

### 2.3 寫 Redis manifests
**檔案**：`infra/k8s/redis.yml`
- [ ] StatefulSet（image: `redis:7-alpine`）
- [ ] Service（name: `redis`）
- [ ] 啟動指令 `redis-server --appendonly yes`

### 2.4 寫 Consul manifests（dev 模式）
**檔案**：`infra/k8s/consul.yml`
- [ ] StatefulSet（image: `hashicorp/consul`，dev 單節點）
- [ ] Service（name: `consul`，對齊 ConfigMap 的 `CONSUL_ADDR=consul:8500`）
- [ ] 啟動指令 `agent -server -ui -bootstrap-expect=1 -client=0.0.0.0`

### 2.5 寫 apply 順序文件
**檔案**：`infra/k8s/README.md`
- [ ] 列出依賴順序：infra → service config → service deployment
- [ ] 寫 `make k8s-up` / `make k8s-down`（之後可以做，現在先手動）

### Phase 2 完成標記
- [ ] auth-service log 沒有任何「retry connect」訊息
- [ ] `kubectl exec -it auth-service-xxx -- nc -zv rabbitmq 5672` 通
- [ ] `kubectl exec -it auth-service-xxx -- nc -zv consul 8500` 通

---

## Phase 3：api-gateway + Ingress（本機）

**目標**：`curl -k https://api.cosmicvoid.local/some-route` 能打到 auth-service。

### 3.1 api-gateway 容器化
- [ ] 為 api-gateway 寫 Dockerfile（參考 auth-service 的多階段 build）
- [ ] build：`docker build -f api-gateway/Dockerfile -t cosmic-void/api-gateway:dev .`

### 3.2 寫 api-gateway K8s manifests
- [ ] `api-gateway/k8s/deployment.yml`（已有 service.yml + ingress.yml）
- [ ] `api-gateway/k8s/configmap.yml`

### 3.3 安裝 ingress-nginx controller
- [ ] `kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/cloud/deploy.yaml`
- [ ] 驗證：`kubectl get pods -n ingress-nginx` 看到 controller Running

### 3.4 改 ingress.yml 以適應本機
- [ ] 拔掉 `cert-manager.io/cluster-issuer` annotation（本機沒 cert-manager）
- [ ] host 從 `api.cosmicvoid.com` 改 `api.cosmicvoid.local`
- [ ] 手動建 self-signed cert：
  ```
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout tls.key -out tls.crt \
    -subj "/CN=api.cosmicvoid.local"
  kubectl create secret tls cosmic-void-tls --cert=tls.crt --key=tls.key
  ```

### 3.5 改 /etc/hosts
- [ ] `sudo vim /etc/hosts` 加一行 `127.0.0.1 api.cosmicvoid.local`
- [ ] 驗證：`curl -k https://api.cosmicvoid.local/<auth-service 的 route>` 回 200

### Phase 3 完成標記
- [ ] HTTPS 通了（self-signed，所以要 `-k`）
- [ ] 想想：本機 manifests 和雲端 manifests 差在哪？這就是 Phase 4 要用 kustomize 的理由

---

## Phase 4：上 AWS EKS

**目標**：同一份程式碼，跑在真 EKS 上，有真 HTTPS 域名。

> 開始這個 phase 之前，先確認 AWS 帳號有預算（小 cluster 一個月約 $70-100 美金）。

### 4.1 建 EKS cluster
- [ ] 安裝 `eksctl`、`aws cli`
- [ ] `aws configure`
- [ ] `eksctl create cluster --name cosmic-void --region ap-northeast-1 --node-type t3.small --nodes 2`
- [ ] 驗證：`kubectl get nodes` 看到 EKS 的 nodes

### 4.2 ECR — 把 image 推上去
- [ ] `aws ecr create-repository --repository-name cosmic-void/auth-service`
- [ ] `aws ecr create-repository --repository-name cosmic-void/api-gateway`
- [ ] 取得 login token：`aws ecr get-login-password | docker login --username AWS --password-stdin <account>.dkr.ecr.<region>.amazonaws.com`
- [ ] tag + push：`docker tag cosmic-void/auth-service:dev <account>.dkr.ecr...` → `docker push`

### 4.3 用 kustomize 做本機 / EKS overlay
新檔結構：
```
auth-service/k8s/
├── base/                  ← 共用（從現有 yml 搬過來）
├── overlays/
│   ├── local/             ← image: cosmic-void/auth-service:dev
│   └── eks/               ← image: <account>.dkr.ecr.../auth-service:dev
```
- [ ] 學 kustomize 基本用法
- [ ] 把現有 manifests 重組成 base + overlay
- [ ] `kubectl apply -k auth-service/k8s/overlays/eks/`

### 4.4 DB 改 RDS（不再用 in-cluster postgres）
- [ ] 在 AWS console 建 RDS Postgres（小 instance）
- [ ] 改 ConfigMap overlay：`DB_HOST: <rds-endpoint>`
- [ ] 驗證 auth-service 連得上 RDS

### 4.5 RabbitMQ 改 AmazonMQ
- [ ] 建 AmazonMQ broker
- [ ] 改 ConfigMap overlay：`RABBITMQ_HOST: <amazonmq-endpoint>`

### 4.6 真 Ingress + TLS
- [ ] 裝 AWS Load Balancer Controller（或繼續用 ingress-nginx）
- [ ] 申請或購買域名（Route53）
- [ ] 裝 cert-manager（這次有真 DNS，可以做 DNS-01 challenge）
- [ ] Ingress overlay 加回 `cert-manager.io/cluster-issuer: letsencrypt-prod`
- [ ] 驗證：`curl https://api.cosmicvoid.com/...` 真 HTTPS 通了

### Phase 4 完成標記
- [ ] EKS 上 auth-service Running
- [ ] 真 domain HTTPS 通
- [ ] 寫一篇總結 learning note：`docs/learning/k8s/eks-deployment.md`
- [ ] 不用了的時候記得 `eksctl delete cluster` 省錢

---

## 怎麼用這份文件

1. **每次開工**：打開這份文件，找到第一個沒打勾的步驟
2. **完成一步**：自己改檔案，把 `- [ ]` 改成 `- [x]`，需要 review 就跟我說「review Phase 1.3 我寫的 postgres.yml」
3. **遇到坑**：在那個步驟下面加註腳，例如：
   ```
   - [x] 1.5 readinessProbe path
     - 坑：原本 path 寫 /healthz，但程式實際是 /_status/healthz
   ```
4. **完成一個 phase**：在「進度總覽」打勾並寫完成日期，同步更新 `_index.md` 的「進行中的實作計劃」狀態
5. **全部完成**：把整份文件 frontmatter 的 `status` 改成 `completed`，準備萃取進 vault

---

## 關聯文件

- 起點：[[ingress]]（剛讀完）
- topic 總覽：[[_index]]
- 中期目標（PV / HPA / Helm）會在 Phase 1.3、Phase 4 自然碰到
