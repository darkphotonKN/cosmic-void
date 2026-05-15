---
title: "Kubernetes Service 入門 + 實戰寫 auth-service.yml"
type: learning-note
project: cosmic-void
topic: k8s
date: 2026-04-30
status: extracted
extracted-to-vault:
  - "[[Kubernetes Service]]"
  - "[[K8s Service Port Layers]]"
  - "[[K8s Service Discovery]] (從 stub 升級)"
  - "[[K8s Apply vs Runtime]]"
  - "[[K8s Labels and Selector]] (補充: Service selector 風險)"
extracted-at: 2026-04-30
archive-mirror: vault/創作庫/projects/cosmic-void/learning-archive/k8s/service.md
related-files:
  - game-server/auth-service/k8s/service.yml
  - game-server/auth-service/k8s/deployment.yml
  - game-server/auth-service/k8s/configmap.yml
tags: [kubernetes, k8s, service, clusterip, nodeport, loadbalancer, externalname, dns, fqdn, coredns, endpoints, hands-on]
---

## 學習目標

理解 Service 解決的問題（Pod IP 不穩定）、四種 type 的差異、port 三層概念、K8s 內部 DNS 運作機制，並為 cosmic-void 的 auth-service 寫出第一份完整 Service YAML。

## 對話脈絡

> Q: configmap.yml 裡 DB_HOST: "auth-service-db" 這些名字，K8s 怎麼讓你寫名字就找到對的 Pod？是什麼機制？
> A: K8s 內建 CoreDNS。每次建立 Service，CoreDNS 自動加一筆 DNS 紀錄。Service name 就是 hostname。完整 FQDN：`<service>.<namespace>.svc.cluster.local`。

> Q: 同 namespace 裡有 3 個 auth-service Pod，外部呼叫 auth-service:7003 一次會打到哪個 Pod？
> A: kube-proxy + iptables（或 IPVS）做負載平衡，預設隨機/round-robin。**只導給 readinessProbe ✅ 的 Pod**。

> Q: 我有 Deployment 跑 3 個 Pod 標籤是 `app: cosmic-void, component: auth-service`，Service selector 只寫 `app: cosmic-void`，會找到 3 個 Pod 嗎？
> A: **會找到**（子集匹配規則）。但這很危險 —— 如果 cluster 裡還有其他標籤包含 `app: cosmic-void` 的 Pod（例如 notification-service），這個 Service 會把流量導到全部，造成服務錯亂。教訓：**selector 要夠精確**。

> Q: targetPort: grpc 但 Pod 沒設 ports.name，套用會成功嗎？打請求會通嗎？
> A: 套用會成功（K8s 不檢查），但請求**不通**。Service 路由時找不到「叫 grpc 的 port」就失敗。`kubectl get endpoints` 會看到 IP 但 Ports: <none>。
> 教訓：**targetPort 用 name 引用時，Pod 的 ports 必須有對應的 name**。

> Q: targetPort: 7003 但程式實際監聽 9000（沒同步），套用會成功嗎？打請求會通嗎？
> A: 套用會成功（K8s 不會去戳 Pod 看程式真的監聽什麼），但請求**不通**（程式沒在 7003 上）。
> 重要原則：**「會套用」≠「會通」**，K8s 只檢查 YAML 語法，runtime 通不通要實測。

> Q: cosmic-void 的 auth-service 該用哪種 Service type？
> A: **ClusterIP**。理由：只給內部其他服務（api-gateway / game-service 等）連，不該對外暴露。
> 延伸原則：**最小暴露原則**——只有 api-gateway 用 LoadBalancer 對外，其他全 ClusterIP。

> Q: ExternalName 把外部 API 包成 Service，好處是什麼？
> A: 不只彈性（環境切換不用改程式碼），還有**資安**：
>   1. 集中管理外部 endpoint（換 endpoint 改一個地方）
>   2. 可以加 NetworkPolicy 控制誰能連外
>   3. 審計容易（所有外部連線都過 K8s）

> Q: 兩個 namespace 都有 auth-service Service，從 default 的 Pod 打 auth-service 會打到誰？
> A: 打到 **default 的 auth-service**。規則：「不寫 namespace = 自己的 namespace」。要連別的 namespace 必須明確寫（`auth-service.production`）。

> Q: 「Could not resolve host postgres」要怎麼 debug？
> A: 這是 DNS 解析失敗（不是連線失敗）。debug：
>   1. `kubectl run -it --rm dns-test --image=busybox:1.28 --restart=Never -- nslookup postgres`
>   2. 看 Service 在不在、namespace 對不對
>   3. 檢查 CoreDNS 自己是否活著
> ⚠️ 區分三種失敗：「Could not resolve」(DNS) / 「Connection refused」(沒 Pod 接) / 「Timeout」(NetworkPolicy 或 Pod 沒回)

## 關鍵理解

