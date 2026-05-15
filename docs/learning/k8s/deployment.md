---
title: "Kubernetes Deployment 入門"
type: learning-note
project: cosmic-void
topic: k8s
date: 2026-04-30
status: extracted
extracted-to-vault:
  - "[[Kubernetes Deployment]]"
  - "[[K8s RollingUpdate Strategy]]"
  - "[[K8s Labels and Selector]]"
  - "[[K8s Liveness vs Readiness Probe]]"
  - "[[K8s Resources Requests Limits]]"
  - "[[K8s Pod Volume]]"
extracted-at: 2026-04-30
archive-mirror: vault/創作庫/projects/cosmic-void/learning-archive/k8s/deployment.md
related-files:
  - game-server/deployment.yml
tags: [kubernetes, k8s, deployment, yaml, pod, container, rolling-update, probe, configmap, secret, volume]
---

## 學習目標

完整看懂 `game-server/deployment.yml` 每一個欄位，並理解每個欄位背後的設計意圖（為什麼存在、不寫會怎樣、寫錯會怎樣）。

## 對話脈絡

> Q: 我想學 K8s 但有點不懂 deployment.yml 要怎麼寫，可以一步一步教我？
> A: 從 YAML 結構本身開始（鍵值對 / 物件 / 清單三種積木），然後分階段：metadata → labels/selector → replicas/strategy → container 基礎四件套 → probes/env/volumes。

> Q: 我全部都不懂，我想理解 yml 的結構
> A: YAML 只有三種積木，全部都是這三種組合：
>   1. **鍵值對**：`name: kiki`
>   2. **物件**（巢狀）：用換行 + 縮排表示「裡面還有東西」
>   3. **清單**：用 `-` 開頭表示「一條一條的」
> 縮排決定父子關係，**不能用 Tab**，通常 2 空白。

> Q: name 該寫什麼？namespace 是什麼？kind 是什麼？
> A:
>   - `kind`：要建立什麼類型（Deployment / Service / ConfigMap...），決定後面 YAML 結構規則
>   - `metadata.name`：你給這個資源取的名字。**規則**：小寫字母+數字+連字號 `-`，不能大寫、不能底線（因為會被當 DNS 名稱用）
>   - `metadata.namespace`：屬於哪個「資料夾/部門」。預設 `default`。同 namespace 內 name 不能重複，跨 namespace 可以重複
>   - 我目前的 `name: COSMIC-VOID` 是錯的（大寫不合法）

> Q: 不太明白哪些是 node 部分、哪些是 pod 部分？
> A: **deployment.yml 裡完全沒有 Node 的東西，整份都是在描述 Pod**。
>   - **Node** = 機器（雲端業者給你的，你不會在 YAML 裡寫機器）
>   - **Pod** = 程式的執行單位（K8s 的 scheduler 會自動把 Pod 分配到某個 Node）
>   - 一個 Node 可以跑多個 Pod
>   - 你完全不用管 Pod 放哪個 Node，K8s 自動排程

> Q: 為什麼同一個 `app: COSMIC-VOID` 出現了三次？
> A: 三個位置角色完全不同：
>   1. `metadata.labels`：貼在 **Deployment 物件本身**的標籤（可省略，跟 Pod 無關）
>   2. `spec.selector.matchLabels`：Deployment **找 Pod** 的條件（必須跟 #3 一致）
>   3. `spec.template.metadata.labels`：**生出來的 Pod** 會貼的標籤（必須跟 #2 一致）
> 第 2 跟第 3 不一致時，K8s 直接擋掉 YAML（會無窮迴圈生 Pod）

> Q: 進階：selector 跟 template labels 是「完全一致」還是「子集匹配」？
> A: **子集匹配**。Pod 標籤多沒關係，只要包含 selector 要求的就匹配。但反過來不行（selector 多 Pod 少 → 不匹配）。
> 這個設計讓同一個 Pod 可以被多個 selector 從不同角度撈（不同 Service / 監控 / 藍綠部署）。

