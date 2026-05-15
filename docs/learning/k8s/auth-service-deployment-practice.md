---
title: "第一次寫 K8s YAML 實戰：auth-service deployment + configmap + secret"
type: learning-note
project: cosmic-void
topic: k8s
date: 2026-04-30
status: extracted
extracted-to-vault:
  - "[[K8s ConfigMap]]"
  - "[[K8s Secret]]"
  - "[[K8s envFrom vs env]]"
  - "[[Docker Image vs Container vs Compose]]"
  - "[[K8s Service Discovery]]"
  - "[[Gitignore Patterns]]"
  - "[[K8s Labels and Selector]] (補充: 多重 labels 模式)"
  - "[[K8s Liveness vs Readiness Probe]] (補充: tcpSocket vs httpGet 實戰)"
extracted-at: 2026-04-30
archive-mirror: vault/創作庫/projects/cosmic-void/learning-archive/k8s/auth-service-deployment-practice.md
related-files:
  - game-server/auth-service/k8s/deployment.yml
  - game-server/auth-service/k8s/configmap.yml
  - game-server/auth-service/k8s/secret.yml.example
  - game-server/auth-service/.env
  - game-server/auth-service/cmd/server/main.go
  - game-server/docker-compose.yml
  - .gitignore
  - docs/learning/k8s/example-deployment.yml
tags: [kubernetes, k8s, deployment, configmap, secret, envfrom, gitignore, docker-compose, auth-service, hands-on]
---

## 學習目標

把 `docs/learning/k8s/deployment.md` 學的概念實際應用，為 cosmic-void 的 auth-service 寫出第一份**完整可被 kubectl 接受**的 K8s 部署設定（deployment.yml + configmap.yml + secret.yml.example），並建立保護真實 secret 不進 git 的機制。

## 對話脈絡

> Q: auth-service 寫一份完整 deployment.yml 應該放在哪？我現在是放在 game-server
> A: 業界三種主流做法：
>   - A：每個微服務自己的 `k8s/`（最常見、推薦）
>   - B：集中在 `game-server/k8s/`（容易不同步）
>   - C：頂層 `deploy/` 或 `infrastructure/`（大型專案）
> 對 cosmic-void 選 A —— 跟 service 程式碼綁一起，每個 service 有自己的 `Dockerfile + migrations + k8s/`。

> Q: docker-compose.yml 算 image 嗎？
> A: **不是**。釐清四個概念：
>   - Dockerfile = 食譜（怎麼蓋房子）
>   - Image = 蓋好的房子模板（壓縮檔）
>   - Container = 用模板實際蓋出的房子（運行中）
>   - docker-compose.yml = 一次跑多棟房子的編排檔
>
> 重要觀察：cosmic-void 的 docker-compose.yml 裡 `image:` 全是別人家的（postgres、rabbitmq、redis、consul），**沒有任何自己寫的 Go 服務**。所以 auth-service 沒 Dockerfile —— 因為目前根本沒在 container 化。要部署到 K8s 要先寫 Dockerfile。

> Q: envFrom 實際上該怎麼寫？是創新檔嗎？image 放在哪？
> A:
>   - `envFrom` 引用的 ConfigMap / Secret **是獨立的 K8s 資源**，要在另外的 YAML 檔定義
>   - 三個檔案的順序：先 `kubectl apply configmap.yml + secret.yml`，後 `kubectl apply deployment.yml`
>   - `image` 已經放在對的位置（`containers[]` 底下，跟 `name` / `resources` / `ports` 同層）

> Q: 為什麼 secret.yml.example 不直接叫 secret.yml？
> A: 密碼不能進 git，但要有 template 讓協作者知道格式。所以建 `.example` 檔（含 REPLACE_ME 占位），真實 `secret.yml` 透過 .gitignore 排除。

> Q: 不小心把密碼 push 到 git，現在改 ConfigMap 把值刪掉夠嗎？
> A: **不夠**。git history 會永久保留，必須 `git filter-branch` 或 `BFG` 從整個歷史移除。**重點**：與其修不如預防 —— 一開始就用 `.example` 模式 + .gitignore。

## 關鍵理解

### 1. K8s YAML 檔案之間的「契約關係」

寫 K8s 不是寫一份 YAML，而是寫**多份互相引用**的 YAML：

```
deployment.yml
   │  envFrom: configMapRef: name: auth-service-config
   │  envFrom: secretRef: name: auth-service-secrets
   ↓  （依名字引用）
configmap.yml (name: auth-service-config)
secret.yml    (name: auth-service-secrets)
```