### 1. Service 解決的核心問題：Pod IP 不穩定

Pod 是 ephemeral 的（隨時死、IP 變）。如果直接記 Pod IP，Pod 重啟一次連線就斷。Service 在前面當「不會變的代理」，提供：
- 固定 cluster IP（虛擬 IP）
- DNS 名稱（CoreDNS 自動建立）
- 負載平衡（自動分散到多個健康 Pod）

### 2. Service 跟 Deployment 是並列關係，不是上下

```
Deployment       ← 管「Pod 怎麼跑」（生命週期）
   │
   │ 生出
   ▼
Pod, Pod, Pod    ← 實際在跑的東西
   ▲
   │ 找到（用 selector）
   │
Service          ← 管「怎麼被連到」（網路）
```

一個服務通常**同時需要兩個**：Deployment + Service。兩個都用 labels selector 找 Pod，但目的不同：
- Deployment selector：「這些是我管的 Pod」（增刪改）
- Service selector：「這些是我要路由流量的 Pod」（路由）

### 3. Service 的 selector 寫法跟 Deployment 不同

```yaml
# Deployment（要包 matchLabels）
selector:
  matchLabels:
    app: cosmic-void

# Service（直接寫）
selector:
  app: cosmic-void
```

Service 是 v1（最早期 API），用簡單格式。Deployment 是 apps/v1（較新），用 matchLabels（還支援 matchExpressions）。

### 4. Service 沒有 template

Deployment 的 `selector` + `template` 是配對使用，Service 只有 selector 沒 template，因為 Service 不負責生 Pod，只負責「找已經存在的 Pod」。

### 5. Endpoints 是流量真相

Service 內部維護「Endpoints 清單」，只有通過 readinessProbe 的 Pod 才在清單裡。
- `kubectl get endpoints xxx` 看清單
- `<none>` = selector 對不上 Pod（debug 第一步）
- 有 IP = Service 確實有對到 Pod

### 6. K8s「apply 不檢查 runtime」

```
apply 階段：只檢查 YAML 語法
runtime 階段：實際能不能通要看 Endpoints / Pod 監聽 / probe
```

**「apply 成功」≠「服務能通」**。這是 K8s debug 最重要原則。

### 7. port 三層概念（最容易暈）

| 在哪 | 欄位 | 是什麼 |
|---|---|---|
| Service | `port` | 對外開的 port（client 撥的代表號）|
| Service | `targetPort` | 轉到 Pod 的哪個 port |
| Deployment（Pod 模板）| `containerPort` | 容器內程式實際監聽的 port |

**為什麼分三層**：讓「對外」跟「對內」可以不同。例如對外 80 好記，對內程式跑 7003。

### 8. targetPort 用 name 引用是業界最佳實踐

```yaml
# Pod
ports:
- containerPort: 7003
  name: grpc          # ← 取 name

# Service
ports:
- port: 7003
  targetPort: grpc    # ← 用 name 引用，不寫數字
```

**好處**：未來程式改 port 數字（7003→9000），只改 Deployment 即可，Service 不用動。

### 9. 四種 type 的選擇

| type | 用途 | 對外 | cosmic-void 用例 |
|---|---|---|---|
| **ClusterIP** | 內部通訊（預設）| ❌ | auth / notification / game / postgres / rabbitmq |
| **NodePort** | 簡單外部存取 | ✅ Node IP:30000-32767 | 本機測試 |
| **LoadBalancer** | 雲端正式對外 | ✅ 雲端 LB（要錢）| api-gateway（正式環境）|
| **ExternalName** | 包外部 DNS | N/A | 第三方 API（Stripe / SendGrid）|

**最小暴露原則**：只有真正要對外的服務（api-gateway）用 LoadBalancer，其他全 ClusterIP。

### 10. K8s 內部 DNS 完整流程

```
auth-service.default.svc.cluster.local
└────┬────┘ └──┬──┘ └┬┘ └───────┬─────┘
   ①         ②     ③         ④
① Service name
② Namespace
③ 永遠是 svc
④ Cluster domain（預設 cluster.local，多 cluster 才會改）
```

Pod 內 `/etc/resolv.conf` 有 `search` 設定，會自動補後綴：
- 寫 `auth-service` → 自動嘗試 `auth-service.<my-ns>.svc.cluster.local`
- **同 namespace 可以省略，跨 namespace 必須寫**

### 11. 完整連線流程（用 cosmic-void 為例）

```
1. auth-service Pod 啟動，envFrom 注入 DB_HOST=auth-service-db
2. Go 程式 sql.Open("postgres://user:pass@auth-service-db:5432/...")
3. DNS 解析 auth-service-db
   → CoreDNS 回 Service cluster IP 10.96.7.42
4. 連線到 10.96.7.42:5432
5. kube-proxy 攔截，查 Endpoints 清單
6. 隨機挑一個 readiness ✅ 的 Pod，轉發
7. Postgres Pod 收到連線 ✅
```