> Q: replicas 跟 strategy 怎麼運作？
> A:
>   - `replicas: N`：我要 N 個 Pod。生產環境**至少 2**，replicas: 1 換版本必中斷
>   - `strategy.type: RollingUpdate`：一邊開新的、一邊關舊的（預設、99% 用這個）
>   - `strategy.type: Recreate`：先全關再全開（會中斷，只在 Pod 不能並存時用）
>   - `maxSurge`：最多可以**超過** replicas 多少（暫時開更多新 Pod）
>   - `maxUnavailable`：最多可以**少於** replicas 多少（暫時砍掉幾個舊 Pod）

> Q: 換版本中可能的 Pod 數量？replicas: 10、maxSurge: 50%、maxUnavailable: 0
> A: 最多 15、最少 10。`maxUnavailable: 0` 意思是「絕不降載」，最安全但最耗資源（要先擴後縮）。

> Q: image / port / resources 怎麼用？
> A:
>   - `image: myapp:latest` 是雷 → 重啟時會偷偷拉到壞版本。生產環境用具體 tag（v1.2.3）或 commit hash
>   - `ports.containerPort` 是**文件不是設定**。容器實際監聽哪個 port 由你的程式決定（`ListenAndServe(":80")`）
>   - `resources.requests`：保證給你的最少資源（K8s 用來排程到 Node）
>   - `resources.limits`：你絕對不能超過的上限
>     - **CPU 超 limits → 被限速**（變慢，Pod 不死）
>     - **記憶體超 limits → OOMKilled**（Pod 被殺）
>   - 設 `requests < limits` 給尖峰彈性 + 平常省資源（QoS Class: Burstable）
>   - 設 `requests == limits` 最安全（QoS Class: Guaranteed），資源不夠時最後才被殺

> Q: livenessProbe 跟 readinessProbe 差在哪？
> A: 完全不同的兩件事，**搞混會出大事**：
>   - **liveness 失敗 → 殺 Pod 重生**（問「還活著嗎？」）
>   - **readiness 失敗 → 從 Service 移除（不導流量過去），Pod 不死**（問「準備好接流量嗎？」）
>
> 用餐廳比喻：
>   - liveness = 「廚師還有呼吸嗎？」沒呼吸 → 叫救護車
>   - readiness = 「廚師準備好接單嗎？」還沒好 → 點單先給別人

> Q: 兩個 probe 都用同一個 /healthz 會怎樣？
> A: **CrashLoopBackOff 災難**。如果 /healthz 會檢查 DB，DB 短暫斷線時：
>   1. liveness 失敗 → Pod 被殺
>   2. 新 Pod 啟動，DB 還沒恢復 → 又失敗
>   3. 又被殺 → 無限循環，整個服務崩潰
>
> 正確做法：
>   - liveness 用 `/healthz`：**只檢查程式本身**（HTTP server 還活著）
>   - readiness 用 `/readiness`：**檢查所有依賴**（包含 DB、cache）
> DB 斷線時：liveness ✅（Pod 不死）、readiness ❌（不接流量），DB 恢復後自動回來，**零停機**。

> Q: env 怎麼用？
> A: 三種寫法：
>   1. `env.value: "..."`：寫死值（密碼絕對不要這樣寫，會進 git 歷史）
>   2. `env.valueFrom.configMapKeyRef`：從 ConfigMap 讀（一般設定）
>   3. `env.valueFrom.secretKeyRef`：從 Secret 讀（密碼、token）
>
> 偷懶寫法：`envFrom` 整批匯入整個 ConfigMap / Secret。

> Q: volumes / volumeMounts 怎麼配對？
> A: 兩層宣告：
>   1. Pod 級宣告 volume（`spec.volumes`）：「我有一個叫 xxx 的儲存空間」
>   2. Container 級掛載（`containers[].volumeMounts`）：「我要把 xxx 掛到容器內的 /yyy 路徑」
> 分兩層的原因：同一個 volume 可被 Pod 內多個 container 共用（掛在不同路徑）。