引用是**靠名字字串**，沒有編譯期檢查。打錯字就引用不到（apply 時 K8s 才會抱怨「ConfigMap not found」）。

### 2. envFrom 用 `Ref`、env 用 `KeyRef`（容易混淆）

| 用法 | 寫法 | 拿什麼 |
|---|---|---|
| `env` 拿單一個 | `configMapKeyRef: { name, key }` | 從某 ConfigMap 拿某個 key |
| `env` 拿單一個 secret | `secretKeyRef: { name, key }` | 從某 Secret 拿某個 key |
| `envFrom` 整批匯入 | `configMapRef: { name }` | 整個 ConfigMap 全拿 |
| `envFrom` 整批匯入 | `secretRef: { name }` | 整個 Secret 全拿 |

**記法**：
- 拿單一 key → 結尾 `KeyRef`（要寫 key）
- 整批拿 → 結尾 `Ref`（不寫 key）

### 3. K8s 內部的 hostname/port 跟本機不同

| 服務 | 本機（docker-compose）| K8s 內部 |
|---|---|---|
| Postgres | `localhost:5103` | `auth-service-db:5432` |
| RabbitMQ | `localhost:5682` | `rabbitmq:5672` |
| Consul | `localhost:8510` | `consul:8500` |

**為什麼變了**：
- K8s Service 提供 cluster 內部 DNS，hostname 就是 Service 的 name
- K8s 內部 port 不需避免衝突（每個 Service 是獨立 namespace），用標準 port
- docker-compose 對外映射的 port（5103 / 5682 / 8510）是為了跟本機其他東西不撞，K8s 不需要

### 4. Secret 的 stringData vs data

```yaml
# 給人類用：stringData（直接寫明文，K8s 自動 base64）
stringData:
  DB_PASSWORD: "mypassword"

# K8s 內部儲存格式：data（必須先手動 base64）
data:
  DB_PASSWORD: bXlwYXNzd29yZA==
```

**關鍵**：base64 **不是加密**，只是編碼（可逆）。真正的安全靠 K8s RBAC 限制誰能讀 Secret。

### 5. ConfigMap 的 value 必須是字串（要加引號）

```yaml
data:
  DB_PORT: "5432"        # ✅ 字串
  DB_PORT: 5432          # ❌ 數字 → K8s 報錯
  SERVICE_VERSION: 1.0   # ❌ 會被解析成數字 1.0
```

**規則**：所有 value 都加引號，最安全。YAML 會「靠值的形狀猜類型」，`yes` / `no` / `1.0` 都可能被誤判。

### 6. 多重 labels（app + component）的設計價值

```yaml
labels:
  app: cosmic-void           # 「我屬於哪個系統」
  component: auth-service    # 「我是這個系統的哪個元件」
```

好處：
- `kubectl get pods -l app=cosmic-void` 一次撈整個系統的所有 Pod
- `kubectl get pods -l component=auth-service` 只撈 auth-service 的 Pod
- 兩個維度的篩選都行

### 7. .gitignore pattern 的位置敏感性

```gitignore
# 在 repo 根目錄的 .gitignore：

*.md                  # ⚠️ 全 repo 所有 .md 都被忽略（過廣！）
docs/learning/**/*.md # ✅ 精準只擋 docs/learning/ 下的 .md
/docs/                # ⚠️ 整個 docs 資料夾都被忽略
.skills/              # ✅ 整個 .skills 資料夾被忽略
```

**踩過的雷**：原本 `.gitignore` 寫 `/docs/` + `*.md`，導致連學習筆記都進不了 git。後來才意識到「我要忽略的是 `docs/learning/`，不是整個 `/docs/`」。

## 程式碼 / 設定

### deployment.yml（最終版）

