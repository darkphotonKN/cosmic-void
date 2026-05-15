---
topic: gke-deployment
subtopic: phase0-blockers
date: 2026-05-15
extracted-to-vault: ""
---

# Phase 0 — 部署前必修 Blockers

部署 cloud 前發現代碼裡有兩個會讓 deploy 直接掛掉的問題，必須先解。

## Blocker 1：listener bind 到 loopback

### 症狀

所有 main.go 都這樣寫：

```go
listener, err := net.Listen("tcp", "localhost:"+grpcAddr)
```

`localhost` 解析成 `127.0.0.1`，**只接 loopback 流量**。Pod 內 listener 只看自己 namespace 內的 127.0.0.1。

### 為什麼這在 docker-compose 沒問題、上 k8s 會壞

| 情境 | 是否會壞 |
|---|---|
| docker-compose（host network 或 bridge） | docker bridge 內就是 127.0.0.1 接得到 |
| K8s Service → Pod | Service 把流量送到 **Pod IP**（不是 localhost），listener 拒絕 |
| LB → NEG → Pod IP | 同上 |

### 修法

```diff
- listener, err := net.Listen("tcp", "localhost:"+grpcAddr)
+ listener, err := net.Listen("tcp", ":"+grpcAddr)
```

`:port` 等於 `0.0.0.0:port`，bind 到所有介面。**不會有額外的安全風險**因為 k8s NetworkPolicy + firewall 才是真正擋外部的層。

### 影響的 7 個檔案

- auth-service / items-service / game-service / stats-service / notification-service / payment-service / example-service
- **api-gateway 不用改**：它用 gin 的 `router.Run(":port")` 已經 bind all interfaces

## Blocker 2：6 個 service 的 k8s manifests 不存在

### 現狀盤點

| Service | 既有 manifests | 需要補的 |
|---|---|---|
| auth-service | 完整（configmap、deployment、postgres、redis、rabbitmq） | configmap K8S_NAMESPACE 換值 |
| api-gateway | 只有 service.yml + ingress.yml | deployment.yml + configmap + secret.yml.example |
| 其他 6 個 | 全部缺 | deployment + service + configmap + secret 4 件套 |

### 為什麼一開始 manifests 缺

之前的 git 紀錄是 `add default k8s` commit 後沒繼續做。auth-service 是 reference impl，其他要照樣建。

### 決策：放寬式架構

- 一開始用 **共享 PostgreSQL Pod**（auth-service-db）+ 7 個 logical DB（per-service DB name）
- 共享 `auth-service-secrets` 的 RABBITMQ_PASS（避免重複維護）
- ConfigMap 統一 `K8S_NAMESPACE: cosmic-void`

### Service Map 對齊（重點）

新建的 Service.metadata.name + port **必須跟** [common/discovery/k8s/k8s.go](../../game-server/common/discovery/k8s/k8s.go) 的 `serviceMap` 一致：

| 邏輯名 (Discover key) | k8s Service name | gRPC port |
|---|---|---|
| `auth` | `auth-service` | 7003 |
| `payments` | `payment-service` | 7021 |
| `items` | `items-service` | 7013 |
| `stats` | `stats-service` | 7011 |
| `notification` | `notification-service` | 7077 |
| `examples` | `example-service` | 7010 |
| `game` | `game-service` | 7004 |
| `api-gateway` | `api-gateway` | 7001 |

不一致 → `Discover` 回 `unknown service` 或 DNS 解不到。

## 額外決策：移除 manifests 的 `namespace: default`

舊 manifests 都寫 `namespace: default`。

**問題**：`kubectl apply -n cosmic-void -f file.yml` 對寫 `namespace: default` 的檔案會 reject（namespace 衝突 error）。

**修法**：用 sed 一次清掉：

```bash
# .yml
find game-server -path '*/k8s/*.yml' -exec sed -i '' '/^  namespace: default$/d' {} \;
# .yml.example / .yml.template（find 沒抓到，手動補一輪）
find game-server -type f \( -name '*.yml.example' -o -name '*.yml.template' \) -path '*/k8s/*' \
  -exec sed -i '' '/^  namespace: default$/d' {} \;
```

新建的 manifests 一律**不寫 `namespace:` 欄位**，靠 `kubectl apply -n cosmic-void` 統一指定。

## 驗證

```bash
# 8 個 go build 都過
for svc in auth-service api-gateway items-service game-service \
           stats-service notification-service payment-service example-service; do
    sub="cmd/server"; [ "$svc" = "api-gateway" ] && sub="cmd"
    (cd $svc && go build -o /tmp/build-$svc ./$sub) \
        && echo "$svc OK" || echo "$svc FAILED"
done
```
