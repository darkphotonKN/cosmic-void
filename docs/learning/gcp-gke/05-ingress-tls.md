---
date: 2026-05-15
topic: gcp-gke
subtopic: 05-ingress-tls
extracted-to-vault:
---

# 05 — GCE Ingress + Cloudflare TLS 演進

## 演進三階段

| 階段 | 架構 | 結果 |
|---|---|---|
| v1 | ingress-nginx + cert-manager + Let's Encrypt | **沒選**（架構討論結束後）|
| v2 | GCE Ingress + ManagedCertificate + GCE LB | 部署了一輪，cert 卡 Provisioning |
| v3（最終）| GCE Ingress + Cloudflare proxy + Flexible SSL | ✓ 成功 |

## 為什麼三次都不同

### 為什麼沒選 ingress-nginx

最初預設方案。但用戶問「為什麼不直接整合 GCP 服務」→ 改 v2。

**ingress-nginx 缺點**：
- TLS 終結在 cluster 內（多一個 nginx pod 顧 CPU/RAM）
- cert-manager 又一個 controller pod
- 跟 GCP 沒整合，沒 GCP 邊緣 CDN/DDoS

### 為什麼 v2 卡死

GCE Ingress + ManagedCertificate 是 production-grade GCP-native 路線：
- Static IP reserve（`gcloud compute addresses create cosmic-void-ingress --global`）
- Ingress annotations:
  ```
  kubernetes.io/ingress.class: "gce"
  kubernetes.io/ingress.global-static-ip-name: "cosmic-void-ingress"
  networking.gke.io/managed-certificates: "cosmic-void-cert"
  networking.gke.io/v1beta1.FrontendConfig: "cosmic-void-frontend-config"
  ```
- ManagedCertificate resource 指定 domain
- FrontendConfig 處理 HTTP→HTTPS redirect

**卡點**：ManagedCertificate 進入 `FailedNotVisible` 狀態。原因是用戶 DNS 設在 Cloudflare 且預設**橘雲 proxied**，DNS 解析回 Cloudflare IP（104.21.x.x），不是 GCP LB IP（34.111.74.79）。Google 的 ACME challenge probe 經過 Cloudflare 走不到。

可以叫用戶把雲朵點成灰色（DNS-only）讓 Google 看到原始 IP → 可以驗 → cert 簽下來。但這樣就**失去 Cloudflare proxy 的 CDN/DDoS 防護**。

### 為什麼 v3 才是「最聰明」

用戶問：「Cloudflare 已經提供 HTTPS 了，為什麼還要 ManagedCert？」

確實 — Cloudflare proxy 自己會給 user 一張 Universal Cert，**完全不需要 Google 再來一張**。架構簡化成：

```
User → Cloudflare HTTPS (CF cert)
       ↓ HTTP (Flexible mode)
       GCP LB → GKE Pod
```

額外好處：
- 拿到 Cloudflare 免費 CDN（靜態 asset 快取在邊緣）
- DDoS 邊緣防護（攻擊先打到 Cloudflare 而不是 GCP）
- 省 Cloudflare → GCP 那段 egress（CF 不算進 GCP egress）

代價：
- TLS 終結點從 GCP 移到 Cloudflare（不在你 control 內，但小專案無感）
- Cloudflare → Origin 是 HTTP（明文，但走 CF 私有網路）

## v3 實作

### Cloudflare 端（用戶手動操作）

1. cosmicvoid.uk → SSL/TLS → Overview → **Flexible** 模式
2. DNS records 三條（cosmicvoid.uk、www.cosmicvoid.uk、api.cosmicvoid.uk）全部點成**橘雲 Proxied**
3. 三條 A record 的 origin IP 填 `34.111.74.79`

### GCE Ingress（無 TLS）

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cosmic-void-ingress
  annotations:
    kubernetes.io/ingress.class: "gce"
    kubernetes.io/ingress.global-static-ip-name: "cosmic-void-ingress"
    # 無 managed-certificates、無 FrontendConfig
