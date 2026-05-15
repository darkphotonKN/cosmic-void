---
topic: gke-deployment
subtopic: ingress-vs-cloudflare
date: 2026-05-15
extracted-to-vault: ""
---

# Ingress 與 TLS：從 ManagedCert 到 Cloudflare 的路徑切換

## 第一次的選擇：GCE Ingress + ManagedCertificate

User 一開始選了「混合型：GCE Ingress + ManagedCert，但 cluster 保持 public」。

理由：
- 跳過 ingress-nginx + cert-manager（cluster 內少兩個 Pod）
- 用 Google 託管的 SSL cert（auto-renew）
- 跳過 Cloud NAT（私 cluster 才需要，~$32/月）

設置：

```bash
# Reserve global static IP
gcloud compute addresses create cosmic-void-ingress --global
# → 34.111.74.79

# 驗證 GKE 內建 CRDs（裝 HttpLoadBalancing addon 就有）
kubectl get crd managedcertificates.networking.gke.io
kubectl get crd backendconfigs.cloud.google.com
```

Ingress 設定：

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: "gce"
    kubernetes.io/ingress.global-static-ip-name: "cosmic-void-ingress"
    networking.gke.io/managed-certificates: "cosmic-void-cert"
    networking.gke.io/v1beta1.FrontendConfig: "cosmic-void-frontend-config"  # HTTP→HTTPS 強制
```

ManagedCertificate：

```yaml
apiVersion: networking.gke.io/v1
kind: ManagedCertificate
spec:
  domains:
    - api.cosmicvoid.uk
```

## 第一個踩雷：Cloudflare proxy 的橘雲讓 DNS 回 Cloudflare IP

User 在 cosmicvoid.uk 的 DNS（Cloudflare 託管）加了 A record `api → 34.111.74.79`，但 dig 出來是：

```
api.cosmicvoid.uk:
104.21.90.53
172.67.153.133
```

這兩個是 **Cloudflare edge IP**，不是 34.111.74.79！

### 不是 bug，是 proxy 設計

當 Cloudflare 的 DNS record 設成「proxied」（橘雲），public DNS 回的是 CF 自己的 anycast IP。流量先到 CF，CF 內部 forward 到 origin（你設的 34.111.74.79）。

要看 origin IP 要在 Cloudflare 後台「Edit DNS record」展開那筆才看得到。

### 對 ManagedCertificate 的影響

ManagedCert 用 HTTP-01 challenge 驗證 domain：Google ACME 對 `api.cosmicvoid.uk` 發 GET request，要看到回 `.well-known/acme-challenge/<token>` 的特定 response。

但 DNS 回的是 Cloudflare IP，Google 的 ACME 探測打到 CF。CF 在 Flexible mode 下會試圖把它 forward 到 origin，但 LB 還沒裝好 cert challenge endpoint → fail。

ManagedCertificate status 變 `FailedNotVisible`。

## 架構決策大改：改用 Cloudflare 處理 TLS

User 問：「為什麼還要 ManagedCert，Cloudflare proxy 已經給 HTTPS 了」

正中要害。最終架構：

```
User ──HTTPS (CF universal cert)──► Cloudflare proxy
                                      │
                                      ↓ HTTP (Flexible SSL mode)
                                    GCP HTTPS LB (port 80 only)
                                      │
                                      ↓
                                    GKE pods
```

### Cloudflare SSL Modes 對比

| Mode | User → CF | CF → Origin | Origin 需要 |
|---|---|---|---|
| **Off** | HTTP | HTTP | 不用 cert |
| **Flexible** ✓ | HTTPS | HTTP | 不用 cert |
| **Full** | HTTPS | HTTPS | 任何 cert（含自簽） |
| **Full (strict)** | HTTPS | HTTPS | CA 簽署的 cert |

我們選 **Flexible**：
- User 看到的是 HTTPS（CF universal cert）
- CF → 34.111.74.79 是 HTTP（不用 cert）
- 缺點：CF 到 Origin 這段是明文（internet 上）— 學習用可接受

## 砍掉 ManagedCert 後的 manifest

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: "gce"
    kubernetes.io/ingress.global-static-ip-name: "cosmic-void-ingress"
    # 不再有 managed-certificates / FrontendConfig
spec:
  rules:
  - host: api.cosmicvoid.uk    # Cloudflare 進來的 Host header 是這個
    http:
      paths:
      - path: /
        ...
```

實際移除：