> Q: emptyDir 在 Pod 重啟時會保留資料嗎？
> A: **會！** 但要分清楚：
>   - **container 重啟（OOMKilled、liveness 失敗）→ 同一個 Pod，emptyDir 保留**
>   - **Pod 重建（rolling update、Node 故障）→ 新 Pod，emptyDir 消失**
> 這是因為 emptyDir 綁在 Pod 上，不是 container 上。

## 關鍵理解

### 1. 三層俄羅斯娃娃

```
Deployment        ← 老闆（你寫的 YAML，管應用程式整體）
   └── ReplicaSet  ← 人資（K8s 自動建，管「保持 N 個副本」）
          └── Pod    ← 員工（隨時會死、會重生）
                 └── Container  ← 員工的工具（你的 Docker image）
```

寫 deployment.yml 時，其實是**同時定義三層**。`spec.template:` 以下整段就是 Pod 的設計圖。

### 2. K8s 是宣告式（declarative）

你寫「**我想要什麼狀態**」（要 N 個 Pod 在跑），K8s 持續監控、自動修復。
不是「**怎麼做**」（你不用寫「先啟動 Pod、再⋯⋯」）。

### 3. Pod 是 ephemeral（一次性的）

Pod 隨時會死、會被重建。重建後：
- name 變了
- IP 變了
- 記憶體狀態歸零

→ 所以你的 app 必須是 **stateless**，資料要存到外部 DB 或持久化 volume。

### 4. 三個 labels 必須對齊

```
spec.selector.matchLabels   ⇄  spec.template.metadata.labels
（找什麼條件的 Pod）          （生出來的 Pod 貼什麼）
                必須一致
```

不一致 → K8s 直接擋掉 YAML（防止無窮生 Pod）。

### 5. 兩個 probe 嚴格分工

| Probe | 檢查什麼 | 失敗後果 | endpoint 應該檢查什麼 |
|---|---|---|---|
| liveness | 程式還活著嗎？ | 殺 Pod | **只看自己**（HTTP server / process） |
| readiness | 準備好接流量嗎？ | 從 Service 移除 | **看所有依賴**（DB / cache / external API） |

絕對不能用同一個 endpoint，否則 DB 一掛全部 Pod 進入 CrashLoopBackOff。

## 程式碼 / 設定

### 我目前的 game-server/deployment.yml（有問題的版本）

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name:  COSMIC-VOID            # ❌ 大寫不合法
  namespace: default
  labels:
    app:  COSMIC-VOID           # ❌ 建議小寫
spec:
  selector:
    matchLabels:
      app: COSMIC-VOID
  replicas: 1                    # ⚠️ 生產環境至少 2
  strategy:
    rollingUpdate:
      maxSurge: 25%              # 配 replicas: 1 沒意義（25% × 1 = 0）
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        app:  COSMIC-VOID
    spec:
      containers:
      - name:  COSMIC-VOID
        image:  MYAPP:latest     # ❌ :latest 是雷
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
          limits:
            cpu: 100m            # requests==limits 是 Guaranteed QoS（沒彈性）
            memory: 100Mi
        livenessProbe:
          tcpSocket:
            port: 80             # ⚠️ tcpSocket 太寬鬆，建議 httpGet
          initialDelaySeconds: 5
          # ...
        readinessProbe:
          httpGet:
            path: /_status/healthz
            port: 80
          # ...
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: COSMIC-VOID
              key: DB_HOST
        ports:
        - containerPort:  80
          name:  COSMIC-VOID
        volumeMounts:
        - name: localtime
          mountPath: /etc/localtime
      volumes:
        - name: localtime
          hostPath:
            path: /usr/share/zoneinfo/Asia/Shanghai  # ⚠️ 應該改成 Asia/Taipei
      restartPolicy: Always       # Deployment 只能 Always，這行可省