spec:
  rules:
  - host: api.cosmicvoid.uk
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-gateway
            port:
              name: http
  - host: cosmicvoid.uk
    http:
      paths:
      - path: /game/ws         # WebSocket → game-service-ws
        pathType: Prefix
        backend:
          service:
            name: game-service-ws
            port:
              name: http
      - path: /                # 其他 → game-client
        pathType: Prefix
        backend:
          service:
            name: game-client
            port:
              name: http
  - host: www.cosmicvoid.uk
    # 同 cosmicvoid.uk
```

### 為什麼**必須**用 NEG（container-native LB）

GCE Ingress 預設用 NodePort 模式：LB → Node IP:NodePort → kube-proxy → Pod。多一跳。

NEG mode：LB → Pod IP 直接打。Service 需 annotation：
```yaml
annotations:
  cloud.google.com/neg: '{"ingress": true}'
```

每個 Ingress backend 對應的 Service 都要加。

### BackendConfig — 設定 LB 的 health check 路徑

GCE LB 預設 health check `GET /` 期望 200。但 api-gateway 用 gin router 的所有 routes 都在 `/api/*`，`/` 直接 404 → backend UNHEALTHY → 整個 cluster 對外 502。

修法：
1. **加 `/healthz` 端點到 api-gateway**（[api-gateway/config/routes.go](../../../game-server/api-gateway/config/routes.go)）：
   ```go
   router.GET("/healthz", func(c *gin.Context) {
       c.String(200, "ok")
   })
   ```
2. **創建 BackendConfig** 明確指定 health check path：
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
   ```
3. **Service annotation 引用 BackendConfig**：
   ```yaml
   cloud.google.com/backend-config: '{"default": "api-gateway-backend-config"}'
   ```

GCE LB 通常需要 1-2 分鐘才把新 BackendConfig push 到 BackendService。等待時 backend 顯示 UNHEALTHY，正常。

### Firewall 別忘了

GKE 自動建 `k8s-fw-l7--<hash>` 規則 allow Google 的 health check IP 段（35.191.0.0/16, 130.211.0.0/22）打 pod port。**如果你看到 backend UNHEALTHY 但 `/healthz` 從 pod 內 OK**，可能是 firewall 沒涵蓋。手動加：

```bash
gcloud compute firewall-rules create k8s-cosmic-void-hc-7001 \
  --network=default --action=ALLOW --direction=INGRESS \
  --source-ranges=35.191.0.0/16,130.211.0.0/22 \
  --target-tags=<NODE_TAG> --rules=tcp:7001
```

但通常 GKE 已經建好，重複建多餘。後來確認 `k8s-fw-l7--*` 已涵蓋 7001+8080，我建的多餘規則最後刪掉。

## Rollout 策略對「資源不夠」的影響

api-gateway deployment.yml 起初用：
```yaml
strategy:
  rollingUpdate:
    maxSurge: 1
    maxUnavailable: 0
```

意義：新 pod 先 Ready 才砍舊。但因為新 pod 帶了 NEG readiness gate（LB 認可才 Ready），LB 又要靠 backend HEALTHY 才認可 → 死鎖。**改成**：

```yaml
strategy:
  rollingUpdate:
    maxSurge: 0
    maxUnavailable: 1
```

舊的先砍，新的才上。會有短暫 downtime 但可破死鎖。

## 「公網都能直接打 LB IP」的安全問題（在 v3 變嚴重）

Cloudflare proxy 把流量導進 GCP LB，但 GCP LB 的 IP `34.111.74.79` 對全網開放。攻擊者直接 `curl http://34.111.74.79/` 繞 Cloudflare。

**為什麼 VPC firewall 不行**：L7 HTTP(S) LB 在 GCP **邊緣**，封包到達後源 IP 已是 Google 內部 IP（VPC firewall 看的是 L3/L4 源 IP，不是 `X-Forwarded-For` header）。

唯一解：**Cloud Armor security policy**（GCP 的 WAF，類比 AWS WAF）。價格 $5/policy + $1/rule + $0.75/M requests ≈ $6/月。詳見 [08-firewall-security.md](08-firewall-security.md)。
