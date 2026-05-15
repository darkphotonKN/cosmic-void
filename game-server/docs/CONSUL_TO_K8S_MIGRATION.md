# Consul → Kubernetes Service Discovery 替換紀錄

> 日期：2026-05-14
> 方案：A（薄包裝，保留 `discovery.Registry` interface）
> 替換對象：8 個 Go 服務（auth / payment / items / stats / notification / example / game / api-gateway）

---

## 1. 為何要換

原本所有 gRPC service-to-service 通訊都透過 Consul 做服務發現：
- 每個服務啟動時 `consul.NewRegistry(...)` → `Register` → 背景 `HealthCheck` goroutine（每秒 TTL renewal）→ `defer Deregister`。
- 呼叫方透過 `discovery.ServiceConnection()` 向 Consul 查 healthy instance 清單，隨機挑一個 dial gRPC。

部署到 GKE 後這些事情 Kubernetes 自己會做：
- **註冊** → Service + Endpoints controller。Pod 一上線（passing readinessProbe），Endpoints 就會自動加入。
- **健康檢查** → `livenessProbe` / `readinessProbe`。Pod 不健康自動從 Endpoints 拔除。
- **發現** → kube-dns / CoreDNS 提供 in-cluster DNS：`<svc>.<ns>.svc.cluster.local`。
- **L4 負載均衡** → kube-proxy（iptables / IPVS）對 ClusterIP 流量自動散到各 Pod。

繼續跑 Consul 在 GKE 上不是錯，但是是「在 K8s 內再造一套 service mesh」，徒增運維成本（要顧 Consul Pod、要顧 cluster bootstrap、要顧 token、要顧 ACL ...），對學習專案沒必要。

---

## 2. 兩個替換方案的比較（決策紀錄）

### 方案 A：薄包裝（採用）

- 保留 `common/discovery/Registry` interface。
- 新增 `common/discovery/k8s/k8s.go`：`Register` / `Deregister` / `HealthCheck` 全部 **no-op**；`Discover` 回傳單一 DNS 名稱 `[]string{"auth-service.default.svc.cluster.local:7003"}`。
- 邏輯名 → k8s Service:port 的對應寫死在 `k8s` package 內部的 `serviceMap`。

**優點**
- main.go / gateway client / 內部 grpc client 等呼叫點全部不動，diff 最小。
- 之後想再換回多 registry 抽象（譬如 Hybrid Consul+k8s）也容易回頭。
- 學習路徑漸進，風險可控。

**缺點**
- 保留沒實際作用的 `Register/HealthCheck` 呼叫路徑（每秒一次 no-op heartbeat，浪費但成本可忽略）。
- 放棄 gRPC client-side load balancing — `ServiceConnection` 每次拿到的都是同一個 ClusterIP，由 kube-proxy 在 L4 做隨機散發。對短連線 OK，對長連線 streaming RPC 會集中在第一個被選中的 Pod。

### 方案 B：直連 + gRPC DNS resolver（未採用）

- 移除 `Registry` 抽象、每個 client 改成 `grpc.NewClient("dns:///<svc>:port", round_robin)` 配 Headless Service。

**為何先不做**：改動面大，且要先把 dial-per-call 模式改成長連線重用，跟「拔掉 Consul」這個目標 orthogonal。等到效能真的不夠時再做。

---

## 3. 完整改動清單

### 3.1 新增

| 檔案 | 內容 |
|---|---|
| [common/discovery/k8s/k8s.go](../common/discovery/k8s/k8s.go) | `k8s.Registry` 實作，`NewRegistry(namespace)` constructor，內含 `serviceMap` |

`serviceMap` 內容（程式碼裡的邏輯名 → k8s Service 名 + gRPC port）：

| 邏輯名 | k8s Service 名 | gRPC port |
|---|---|---|
| `auth` | `auth-service` | 7003 |
| `payments` | `payment-service` | 7021 |
| `items` | `items-service` | 7013 |
| `stats` | `stats-service` | 7011 |
| `notification` | `notification-service` | 7077 |
| `examples` | `example-service` | 7010 |
| `game` | `game-service` | 7004 |
| `api-gateway` | `api-gateway` | 7001 |

