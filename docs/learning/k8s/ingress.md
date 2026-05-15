---
title: "K8s Ingress 入門 + 實戰寫 cosmic-void HTTPS 路由"
type: learning-note
project: cosmic-void
topic: k8s
date: 2026-05-08
status: learning
extracted-to-vault: []
related-files:
  - game-server/api-gateway/k8s/ingress.yml
  - game-server/api-gateway/k8s/service.yml
  - game-server/api-gateway/.env
  - game-server/api-gateway/cmd/main.go
tags: [kubernetes, k8s, ingress, https, tls, cert-manager, lets-encrypt, nginx-ingress, routing, hands-on]
---

## 學習目標

理解 Ingress 解決的問題（多 service 對外時的 LoadBalancer 浪費）、Ingress vs Ingress Controller 的差異、HTTP 路由規則（host / path / pathType）、HTTPS / TLS termination + cert-manager 自動發證、實戰寫 cosmic-void 的 ingress.yml + api-gateway service.yml。

## 對話脈絡

> Q: cosmic-void 之後想用什麼當對外入口？
> A: api-gateway（cosmic-void 已有這個 service），未來部署 AWS。Ingress Controller 選 NGINX（業界最通用、不綁定特定雲端）。

> Q: Ingress 跟其他 K8s 資源最大的差異？
> A: **Ingress 本身只是「路由規則的描述」，不會自己跑**。需要先裝 Ingress Controller（一個跑在 cluster 內的 Pod，內部其實是 nginx），它持續監聽 Ingress 資源並根據規則設定自己的 nginx。
>
> 比喻：
>   - Ingress（你的 YAML）= 送貨地址表 📋
>   - Ingress Controller = 真的有人在送貨 🚚
>
> 所以 Ingress 是「兩階段設定」：先裝 Controller（一次性），再寫 Ingress YAML（每個 app）。

> Q: 沒裝 Controller 寫了 Ingress 並 apply，會發生什麼？
> A: apply **會成功**（K8s 接受 YAML），但**沒人執行規則**。kubectl get ingress 看 ADDRESS 欄位是空的（沒 controller 接 = ADDRESS 空白）。

> Q: cosmic-void 的 api-gateway Service 該用什麼 type？
> A: **ClusterIP**。Ingress 後面的 Service 一律用 ClusterIP，否則繞過 Ingress 變成兩個對外入口。架構：「**整個 cluster 只有 Ingress Controller 一個 LoadBalancer，其他全部 ClusterIP**」。

> Q: Postgres TCP 能用 Ingress 暴露給外部嗎？
> A: **不行**。Ingress 是 L7（HTTP/HTTPS）路由器，只看 HTTP Host header / path。Postgres 是純 TCP，沒這些東西，Ingress 看不懂。
>   - L4（TCP/UDP）→ LoadBalancer / NodePort
>   - L7（HTTP/HTTPS）→ Ingress
>
> 但 DB 不該對外，需要遠端連用 `kubectl port-forward svc/postgres 5432:5432`。

> Q: pathType 三種差在哪？
> A:
>   - **Prefix**（最常用）：路徑層級前綴，`/api` 匹配 `/api`、`/api/v1`，但**不匹配 `/apiv2`**
>   - **Exact**：完全相等，`/healthz` 不匹配 `/healthz/`
>   - **ImplementationSpecific**：依 controller 而定（nginx 支援 regex），不可移植

> Q: 有兩條 path 規則 `/api` 跟 `/api/v2`，請求 `/api/v2/users` 會被導去哪？
> A: **只有 /api/v2 一條**（這題我答錯了）。Ingress 路由是排他的，**「路徑越長越優先」**。讓你能寫「特殊規則 + 通用 fallback」。

> Q: `/auth` (Prefix) 會匹配 `/authentication` 嗎？
> A: **不會**（這題我答錯了）。Prefix 不是字串前綴，是**「資料夾前綴」**：
>   - ✅ `/auth` 匹配 `/auth`、`/auth/login`、`/auth/v1/me`
>   - ❌ `/auth` 不匹配 `/authentication`（因為 `auth` ≠ `authentication`）
>
> 設計理由：避免新加的 `/authentication` 路徑誤被 `/auth` 規則抓走。

> Q: 「TLS termination」是什麼意思？
> A: **加密在 Ingress「終止」**（這題我答錯，誤把「發證」當成 termination）：
>   - 外部 → Ingress：HTTPS（加密）
>   - Ingress → Pod：HTTP（明文）
>
> 後端服務（Go 程式）完全不用懂 TLS。證書集中在 Ingress 一處管理，更新證書只動 Ingress 不影響業務 Pod。
> 「邊界加密、內部明文」是 99% 場景的正確設計。