```

### 修正後的生產等級版本（給未來的我參考）

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cosmic-void              # ✅ 小寫
  namespace: default
  labels:
    app: cosmic-void
spec:
  replicas: 3                    # ✅ 至少 2，3 個更穩
  selector:
    matchLabels:
      app: cosmic-void
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1                # ✅ replicas 小用絕對數字
      maxUnavailable: 0          # ✅ 0 表示換版本不降載
  template:
    metadata:
      labels:
        app: cosmic-void
    spec:
      containers:
      - name: cosmic-void
        image: myregistry/cosmic-void:v1.2.3   # ✅ 不要用 :latest
        resources:
          requests:
            cpu: 100m
            memory: 256Mi        # ✅ 平常需要的
          limits:
            cpu: 500m            # ✅ 給尖峰彈性（Burstable QoS）
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /healthz       # ✅ 只檢查程式本身
            port: 80
          initialDelaySeconds: 10
          periodSeconds: 10
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /readiness     # ✅ 檢查所有依賴
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
        envFrom:                 # ✅ 整批匯入比一個一個寫清爽
        - configMapRef:
            name: cosmic-void-config
        - secretRef:
            name: cosmic-void-secrets   # ✅ 密碼放 Secret
        ports:
        - containerPort: 80
          name: http
        volumeMounts:
        - name: localtime
          mountPath: /etc/localtime
          readOnly: true         # ✅ 系統檔案唯讀更安全
      volumes:
      - name: localtime
        hostPath:
          path: /usr/share/zoneinfo/Asia/Taipei  # ✅ 台灣時間
```

## 踩過的坑

- 問題：把 selector 跟 template labels 寫成不同的值會怎樣？
  解法：K8s 會擋掉 YAML（apply 時報錯 `selector does not match template labels`）。
  為什麼：避免無窮迴圈生 Pod。Deployment 找不到自己生的 Pod 會一直再生。

- 問題：兩個 probe 都用 `/healthz` 且 endpoint 檢查 DB
  解法：拆兩個 endpoint。liveness `/healthz`（只看程式本身）、readiness `/readiness`（檢查 DB 等依賴）。
  為什麼：DB 短暫斷線時，liveness 失敗會殺 Pod，但新 Pod 啟動 DB 還沒恢復，又被殺，進入 CrashLoopBackOff。

- 問題：`replicas: 1` 配 `maxUnavailable: 25%` 沒意義
  解法：replicas 小（< 4）用絕對數字（`maxSurge: 1, maxUnavailable: 0`）。
  為什麼：25% × 1 = 0.25，K8s 向下取整為 0，實際上等於 unavailable=0。

- 問題：把 DB 密碼直接寫在 `env.value`
  解法：用 `secretKeyRef` 從 Secret 讀。
  為什麼：寫進 YAML = 永久進 git 歷史；CI log 會印出；任何能看 Deployment 的人都看得到（RBAC 權限 Deployment 通常比 Secret 寬鬆）。

- 問題：用 `image: myapp:latest` 部署兩週後突然壞掉
  解法：用具體版本 tag 或 commit hash。
  為什麼：YAML 沒變但 registry 的 latest 指標會變，Pod 重啟（OOMKilled、Node 維護、scale up）時會拉到新版本。

- 問題：`time.Now()` 印出來時間不對
  解法：volumes hostPath 從 `Asia/Shanghai` 改成 `Asia/Taipei`。
  為什麼：容器內 `/etc/localtime` 軟連到 Node 上的時區檔案。

## 待釐清