> 任何新增服務 / 改 port 都要回頭更新這張表，否則 Discover 會回 `unknown service`。

### 3.2 移除

| 路徑 | 為什麼 |
|---|---|
| `common/discovery/consul/` (整個資料夾) | 不再需要 Consul SDK 包裝 |
| `auth-service/k8s/consul.yml` | 不再部署 Consul Pod 到叢集 |
| `common/go.mod` 的 `github.com/hashicorp/consul/api` 與連帶 indirect deps | 用 `go build` 跑過後自動清掉（serf / armon / mitchellh 等都跟著走） |

### 3.3 修改

#### 3.3.1 每個 main.go 的固定 pattern

8 個 main.go 都是同樣三處改動：

**Before**
```go
import (
    "github.com/darkphotonKN/cosmic-void-server/common/discovery"
    "github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
)

var (
    consulAddr = commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")
)

registry, err := consul.NewRegistry(consulAddr, serviceName)
if err != nil {
    log.Fatal("Failed to create Consul registry")
}
```

**After**
```go
import (
    "github.com/darkphotonKN/cosmic-void-server/common/discovery"
    "github.com/darkphotonKN/cosmic-void-server/common/discovery/k8s"
)

var (
    k8sNamespace = commonhelpers.GetEnvString("K8S_NAMESPACE", "default")
)

registry, err := k8s.NewRegistry(k8sNamespace)
if err != nil {
    log.Fatal("Failed to create k8s registry")
}
```

`Register` / `HealthCheck` / `Deregister` 的呼叫保留不動 — 由 no-op registry 接住，效益是「保留 Registry 抽象」、缺點是「跑沒有意義的 heartbeat goroutine」。Trade-off 屬於 diff size vs runtime cleanliness，本次選 diff size。

**影響的檔案**：
- [auth-service/cmd/server/main.go](../auth-service/cmd/server/main.go)
- [api-gateway/cmd/main.go](../api-gateway/cmd/main.go)
- [items-service/cmd/server/main.go](../items-service/cmd/server/main.go)
- [game-service/cmd/server/main.go](../game-service/cmd/server/main.go)
- [stats-service/cmd/server/main.go](../stats-service/cmd/server/main.go)
- [notification-service/cmd/server/main.go](../notification-service/cmd/server/main.go)
- [payment-service/cmd/server/main.go](../payment-service/cmd/server/main.go)
- [example-service/cmd/server/main.go](../example-service/cmd/server/main.go)

#### 3.3.2 ConfigMap

[auth-service/k8s/configmap.yml](../auth-service/k8s/configmap.yml)

```diff
- CONSUL_ADDR: "consul:8500"
+ K8S_NAMESPACE: "default"
```

#### 3.3.3 Test 檔（順手修）

兩個 test 檔本來就因為其他理由 broken（mock interface 不對等），但裡頭引用 consul package 會在刪除 consul 後造成 compile error，所以一併把 import 換掉。**沒有把 test 本身的問題修好**，那是另一個工作。

- [game-service/internal/gameserver/integration_test.go](../game-service/internal/gameserver/integration_test.go) — import / `NewRegistry` 換成 k8s
- [game-service/internal/game/service_test.go](../game-service/internal/game/service_test.go) — 移除沒被使用的 `consulAddr` 變數宣告

#### 3.3.4 文件 docstring

[api-gateway/internal/gateway/example/client.go](../api-gateway/internal/gateway/example/client.go) 開頭的範例註解：

```diff
-    registry := consul.NewRegistry(...)
+    registry := k8s.NewRegistry(...)
```

---

## 4. 驗證

8 個服務都用 `go build` 過測：

```bash
cd game-server
for svc in auth-service api-gateway items-service game-service \
           stats-service notification-service payment-service example-service; do
    sub="cmd/server"; [ "$svc" = "api-gateway" ] && sub="cmd"
    (cd $svc && go build -o /tmp/build-$svc ./$sub) \
        && echo "$svc OK" || echo "$svc FAILED"
    rm -f /tmp/build-$svc
done
```

全部回報 `OK`。