## 程式碼 / 設定

### service.yml（最終版）

```yaml
# game-server/auth-service/k8s/service.yml
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: default
  labels:
    app: cosmic-void
    component: auth-service
spec:
  selector:
    app: cosmic-void
    component: auth-service
  type: ClusterIP
  ports:
  - name: grpc
    port: 7003
    targetPort: grpc          # 用 name 引用 Pod 的 ports.name
    protocol: TCP
  - name: http
    port: 8081
    targetPort: http
    protocol: TCP
```

### 跟 deployment.yml 的契約對齊

```yaml
# deployment.yml 的 Pod ports（命名）
ports:
- containerPort: 7003
  name: grpc       # ← Service 引用這個
- containerPort: 8081
  name: http       # ← Service 引用這個
```

兩邊靠「port name」互相串聯，未來改 port 數字只動 Deployment。

## 踩過的坑

- 問題：誤以為「selector 條件少 = 範圍窄」
  解法：理解「子集匹配」規則 —— selector 是「最低門檻」，條件越多範圍越窄
  為什麼：Q1 思考題答錯。selector 寫 `app: cosmic-void` 會匹配「所有有這個標籤的 Pod」，包括 notification / game 等其他系統元件。**寫 Service 時 selector 必須夠精確**，避免流量導到不該去的地方。

- 問題：誤以為「targetPort name 對不上會 apply 失敗」
  解法：理解 K8s 的 apply 階段不檢查 runtime
  為什麼：Q2 思考題答錯。K8s apply 只檢查 YAML 語法。targetPort: grpc 但 Pod 沒設 name，apply 會成功，但 Endpoints 的 Ports 會是 <none>，請求 connection refused。

- 問題：誤以為「targetPort 寫 7003 但程式跑 9000 → apply 失敗」
  解法：同上 —— apply 不會去戳 Pod 看程式
  為什麼：Q3 思考題答錯。K8s 完全不知道你的程式在哪個 port。**「會 apply」≠「會通」是 K8s debug 第一原則**。

- 問題：第一版 service.yml 帶了範本自動補的 `sessionAffinityConfig`
  解法：刪掉
  為什麼：`sessionAffinity: None` 跟 `sessionAffinityConfig.clientIP.timeoutSeconds` 邏輯衝突。VS Code K8s 擴充會自動補所有可能欄位，但很多是預設值，**省略反而更清楚**。

- 問題：兩個 ports 的欄位順序不一致（一個 protocol 在中間、一個在最後）
  解法：統一順序 `name → port → targetPort → protocol`
  為什麼：K8s 不在乎，但人類閱讀會困惑。業界慣例是這個順序。

- 問題：debug DNS 失敗用了 `kubectl get endpoints`
  解法：分清楚兩種失敗 —— DNS 失敗用 `nslookup`，連線失敗才用 `get endpoints`
  為什麼：「Could not resolve host」是 DNS 階段失敗，連 IP 都沒拿到，這時 endpoints 看再多都沒用。要從 `kubectl run --image=busybox:1.28 -- nslookup` 開始查。

- 問題：思考題 Q3 把「cluster.local 能改嗎」答成「用 version 標籤分流」
  解法：分清楚兩個概念
  為什麼：cluster.local 是「DNS 後綴」，多 cluster 環境才會改（cosmic-void 不會碰）；version 標籤是「selector 細分流量」，是藍綠/金絲雀部署的技巧。兩個都跟 selector 有關，但解決不同問題。

## 待釐清

- [ ] 之後寫 cosmic-void 其他服務的 Service 時，要不要把基礎設施（postgres / rabbitmq / consul）放到獨立 namespace？（純業務 vs 基礎設施分 namespace 是業界慣例）
- [ ] api-gateway 之後對外要用 LoadBalancer 還是 NodePort（看部署環境是雲端還是本機 minikube）
- [ ] Headless Service（type: ClusterIP, clusterIP: None）什麼時候用？目前只知道是 StatefulSet 用的，具體場景待研究
- [ ] sessionAffinity 什麼時候真的需要 ClientIP 模式？（cosmic-void 目前的 stateless gRPC 應該不需要）
- [ ] 多 cluster / cluster federation 是什麼？什麼規模才需要？

## 相關專案檔案

- `game-server/auth-service/k8s/service.yml` ← 這次寫的
- `game-server/auth-service/k8s/deployment.yml` ← 提供 ports.name 給 Service 引用
- `game-server/auth-service/k8s/configmap.yml` ← DB_HOST/RABBITMQ_HOST 等都會是 Service name

## 相關 learning notes

- [[deployment]] — Service 跟 Deployment 是並列關係，必須同時懂
- [[auth-service-deployment-practice]] — 寫 Deployment 時的踩坑（labels / envFrom）
- [[kubectl-cheatsheet]] — 操作 Service 的 kubectl 指令