- [ ] 我猜應該寫 auth-service 自己的 deployment.yml，DB（5103）、RabbitMQ、Consul、gRPC 7003 這些怎麼接？要查每個依賴是 Service 還是直接 hostname？
- [ ] `initContainers` 我目前只看了概念，實作上：等 DB ready / 跑 migration 的 init container 應該寫 shell 還是用專門的 image？
- [ ] `requests < limits` 的具體比例怎麼抓？1.5x、2x、5x？（先抓 2x 試試）
- [ ] readiness 要怎麼實作？我的 Go service 要新增 `/readiness` endpoint，然後 ping DB / RabbitMQ？

## 補充：docker-compose vs Dockerfile vs Image vs K8s 的關係

問題起因：盤點 auth-service 部署需求時發現它沒 Dockerfile，但 `game-server/docker-compose.yml` 卻存在 → 一度以為 docker-compose 是 image。

### 概念釐清

| 概念 | 是什麼 | 比喻 |
|---|---|---|
| Dockerfile | 「怎麼蓋房子」的施工說明書 | 食譜 |
| Docker image | 蓋好的房子模板（壓縮檔）| 食譜做出來的菜（裝盒打包）|
| Docker container | 用模板實際蓋出的房子（運行中）| 端上桌正在吃的菜 |
| docker-compose.yml | 「一次跑多棟房子」的編排檔 | 整套餐廳出餐 SOP |

**docker-compose.yml 不是 image**，它是「**用 image 啟動多個 container**」的編排檔。

### cosmic-void 目前的開發流程觀察

`game-server/docker-compose.yml` 裡每個 service 的 `image:` 欄位**全部是現成的 image**（postgres、rabbitmq、redis、consul），**沒有任何一個是自己寫的 Go 服務**。

→ 開發流程其實是：
```
docker-compose up -d         # 啟動依賴（DB / MQ / Consul / Redis）
go run cmd/server/main.go    # Go 服務「在本機」直接跑（不在 container 裡）
本機 Go → 連 container 內的 DB（透過 localhost:5103）
```

→ 這就是為什麼 auth-service 沒 Dockerfile：**目前沒在 container 化自己寫的服務**。要部署到 K8s，每個 Go 服務都要先寫 Dockerfile + build 成 image。

### docker-compose vs K8s

| | docker-compose | K8s |
|---|---|---|
| 跑在哪 | 一台機器 | 多台機器組成的 cluster |
| 用途 | 開發環境 | 生產環境 |
| 自動修復 | ❌ | ✅ |
| 擴縮容 | ❌ | ✅ HPA |
| Rolling update | ❌ | ✅ 內建 |

**正確 mental model**：兩者不衝突，本機開發用 compose、正式部署用 K8s，大公司都是這樣。

### docker-compose 概念對照 K8s

| docker-compose | K8s |
|---|---|
| `services:` 下的一個 service | Deployment |
| `image:` | `containers[].image` |
| `ports:` | `containers[].ports` + Service |
| `environment:` | `env:` + ConfigMap/Secret |
| `volumes:` | `volumes:` + PVC |
| `networks:` | Service + NetworkPolicy |
| `depends_on:` | initContainers / readinessProbe |
| `restart: always` | `restartPolicy: Always` |
| `healthcheck:` | `livenessProbe` + `readinessProbe`（K8s 拆兩種）|

### 對 cosmic-void 的部署啟示

要把 cosmic-void 搬上 K8s，需要做的事：
1. 每個 Go 服務寫 Dockerfile（auth / notification / game / ...）
2. build image 並推到 registry
3. K8s Deployment 用這些 image
4. 依賴服務（DB / RabbitMQ / Redis）也轉成 K8s YAML（之後學 Service 後做）

## 相關專案檔案

- `game-server/deployment.yml` ← 這次學習的範本
- `game-server/auth-service/.env` ← 之後寫 auth-service deployment.yml 的設定來源
- `game-server/auth-service/cmd/server/main.go` ← 看 auth-service 啟動了什麼（gRPC、outbox worker）

## 相關 learning notes

- （目前還沒有其他 K8s 筆記，未來會有 service.md / configmap-secret.md / ingress.md）