> Test 檔還是會編譯失敗，但那是既有的 mock interface 問題（mockQueueService 缺 `AddPlayer`、MockEventEmitter 的 `PublishMatchComplete` 簽名錯等），跟本次替換無關。

---

## 5. 已知 blocker（**這次沒做，但部署上 GKE 前必須處理**）

### 5.1 listener bind 到 loopback

所有 main.go 都是：
```go
listener, err := net.Listen("tcp", "localhost:"+grpcAddr)
```

`localhost` 會解析成 `127.0.0.1`，只接 loopback 流量。Pod IP 的請求進不來，k8s Service 把流量轉到 Pod 上會直接被拒。

**該改成**：
```go
listener, err := net.Listen("tcp", ":"+grpcAddr)
```

（或顯式寫 `"0.0.0.0:"+grpcAddr`，效果一樣。）

為何本次不一起改：跟 consul 替換 orthogonal，且改 8 個檔案；想保持本次 PR scope 乾淨。但這個是真 blocker，下一個 PR 要做。

### 5.2 k8s manifests 大部份缺漏

目前只有：
- `auth-service/k8s/` — 完整（configmap / deployment / service / postgres / redis / rabbitmq）
- `api-gateway/k8s/` — 只有 service.yml 和 ingress.yml

**還沒寫的**（6 個）：
- payment-service
- items-service
- stats-service
- notification-service
- example-service
- game-service

任何一個沒部署，那個服務的 DNS 名解不到，呼叫它的 service 就會收到 `ServiceConnection` 的錯。

優先順序建議：照「被呼叫的次數」由多到少做 — `auth-service` 已經有了，下一個做 `items-service`（被 game-service 呼叫）、再做 `stats-service` / `notification-service`（被 api-gateway 呼叫）。

### 5.3 跨 module 的怪異 import

`common/constants/types/item.go` import 了 `game-service/grpc/items`，這讓 `cd common && go mod tidy` 失敗（會去 git 上找 `cosmic-void-server` repo）。本次靠 go.work 繞過，但這個依賴方向是反的（common 不該知道 game-service），應該把 grpcitems 的 client interface 抽到 common，或把 item types 改成不依賴 grpcitems。

### 5.4 dial-per-call 模式

所有 gateway client 與內部跨服務 client 都是：

```go
conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
defer conn.Close()
client := pb.NewXxxServiceClient(conn)
return client.Method(ctx, req)
```

每個 API call 都建一條新 TCP 連線、做完 close。在 GKE 上仍然 work，但每個 request 都背負 connection setup 成本。

修法是把 `conn` 抽到 client 建構時建立、長期持有，搭配 `dns:///` + `round_robin` 做 client-side LB。這是「方案 B」的內容，等到效能不夠時再做。

---

## 6. 之後要小心的事

- **新增服務**：要同時更新 `common/discovery/k8s/serviceMap` 和 `<svc>/k8s/service.yml` 兩處。兩邊的 Service name 與 port 必須對齊，否則 Discover 會 fail。
- **改 port**：同上。
- **改 namespace**：用 env `K8S_NAMESPACE` 注入，預設 `default`。多 namespace 部署時記得每個 Deployment 都注入正確的值。
- **debug DNS 解析失敗**：先在 Pod 內跑 `nslookup auth-service.default.svc.cluster.local`，確認 CoreDNS 有 Endpoints。如果 nslookup OK 但 gRPC dial fail，問題就是 5.1 的 listener bind。

---

## 7. Rollback

如果發現 k8s 路徑有問題、需要短期回滾：

1. `git revert` 本次 commit。
2. `auth-service/k8s/consul.yml` 重新部署 Consul Pod。
3. 把 ConfigMap 的 `K8S_NAMESPACE` 改回 `CONSUL_ADDR: "consul:8500"`。
4. 重啟所有 Deployment。

但只要 listener bind 還是 `localhost`，rollback 後在 k8s 上仍然不會通 — 因為 Consul 注入的 Pod IP 給呼叫端，呼叫端連 Pod IP 還是會被 loopback 擋。換句話說，5.1 是「跟 Consul 與否無關的部署阻塞」，不能靠 rollback 解決。