```yaml
# game-server/auth-service/k8s/deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: default
  labels:
    app: cosmic-void
    component: auth-service
spec:
  selector:
    matchLabels:
      app: cosmic-void
      component: auth-service
  replicas: 2
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: cosmic-void
        component: auth-service
    spec:
      containers:
      - name: auth-service
        image: cosmic-void/auth-service:dev
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          tcpSocket:
            port: 7003                # 戳 gRPC port（暫時用 TCP，待補 HTTP healthz）
          initialDelaySeconds: 5
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 3
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /_status/healthz
            port: 8081                # auth-service 的 HTTP port
          initialDelaySeconds: 5
          timeoutSeconds: 2
          successThreshold: 1
          failureThreshold: 3
          periodSeconds: 10
        envFrom:
        - configMapRef:               # ✅ 整批匯入用 Ref（沒 Key）
            name: auth-service-config
        - secretRef:
            name: auth-service-secrets
        ports:
        - containerPort: 7003
          name: grpc
        - containerPort: 8081
          name: http
        volumeMounts:
        - name: localtime
          mountPath: /etc/localtime
          readOnly: true
      volumes:
        - name: localtime
          hostPath:
            path: /usr/share/zoneinfo/Asia/Taipei
      restartPolicy: Always
```

### configmap.yml（最終版）

```yaml
# game-server/auth-service/k8s/configmap.yml
kind: ConfigMap
apiVersion: v1                  # 注意：ConfigMap 是 v1，不是 apps/v1
metadata:
  name: auth-service-config
  namespace: default
  labels:
    app: cosmic-void
    component: auth-service
data:
  SERVICE_VERSION: "1.0.0"
  ENVIRONMENT: "PRODUCTION"
  GRPC_AUTH_ADDR: "7003"
  DB_HOST: "auth-service-db"      # K8s Service 名稱（不是 localhost）
  DB_PORT: "5432"                 # K8s 內部用標準 port
  DB_NAME: "cosmic_void_auth_service_db"
  DB_USER: "user"
  RABBITMQ_HOST: "rabbitmq"
  RABBITMQ_PORT: "5672"
  RABBITMQ_USER: "cosmicvoid"
  CONSUL_ADDR: "consul:8500"
  COLLECTOR_ENDPOINT: "otel-collector:4317"
```

### secret.yml.example（範本，真實 secret.yml 不進 git）

```yaml
# game-server/auth-service/k8s/secret.yml.example
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secrets
  namespace: default
  labels:
    app: cosmic-void
    component: auth-service
type: Opaque
stringData:                       # 給人類寫：直接明文
  DB_PASSWORD: "REPLACE_ME"
  RABBITMQ_PASS: "REPLACE_ME"
  JWT_SECRET: "REPLACE_ME"
```

### .gitignore 修正後的關鍵段落

```gitignore
# Skills 定義（個人配置不進 git）
.skills/

# Learning notes（個人學習筆記不進 git，靠 vault backup）
docs/learning/**/*.md

# K8s real secrets (only .example versions go to git)
**/k8s/secret.yml
**/k8s/secret.yaml
**/k8s/*-secret.yml
**/k8s/*-secret.yaml
```

## 踩過的坑

- 問題：`metadata.name` 第一版寫 `cosmic-void`
  解法：改成 `auth-service`
  為什麼：name 是「這個資源的識別」，要能精確指出「哪個服務」。`cosmic-void` 是專案名，不是服務名。同 namespace 下其他服務也叫 `cosmic-void` 會撞名。

- 問題：`metadata.labels` 跟 `selector.matchLabels` 寫不同的值
  解法：三個地方（metadata.labels / selector.matchLabels / template.metadata.labels）統一
  為什麼：雖然 metadata.labels 跟 selector 沒有強制關聯，但讀起來很混亂。統一比較好維護。後來改成 `app + component` 雙重標籤，提升可篩選性。

- 問題：`envFrom` 第一次寫成 `configMapKeyRef`
  解法：改成 `configMapRef`（沒 Key）
  為什麼：`KeyRef` 是用來在 `env` 拿單一個 key，整批匯入要用 `Ref`。記法：「整批拿不需要 key」。

- 問題：`envFrom` 第二項寫 `- name: auth-service-secrets` 直接接 name
  解法：改成 `- secretRef: { name: ... }`
  為什麼：每個 envFrom item 必須先指定**來源類型**（configMapRef 或 secretRef），再寫 name。

- 問題：`image: MYAPP:latest` 沒改（從範本 copy 忘了改）
  解法：改成 `cosmic-void/auth-service:dev`
  為什麼：範本只是占位，每個服務要用自己的 image 名。`:latest` 也是雷（之前 deployment.md 有講）。

- 問題：`probe.port: 80` 跟 `containerPort: 80`，但 auth-service 沒在 80
  解法：改成 7003（gRPC）跟 8081（HTTP）
  為什麼：要查程式實際監聽哪些 port（看 .env 的 `*_ADDR` + main.go 的 `Listen`）。auth-service 開 7003 (gRPC) + 8081 (HTTP)。