> Q: 三種拿 TLS 證書的方法？
> A:
>   - **A 自簽**（dev / 學習）：openssl 自己簽，瀏覽器警告，免費
>   - **B 商業 CA**（傳統）：DigiCert 等買，不警告但花錢，手動更新
>   - **C cert-manager + Let's Encrypt**（業界標準）⭐：免費、受信任、自動續期

> Q: cert-manager 自動發證流程？
> A:
>   1. 你寫 ingress.yml + annotation `cert-manager.io/cluster-issuer: letsencrypt-prod`
>   2. cert-manager 看到 → 跟 Let's Encrypt 申請
>   3. Let's Encrypt 要驗證網域擁有權 → cert-manager 在 Ingress 加臨時規則 `/.well-known/acme-challenge/<token>`
>   4. Let's Encrypt 戳這個 URL 驗證通過 → 簽發證書
>   5. cert-manager 自動建 Secret（type: kubernetes.io/tls）
>   6. Ingress 自動載入 → HTTPS 啟用 ✅
>
> 證書到期前 30 天自動續期，全程你不用做事。

> Q: HTTPS Ingress 但沒設 cert-manager / 沒手動建 TLS Secret，瀏覽器打 https 會怎樣？
> A: apply **會成功**（K8s 不檢查 Secret 存在），但 nginx-ingress 用 **fake/default certificate** 回應，瀏覽器顯示「NET::ERR_CERT_AUTHORITY_INVALID」警告。HTTP 反而正常運作（沒加密但能連）。

> Q: Let's Encrypt staging vs prod？
> A: 上 prod 之前**先用 staging 測試**：
>   - staging 證書不受信任（測試用）但 rate limit 寬鬆
>   - prod 真實證書但每週只能 50 個（反覆測會被擋一週）
>
> 業界鐵律：先 staging 測通，再切 prod。

## 關鍵理解

### 1. Ingress 跟其他資源不同：要先裝 Controller

```
寫 deployment / service / configmap → kubectl apply → K8s 自動跑 ✅
寫 ingress → kubectl apply → 沒人接 ❌（除非裝了 Ingress Controller）
```

Ingress 是「描述路由規則」，Controller 才是「真的執行」。

### 2. 業界標準架構：1 LB + 1 Ingress + N ClusterIP Services

```
Internet → 1 個 LoadBalancer (給 Ingress Controller)
              ↓
         Ingress Controller (Pod，內部跑 nginx)
              ↓ 根據 ingress.yml 路由
         N 個 ClusterIP Services
              ↓
         業務 Pods
```

省錢（1 個 LB 而非 N 個）+ 統一管理 HTTPS / 路由 / annotation。

### 3. Ingress 是 L7（HTTP/HTTPS）路由器

- 看 HTTP Host header → host-based routing
- 看 HTTP path → path-based routing
- 看 HTTP method / headers → 進階 annotation

**純 TCP/UDP 不能用 Ingress**（Postgres、Redis、自訂 TCP 服務）→ 用 NodePort / LoadBalancer。

### 4. pathType: Prefix 是「資料夾前綴」不是「字串前綴」

```
✅ /auth 匹配 /auth, /auth/login, /auth/v1/me
❌ /auth 不匹配 /authentication, /authority

按 / 分隔的路徑層級匹配，不是純字串比對
```

設計理由：避免「字串相似但語義不同」的路徑誤入規則。

### 5. 路由排他 + 最長路徑優先

```
請求只會被一條規則處理（不分裂）
   ↓
規則越具體（路徑越長）越優先
   ↓
讓「特殊處理 + 通用 fallback」成為可能
```

範例：
```yaml
- path: /api/v2/admin    # 最具體 → admin handler
- path: /api/v2          # 中等 → v2 handler
- path: /api             # 通用 → legacy handler
```

### 6. rules 按 host 分組

```
3 個不同 host = 3 條 rules
同 host 不同 path = 1 條 rule + 多個 paths
```

### 7. TLS termination：邊界加密、內部明文

```
Internet ──HTTPS──→ Ingress ──HTTP──→ Pod
                       ↑
                  TLS 在這裡終止
                  證書集中管理
```

後端服務完全不用懂 TLS。「邊界加密、內部明文」是 99% 場景的正確設計。要更安全（mTLS）才需要 service mesh（Istio / Linkerd）。

### 8. cert-manager + Let's Encrypt = 自動發證 + 自動續期

```
你寫：
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  spec:
    tls:
    - hosts: [...]
      secretName: ...

cert-manager 自動：
  1. 跟 Let's Encrypt 申請
  2. 完成 ACME HTTP01 驗證
  3. 建 Secret (type: kubernetes.io/tls)
  4. 證書到期前 30 天自動續期
```