```bash
kubectl -n cosmic-void delete managedcertificate cosmic-void-cert
kubectl -n cosmic-void delete frontendconfig cosmic-void-frontend-config

# 砍掉 ingress 重建（讓 LB rebuild 不要保留舊的 HTTPS listener）
kubectl -n cosmic-void delete ingress cosmic-void-ingress
sleep 5
kubectl -n cosmic-void apply -f api-gateway/k8s/ingress.yml
```

## 第二個踩雷：Backend UNHEALTHY

```
{"k8s1-89c4df67-cosmic-void-api-gateway-80-a25d572d":"UNHEALTHY"}
```

api-gateway 的 LB backend 一直 UNHEALTHY。原因：GCE LB 預設 health check 是 `GET /` expect 200，但 api-gateway 所有路由都在 `/api/*`，**`/` 回 404**。

### 修法 1：加 /healthz 端點

```go
// api-gateway/config/routes.go
router.GET("/healthz", func(c *gin.Context) {
    c.String(200, "ok")
})
```

### 修法 2：BackendConfig 自訂 health check

```yaml
apiVersion: cloud.google.com/v1
kind: BackendConfig
metadata:
  name: api-gateway-backend-config
spec:
  healthCheck:
    type: HTTP
    requestPath: /healthz
    port: 7001
    checkIntervalSec: 10
    timeoutSec: 5
    healthyThreshold: 1
    unhealthyThreshold: 3
```

### 修法 3：Service annotation 引用 BackendConfig

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    cloud.google.com/neg: '{"ingress": true}'
    cloud.google.com/backend-config: '{"default": "api-gateway-backend-config"}'
```

## 第三個踩雷：health check firewall

GCE LB 從 `35.191.0.0/16` + `130.211.0.0/22` 兩個網段 probe pod。Pod 在 cluster 內，VPC firewall 默認不讓這些 IP 進。

幸好 **GKE Ingress controller 會自動建 firewall rule** `k8s-fw-l7--<hash>` 開放需要的 port：

```
NAME                         SOURCE_RANGES                        ALLOW
k8s-fw-l7--89c4df675995becd  130.211.0.0/22, 35.191.0.0/16        tcp:7001,tcp:8080
```

我一開始**多手動建了一條** `k8s-cosmic-void-hc-7001`，後來發現重複，刪掉。

教訓：**先看 `k8s-fw-l7--*` 是否已自動建立再下手**。GKE 會幫你。

## 第四個踩雷：BackendService 改設定不會自動生效

我用 `kubectl apply` 改了 BackendConfig（health check path），但 LB 還是用舊設定 probe `GET /`。

原因：GCE Ingress controller 把現有 BackendService 的設定 cache 住，不會主動 sync。

**強制 re-sync 招數**：在 Service 加一個無關 annotation 觸發 controller 重看 spec：

```bash
kubectl -n cosmic-void annotate service api-gateway "cosmic-void.dev/refresh=$(date +%s)" --overwrite
```

或刪 ingress 重建（更狠但有效）。

## Path-based routing：cosmicvoid.uk + api.cosmicvoid.uk 同一個 LB

最終 Ingress 配置：

```yaml
spec:
  rules:
  - host: api.cosmicvoid.uk
    http:
      paths:
      - path: /, backend: api-gateway:80

  - host: cosmicvoid.uk
    http:
      paths:
      - path: /game/ws, backend: game-service-ws:5555    # WebSocket
      - path: /,         backend: game-client:80          # Next.js

  - host: www.cosmicvoid.uk    # 同上
```

**GCE LB URL Map 用 longest-prefix-match**，所以 `/game/ws` 自動勝過 `/`，不用擔心順序。

## DNS 設定（Cloudflare 端）

| Type | Host | Value | Proxy |
|---|---|---|---|
| A | `@` | 34.111.74.79 | orange-cloud ✓ |
| A | `www` | 34.111.74.79 | orange-cloud ✓ |
| A | `api` | 34.111.74.79 | orange-cloud ✓ |

**SSL/TLS → Overview → Flexible**

## 教訓

1. **如果你有 Cloudflare proxy，就讓它處理 TLS** — Google ManagedCert + Cloudflare 是過度設計
2. **GCE LB 預設 health check 是 GET /** — 對 API gateway 不適用，必須 BackendConfig
3. **GCE Ingress controller 不會主動 re-sync BackendService**，要 trigger annotation change
4. **Cloudflare orange-cloud DNS 回 CF IP 不是 origin IP** — 設計，不是 bug