- 問題：configmap 把 RabbitMQ port 寫 5682、Consul 寫 8510
  解法：改成標準 port（5672 / 8500）
  為什麼：5682 / 8510 是 docker-compose 為了避免本機 port 衝突而映射的「對外」port。K8s 內部用 Service DNS，每個 Service 是獨立 port namespace，不會撞，所以用服務原生的標準 port。

- 問題：configmap `SERVICE_VERSION: 1.0.0` 沒加引號
  解法：加引號 `"1.0.0"`
  為什麼：ConfigMap 規定 value 必須是字串。`1.0.0` 因為兩個點剛好被解析成字串（運氣好），但 `1.0` 會變數字 1.0，會報錯。**所有 ConfigMap value 都加引號最安全**。

- 問題：secret.yml.example 檔名末尾有兩個空格
  解法：mv 改名移除空格
  為什麼：副檔名後面有空白會讓很多工具無法處理（kubectl apply、Read 工具）。檔名永遠不要有尾隨空白。

- 問題：secret 的 `app: cosmicvoid`（缺連字號）
  解法：改成 `cosmic-void`
  為什麼：跟 deployment.yml / configmap.yml 不一致。三個檔案 label 必須統一，否則 `kubectl get all -l app=cosmic-void` 會漏掉這個 Secret。

- 問題：`.gitignore` 寫 `/docs/` 跟 `*.md`
  解法：刪掉這兩條，改成精準的 `docs/learning/**/*.md`
  為什麼：`/docs/` 把整個 docs 資料夾忽略；`*.md` 把全 repo 所有 .md 忽略。導致連 README、學習筆記、`game-server/k8s-game-microservices-guide.md` 都不進 git。**規則要精準，越廣的規則越容易誤傷**。

- 問題：補的 K8s secret pattern 行尾有大量尾隨空白
  解法：用 python 把所有行 rstrip
  為什麼：尾隨空白雖然不影響 .gitignore 行為，但是不乾淨的編輯痕跡，linter 會警告。

- 問題：寫 deployment.yml 時不知道 ConfigMap / Secret 的 hostname 從哪來（K8s 裡 `auth-service-db`、`rabbitmq` 這些是什麼？）
  解法：暫時當作「之後會建的 K8s Service 名字」，先寫進 ConfigMap。學完 Service 章節再回來實際建這些 Service。
  為什麼：K8s 的 Service 提供 cluster 內部 DNS，hostname 就是 Service 的 metadata.name。這就是為什麼 Service 是 Deployment 的下一步。

## 待釐清

- [ ] 補 auth-service 的 Dockerfile（目前 image: cosmic-void/auth-service:dev 還沒對應的真 image）
- [ ] 改 main.go 的 `net.Listen("tcp", "localhost:"+grpcAddr)` —— K8s 裡 localhost 只看自己，要改成 `:7003` 聽所有介面
- [ ] 補 HTTP /healthz endpoint（讓 livenessProbe 從 tcpSocket 升級成 httpGet）
- [ ] 補 /readiness endpoint（包含 DB / RabbitMQ / Consul 連線檢查）
- [ ] auth-service main.go 裡 8081 port 的用途確認（pprof？metrics？）
- [ ] 學完 Service 後回來實際建 auth-service-db / rabbitmq / consul Service 物件
- [ ] 學 ConfigMap / Secret 章節時，研究「值改了 Pod 要不要重啟才會生效」（envFrom 是不會自動 reload 的）

## 相關專案檔案

- `game-server/auth-service/k8s/deployment.yml` ← 這次寫的
- `game-server/auth-service/k8s/configmap.yml` ← 這次寫的
- `game-server/auth-service/k8s/secret.yml.example` ← 這次寫的
- `game-server/auth-service/.env` ← 設定來源（盤點時參考）
- `game-server/auth-service/cmd/server/main.go` ← 確認程式 listen 哪些 port
- `game-server/docker-compose.yml` ← 確認依賴服務的標準 port（vs 對外映射 port）
- `.gitignore` ← 修正規則保護 secret 和 learning notes
- `docs/learning/k8s/example-deployment.yml` ← 上一輪建立的乾淨教學範本

## 相關 learning notes

- [[deployment]] — 概念入門（這次的前置學習）
- [[../../README]] —（未來）TBD：Service 學完後新增 [[service]] 連結這份
