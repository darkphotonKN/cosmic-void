---
topic: gke-deployment
subtopic: firewall-cloud-armor
date: 2026-05-15
extracted-to-vault: ""
---

# 安全強化：VPC firewall vs Cloud Armor

## User 的核心疑問

> 「為什麼 VPC firewall 無法擋公網直接打 34.111.74.79？」
> 「Cloud Armor 不就是 AWS WAF？GCP 有沒有更便宜的等價物？」

## GCP 三層網路面結構

```
[Layer A] Internet (任何 IP)
   ↓
[Layer B] Google Edge POPs (全球 anycast)
          └─ L7 HTTPS LB frontend (34.111.74.79)
   ↓
[Layer C] VPC (你的 project)
          └─ Compute instances / GKE pods
```

**VPC firewall rules 只能套在 Layer C 的 VM 介面**。它**完全看不到 Layer B 的 LB frontend traffic**。

## 為什麼 VPC firewall 對 L7 LB 無效

當 user `curl http://34.111.74.79`：

1. 封包進 **Layer B Google edge**（你管不到的層）
2. LB 在 B 層**解 TLS、看 HTTP host header、跑 URL Map**
3. LB **重新打包**封包用 Google internal network 傳到你的 VPC NEG
4. VPC firewall 在 Layer C **看到的源 IP 是 Google internal IP**，不是攻擊者的 IP
5. 攻擊者的真實 IP **塞在 HTTP header `X-Forwarded-For`**，firewall 不解析 header

**結論：VPC firewall 看 L3/L4，L7 LB 是 proxy 不是 passthrough，所以 firewall 看不到原始源 IP。**

## 路徑 1：Cloud Armor（採用）

Cloud Armor = **GCP 版 AWS WAF**。掛在 L7 LB 的 backend service，可以：
- IP 允許/拒絕清單（包含 CIDR）
- Geo 阻擋
- Rate limiting
- 預設 OWASP 規則（SQL injection、XSS）
- 自訂 expression language

### 價格對比

| 項目 | AWS WAF | GCP Cloud Armor |
|---|---|---|
| Web ACL / Policy | $1/月 | $5/月 |
| 每條規則 | $1/月 | $1/月 |
| 每百萬 requests | $0.60 | $0.75 |

我們的設計：1 policy + 1 rule（Cloudflare CIDR allowlist）= **$6-7/月**。

### 部署計畫

```bash
# 1. 拿最新 Cloudflare IPv4 CIDR
curl -s https://www.cloudflare.com/ips-v4 > /tmp/cf-ipv4.txt
CF_IPS=$(paste -sd, /tmp/cf-ipv4.txt)

# 2. 建 security policy
gcloud compute security-policies create cosmic-void-cf-only \
  --description="Allow only Cloudflare IPs"

# 3. allow Cloudflare（priority 1000）
gcloud compute security-policies rules create 1000 \
  --security-policy=cosmic-void-cf-only \
  --src-ip-ranges=$CF_IPS \
  --action=allow

# 4. 預設 deny（priority 2147483647 內建）
# 預設規則就是 deny-all，不用再設

# 5. 掛到每個 backend service
for bs in $(gcloud compute backend-services list --global --format="value(name)" | grep cosmic-void); do
  gcloud compute backend-services update "$bs" --global \
    --security-policy=cosmic-void-cf-only
done

# 6. 驗證
curl http://34.111.74.79/                        # → 403 (被擋)
curl https://cosmicvoid.uk/                      # → 200 (CF 進來 OK)
```

## 路徑 2：Application-layer 過濾（免費替代）

如果不要付 Cloud Armor 的 $6/月，可以在 api-gateway 加 middleware：

```go
import "net"

var cloudflareCIDRs = []*net.IPNet{ /* parse from CF IP list */ }

func cloudflareOnly(c *gin.Context) {
    xff := c.Request.Header.Get("X-Forwarded-For")
    realIP := net.ParseIP(strings.Split(xff, ",")[0])
    for _, cidr := range cloudflareCIDRs {
        if cidr.Contains(realIP) {
            c.Next()
            return
        }
    }
    c.AbortWithStatus(403)
}

router.Use(cloudflareOnly)
```