業界鐵律：**先用 letsencrypt-staging 測通，再切 letsencrypt-prod**（避免被 50/週 rate limit 擋）。

### 9. SSL redirect 強制 HTTPS

```yaml
annotations:
  nginx.ingress.kubernetes.io/ssl-redirect: "true"
  nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
```

效果：HTTP 請求自動 301 redirect 到 HTTPS。⚠️ value 必須加引號（YAML 會把 true 解析成布林）。

### 10. K8s 物件靠「name 字串」對齊（沒型別檢查）

```
ingress.backend.service.name: api-gateway    ┐
                                              ↓ 必須對齊
service.metadata.name:        api-gateway     ✅

service.spec.ports[].name:    http            ┐
                                              ↓ 必須對齊
ingress.backend.service.port.name: http       ✅

service.spec.selector.component: api-gateway  ┐
                                              ↓ 找 Pod
deployment.template.metadata.labels.component: api-gateway ✅
```

打錯一個字符就壞掉，apply 會成功但 runtime 會 connection refused。

### 11. apply 不檢查 runtime（再次驗證）

整個 Ingress 寫了 + cert-manager annotation + tls 區塊：
- apply ✅ 成功
- 但 Secret 不存在 → 用 fake cert
- backend service name 對不上 → connection refused

**「apply 成功 ≠ 服務能通」永遠都是真理**。寫完要實際 curl 驗證。

## 程式碼 / 設定

### ingress.yml 最終版（game-server/api-gateway/k8s/ingress.yml）

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cosmic-void-ingress
  namespace: default
  labels:
    app: cosmic-void
    component: api-gateway
  annotations:
    # cert-manager 自動發 Let's Encrypt 證書（先用 staging 測試流程）
    cert-manager.io/cluster-issuer: letsencrypt-staging
    # 強制 HTTPS：HTTP 請求自動 301 redirect 到 HTTPS
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.cosmicvoid.com
    secretName: cosmic-void-tls       # cert-manager 會自動建立這個 Secret
  rules:
  - host: api.cosmicvoid.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-gateway
            port:
              name: http              # 用 name 引用 Service 的 ports.name
```

### service.yml 最終版（game-server/api-gateway/k8s/service.yml）

```yaml
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
  namespace: default
  labels:
    app: cosmic-void
    component: api-gateway
spec:
  type: ClusterIP
  selector:
    app: cosmic-void
    component: api-gateway       # 對應未來 api-gateway Deployment 的 Pod labels
  ports:
  - name: http
    port: 80                     # Ingress 連這個 port（HTTP 標準）
    targetPort: http             # 用 name 引用 Pod 的 containerPort.name
    protocol: TCP
```

### 整套契約對齊

```
ingress.yml:
  backend.service.name: api-gateway     ─┐
  backend.service.port.name: http       ─┤
                                         ↓
service.yml:
  metadata.name: api-gateway             ✅
  spec.ports[].name: http                ✅
  spec.selector.component: api-gateway  ─┐
                                         ↓ 等未來 Deployment 對齊
  (deployment.yml 之後寫)
    spec.template.metadata.labels:
      component: api-gateway
    spec.template.spec.containers[].ports:
    - containerPort: 7001
      name: http
