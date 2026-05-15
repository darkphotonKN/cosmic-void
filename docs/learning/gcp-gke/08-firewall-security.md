---
date: 2026-05-15
topic: gcp-gke
subtopic: 08-firewall-security
extracted-to-vault:
---

# 08 — Firewall 稽查、Cloud Armor 為何必要、剩下的安全強化

## Firewall 完整稽查

```
NAME                            SOURCE_RANGES                   ALLOW                       
default-allow-icmp              0.0.0.0/0                       icmp                          
default-allow-internal          10.128.0.0/9                    tcp:0-65535,udp:0-65535,icmp 
gke-cosmic-void-2afe27ea-all    10.28.0.0/14 (Pod CIDR)         all protocols              
gke-cosmic-void-2afe27ea-exkubelet  0.0.0.0/0                   (deny — kubelet 外擋)
gke-cosmic-void-2afe27ea-inkubelet  10.28.0.0/14                tcp:10255
gke-cosmic-void-2afe27ea-vms    10.128.0.0/9                    icmp,tcp:1-65535,udp:1-65535
k8s-fw-l7--<hash>               130.211.0.0/22, 35.191.0.0/16   tcp:3000,7001,8080 (LB HC)
```

已刪：
- ~~default-allow-rdp~~（Linux 用不到 RDP）
- ~~default-allow-ssh~~（改用 `kubectl exec`，不需要 SSH 進 node）

### gRPC 對外暴露檢查

gRPC ports: `7003 7004 7010 7011 7013 7021 7077` + game HTTP `5555` + api-gateway HTTP `7001`

| 流量 | 防火牆 | 結果 |
|---|---|---|
| Pod ↔ Pod（intra-cluster gRPC）| `gke-...-all`（10.28.0.0/14 全 protos）| ✓ 通 |
| Node ↔ Node | `default-allow-internal` + `vms` | ✓ 通 |
| LB → Pod HTTP 7001 / 3000 / 5555 | `k8s-fw-l7--*` | ✓ 通 |
| **0.0.0.0/0 → gRPC ports** | **沒任何規則允許** | ✓ **封閉**（gRPC 不對外）|

結論：gRPC 完全 intra-cluster，沒對外洩漏。對外只有 LB 上的 80 port。

## 為什麼 VPC firewall 擋不住「公網 → 34.111.74.79」

用戶反覆問這個。完整解答：

```
[Layer A] Internet           [Layer B] Google Edge POPs        [Layer C] Your VPC
                                ▲                                 ▲
                          L7 HTTPS LB                          Pod/Node/VM
                          IP 34.111.74.79                      在這裡
                          在這裡（Google anycast）           
```

VPC firewall 只能套在 **[Layer C] 的 instance 網路介面**，[Layer B] 在 Google 邊緣，VPC 觸不到。

當有人 `curl http://34.111.74.79/`：
1. 封包進 [Layer B] 的 LB
2. LB 在 B 層**解掉 TLS、看 Host header、跑 URL Map**
3. LB 把請求**重新打包**從 Google internal network 傳到 VPC 後端 NEG
4. 進到 [Layer C] 的 Pod

「攻擊者的源 IP」這時候塞在 `X-Forwarded-For` HTTP header 裡，封包源 IP 已是 **Google 內部 IP**。VPC firewall 看 L3/L4 源 IP，不解 header → 永遠看到 Google 內部 IP，無法區分是不是 Cloudflare 來的。

**L7 LB 的本質**：proxy 不是 passthrough。源 IP 不保留到 backend。

## 三條可行路徑

### A. Cloud Armor + Cloudflare IP allowlist（GCP WAF）

**Cloud Armor 就是 GCP 對應 AWS WAF 的產品**：
| AWS | GCP |
|---|---|
| AWS WAF | Cloud Armor |
| Security Groups | VPC Firewall Rules |
| AWS Shield | Cloud Armor (DDoS 防護內建) |

掛在 LB Backend Service 上，policy 規則：
- priority 1000: source ∈ Cloudflare CIDR → ALLOW
- priority 2147483647 (default): DENY

**成本**：$5/policy/月 + $1/rule/月 + $0.75/M requests ≈ $6-7/月。90 天 ≈ +$20。

**取捨**：production-grade，無架構改動。

### B. L4 Network LB + in-cluster nginx ingress

L4 LB 是 passthrough，**真正客戶端 IP 直接到 Pod**。VPC firewall 在 Pod node 看得到源 IP，能擋。

**代價**：
- 失去 GCE Ingress URL map / path routing
- 多 Service 要每個自己一顆 LB（每顆 $18/月）
- 或改用 cluster 內 nginx-ingress 做 path routing（多 nginx pod + cert-manager）
- TLS 終結放回 cluster

實際上等於回到原本「不選 ingress-nginx」的路線。

### C. App-layer 過濾 `X-Forwarded-For`（最便宜）

api-gateway 加 gin middleware：
```go
router.Use(func(c *gin.Context) {
    clientIP := c.Request.Header.Get("X-Forwarded-For")
    if !isCloudflareIP(clientIP) {
        c.AbortWithStatus(403)
        return
    }
    c.Next()
})
```

**好處**：免費。
**壞處**：
- 每個 backend 都得寫一份（或集中 gateway）
- TCP socket 仍能連 LB，Pod 仍要計算（DDoS 還是吃 CPU）
- WAF 級規則自己實作

## 用戶選擇

**A. Cloud Armor**（採用，+$6/月，~$20/90 天）
- 已在剩餘任務清單，下節執行

## 用戶**沒**選的安全強化

### Master Authorized Networks（免費）

把 `34.71.54.60:443`（K8s API endpoint）從 `0.0.0.0/0` 收回給特定 IP。

```bash
gcloud container clusters update cosmic-void --zone=us-central1-a \
  --enable-master-authorized-networks \
  --master-authorized-networks=<YOUR_IP>/32
```

用戶選擇「不要 — 保留 0.0.0.0/0」。理由：換網路（家 vs 公司）需手動加 IP 麻煩。

**風險**：K8s API endpoint port scan / brute force 可能性（但有 IAM auth 保護）。

### Node 直接無 public IP

需要 private cluster + Cloud NAT。Cloud NAT 多 $32/月，超預算。本次未做。

## 為什麼這些決定加總起來合理

- ✓ gRPC ports 全部 intra-cluster：firewall 自動擋
- ✓ SSH/RDP 刪除：減 attack surface
- ✓ Cloud Armor + Cloudflare allowlist：擋直接打 LB
- ◐ K8s master 仍對外開：靠 IAM auth 保護，可接受
- ◐ Node 有 public IP：靠 firewall 擋（已驗證）+ 沒對外開放 port

對學習 + $300 預算來說是**合理的安全姿態**。Production 還會加：
- Master Authorized Networks
- Private cluster + Cloud NAT
- Workload Identity 取代 GSA key
- Cloud Logging audit + alerting
- Pod Security Standards / OPA Gatekeeper