**缺點**：
- attacker 的 TCP 還是會打到 Pod（吃 CPU、networking quota）
- 每個 service 都要寫一份（或集中在 gateway）
- DDoS 防護 = 0

## 路徑 3：L4 LB + in-cluster ingress（免費但要重做架構）

L4 Network LB 是 passthrough（不解 TLS、不重打包），VPC firewall **看得到** 原始源 IP。

但：
- 失去 GCE Ingress 的 URL map / path routing（一個 LB 對一個 Service）
- 要把 TLS 放回 cluster（裝 ingress-nginx + cert-manager）
- 多 Service 路由要靠 in-cluster nginx ingress

**等同把已經做好的架構整個重來**。

## VPC firewall 還是有它的地方：node-level

VPC firewall 在 Layer C 仍然有效，擋住：
- Cluster node 的 SSH（tcp:22）
- Cluster node 之間的非預期流量
- pod-to-pod 不該允許的端口

### 我們的 VPC firewall 清理結果

```
| Rule                                | Source            | Allow            | 狀態 |
| default-allow-icmp                  | 0.0.0.0/0         | icmp             | 保留 |
| default-allow-internal              | 10.128.0.0/9      | tcp/udp/icmp     | 保留 |
| default-allow-rdp                   | 0.0.0.0/0         | tcp:3389         | 已刪 ✓ |
| default-allow-ssh                   | 0.0.0.0/0         | tcp:22           | 已刪 ✓（更嚴格）|
| gke-cosmic-void-2afe27ea-all        | 10.28.0.0/14      | all              | 保留（cluster pod-to-pod） |
| gke-cosmic-void-2afe27ea-exkubelet  | 0.0.0.0/0         | (deny)           | 保留（擋外部 kubelet） |
| gke-cosmic-void-2afe27ea-inkubelet  | 10.28.0.0/14      | tcp:10255        | 保留 |
| gke-cosmic-void-2afe27ea-vms        | 10.128.0.0/9      | all              | 保留（node-to-node） |
| k8s-fw-l7--89c4df675995becd         | 35.191/130.211    | tcp:7001,8080    | 保留（GKE 自動的 LB health check） |
```

### gRPC 流量檢查（user 特別問）

```bash
# 所有 gRPC port: 7003 7004 7010 7011 7013 7021 7077 + game HTTP 5555
# 對外 0.0.0.0/0 → gRPC port 的 firewall：應該無
gcloud compute firewall-rules list --filter="sourceRanges:0.0.0.0/0" \
  --format="value(name,allowed)" | grep -E "7003|7004|7010|7011|7013|7021|7077|5555"
# (應該為空)
```

gRPC 全部在 `gke-cosmic-void-2afe27ea-all`（10.28.0.0/14 → node）內部解決，不暴露對外。✓

## User 選了哪些強化

| 項目 | 狀態 | 理由 |
|---|---|---|
| Cloud Armor + CF IP allowlist | ✓ approve | 願意花 $6/月買 production-grade 保護 |
| Master Authorized Networks | ✗ 不做 | 怕換 IP 後 kubectl 斷，依賴 IAM auth |
| `default-allow-ssh` source 限縮 | ✓ 整條刪 | 用 kubectl exec 取代 SSH |
| `default-allow-rdp` 刪除 | ✓ 已刪 | Linux node 用不到 |

## 教訓

1. **VPC firewall 與 Cloud Armor 在不同 layer 工作**，不能互相取代
2. **L7 LB 是 proxy，源 IP 在 header 不在封包**，這決定了所有 IP 過濾架構
3. **Cloud Armor === GCP 版 AWS WAF**，是唯一在 L7 LB frontend 過濾 IP 的方案
4. **要省 $6/月可以 app-layer 過濾，但 attacker TCP 仍進 Pod**
5. **不必要的對外規則（RDP/SSH 0.0.0.0/0）刪掉一條少一條 attack surface**