```

## 踩過的坑

- 問題：Q1 把「TLS termination」答成「發放證書」
  解法：理解兩個是不同階段的事
  為什麼：cert-manager（事前發證）跟 Ingress Controller（拿證書解密 HTTPS）是兩個 component。「termination」就是字面意思「終止」—— 加密連線在 Ingress 終止，內部變成明文。

- 問題：Q1（路徑匹配）以為「兩條 path 都會被路由」
  解法：理解 Ingress 路由是排他的，最長優先
  為什麼：直覺以為「匹配多條 = 都會處理」，但 Ingress 是 routing decision，一個請求只會被一條規則處理。讓「特殊規則 + 通用 fallback」成為可能。

- 問題：Q2 以為 `/auth` 會匹配 `/authentication`
  解法：Prefix 是路徑層級前綴，不是字串前綴
  為什麼：直覺把 Prefix 當字串比對，但實際上是按 `/` 分隔的層級匹配。`/auth` = 「資料夾 /auth/」匹配，不會誤抓 `/authentication`。設計理由：避免新加路徑誤入舊規則。

- 問題：第一版 ingress.yml 整體有 2 格多餘縮排
  解法：頂層 apiVersion / kind / metadata / spec 從第 1 格開始
  為什麼：YAML 雖然能解析，但業界慣例頂層不縮排。git diff / merge / 工具相容性更好。

- 問題：第一版 ingress.yml 完全沒設 HTTPS（沒 cert-manager / 沒 tls / 沒 SSL redirect）
  解法：補三組 annotation + tls 區塊
  為什麼：學習目標是「HTTPS 強制 + cert-manager」但寫的是純 HTTP 路由。學完知識要對應到實作。

- 問題：第一版 ingress.yml 用 `port.number: 80`（寫死數字）
  解法：改成 `port.name: http`（用 name 引用）
  為什麼：未來 api-gateway port 改變只動 Service / Deployment，不用動 Ingress。前提：Service 的 ports[] 必須有對應的 name: http。

- 問題：第一版 service.yml 從 auth-service 複製貼上沒改完（name: auth-service / component: api-service / 多了 grpc port / port 數字錯）
  解法：每個欄位都要重新審視，不能盲目複製
  為什麼：複製貼上模板很方便但容易留下「身份混亂」的物件。Service name = `auth-service` 卻屬於 api-gateway 是大雷（會跟原本 auth-service 撞名，後 apply 蓋前 apply）。

- 問題：第二版 service.yml 改了 component 但 metadata.name 還是 `api-service`
  解法：「一個邏輯服務 = 一個名字，到處都用同一個」
  為什麼：Service name = api-service 但 Ingress 引用的是 api-gateway，對不上。整套生態系應該用同一個名字串聯（labels / selector / metadata.name 全部 `api-gateway`），讓 `kubectl get all -l component=api-gateway` 能一次撈整套。

- 問題：service 的 ports 多寫了 grpc（api-gateway 沒 gRPC）
  解法：盤點程式實際開的 port 才寫
  為什麼：api-gateway 的 .env 只有 PORT=7001（HTTP，gin router）。從 auth-service 複製過來時把 grpc 帶進來。**要寫 K8s YAML 之前必須先盤點服務實際的 port 設計**（可參考 auth-service-deployment-practice.md 的盤點清單）。

- 問題：Service 的 port 該寫 80 還是 7001 困惑
  解法：理解 port 三層概念
  為什麼：
    - Service.port: 對 Ingress 的契約（HTTP 標準 80）
    - Service.targetPort: 對 Pod 的 port name 引用
    - Pod.containerPort: 程式實際監聽的 7001
  Service.port 跟 Pod port 解耦是設計優點 —— 對外用標準 port 80，內部用任何 port 都行。

## 待釐清

- [ ] 安裝 nginx-ingress 到本機 minikube 跑通流程
- [ ] 安裝 cert-manager 跑 Let's Encrypt staging 流程（DNS 怎麼設？本機沒法驗證網域擁有權，可能要用 dns01 而非 http01）
- [ ] api-gateway 的 deployment.yml 還沒寫，整套還跑不起來
- [ ] 補 api-gateway 的 Dockerfile（同 auth-service 的問題）
- [ ] HTTP /healthz endpoint：api-gateway 還沒有，未來要補（讓 readinessProbe 用 httpGet 而非 tcpSocket）
- [ ] 多 host 的 Ingress 怎麼設計？例如 admin.cosmicvoid.com 也加進來
- [ ] gRPC 走 Ingress 行不行？auth-service 的 7003 gRPC 要怎麼對外？（nginx-ingress 支援 gRPC 但要特殊 annotation）
- [ ] 沒 healthz 用 cert-manager 的 ACME HTTP01 驗證，會不會有衝突？（路徑都是 / 開頭）
- [ ] 上 prod 之前的 checklist：cert-manager / nginx-ingress 安裝、DNS 設定、staging 測試、prod 切換

## 相關專案檔案

- `game-server/api-gateway/k8s/ingress.yml` ← 這次寫的 HTTPS Ingress
- `game-server/api-gateway/k8s/service.yml` ← 這次寫的 ClusterIP Service（給 Ingress 引用）
- `game-server/api-gateway/.env` ← 確認 PORT=7001
- `game-server/api-gateway/cmd/main.go` ← 確認程式 listen 7001
- `game-server/auth-service/k8s/service.yml` ← 之前寫的，命名慣例參考來源

## 相關 learning notes

- [[deployment]] — Pod / Deployment 概念（Ingress 後面接 Service 接 Pod）
- [[service]] — Service 提供 hostname 給 Ingress 引用
- [[configmap-and-secret]] — TLS 證書是 type: kubernetes.io/tls 的 Secret
- [[auth-service-deployment-practice]] — K8s 物件命名慣例 + 盤點服務需求的方法
- [[kubectl-cheatsheet]] — debug Ingress / 查 Endpoints 等指令
