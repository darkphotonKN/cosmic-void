---
topic: gke-deployment
subtopic: gke-cluster
date: 2026-05-15
extracted-to-vault: ""
---

# Phase 3 — 建立 GKE Cluster

## 3.1 完整指令 + 每個 flag 為什麼

```bash
gcloud container clusters create cosmic-void \
  --zone=us-central1-a \
  --num-nodes=3 \
  --machine-type=e2-medium \
  --disk-type=pd-balanced \
  --disk-size=20 \
  --enable-autoscaling --min-nodes=2 --max-nodes=4 \
  --enable-autorepair \
  --enable-autoupgrade \
  --release-channel=regular \
  --enable-ip-alias \
  --addons=HorizontalPodAutoscaling,HttpLoadBalancing \
  --logging=SYSTEM,WORKLOAD \
  --monitoring=SYSTEM
```

| Flag | 意義 | 選這個的理由 |
|---|---|---|
| `--zone=us-central1-a` | zonal cluster（單 zone） | regional 多 2× 個 master，會吃掉 $74.40 free tier credit |
| `--num-nodes=3` | 初始 3 nodes | 跑 8 service + middleware + frontend + ingress，最少 3 才夠 |
| `--machine-type=e2-medium` | 2 vCPU shared + 4 GB | 最便宜的 standard machine type；e2-small 太小 |
| `--disk-type=pd-balanced` | balanced disk（SSD-like 速度） | 比 pd-standard 快 5×，比 pd-ssd 便宜 2× |
| `--disk-size=20` | boot disk 20 GB | **預設 100 GB**，省 ~$24/月 |
| `--enable-autoscaling --min=2 --max=4` | 2-4 nodes 自動伸縮 | Pod scheduling pressure 時自動加 |
| `--enable-autorepair` | node 壞掉自動修 | 學習用無需顧 |
| `--enable-autoupgrade` | k8s 版本自動升級 | release-channel=regular 的階段性升級 |
| `--release-channel=regular` | 穩定但不太老 | rapid 太前沿、stable 太保守 |
| `--enable-ip-alias` | VPC-native cluster | **container-native LB 必須**，後面 Ingress NEG 模式 depends |
| `--addons=HttpLoadBalancing` | 內建 GCE Ingress controller | 我們要用 GCE Ingress |
| `--logging=SYSTEM,WORKLOAD` | system + 我們 pod 的 log 都送 Cloud Logging | 50 GiB/月免費 |
| `--monitoring=SYSTEM` | system metrics 送 Cloud Monitoring | 免費 |

## 3.2 過程

```
Creating cluster cosmic-void in us-central1-a...
...........(約 5-8 分鐘).....done.

NAME         LOCATION       MASTER_VERSION      MASTER_IP    MACHINE_TYPE  NODE_VERSION        NUM_NODES  STATUS
cosmic-void  us-central1-a  1.35.3-gke.1389000  34.71.54.60  e2-medium     1.35.3-gke.1389000  3          RUNNING
```

Master IP 34.71.54.60 — 這是 GKE 控制面的 public endpoint。

## 3.3 抓 kubeconfig

```bash
gcloud container clusters get-credentials cosmic-void --zone=us-central1-a
# 自動更新 ~/.kube/config，添加 cosmic-void context

kubectl get nodes  # 看到 3 個 Ready
```

## 3.4 觀察：autoscaler 啟動

deploy 過程中發現 cluster 短暫升到 4 node。原因：

- 8 service rollout 同時發生
- 舊+新 ReplicaSet 並存（maxSurge:1 期間）
- 短期 11+ pods 集中要排在 3 個 node
- autoscaler 偵測 Pending → 加 1 node

之後 pod 都穩定後可以縮回 3。

## 3.5 為什麼選 zonal 不是 regional：free tier 計算實證

| 模式 | Master 數量 | 收費 | Free tier 抵免 |
|---|---|---|---|
| zonal | 1（單 zone） | $0.10/cluster/hr | ✅ $74.40/月 抵掉 |
| regional | 3（跨 zone） | 同 $0.10/cluster/hr | ❌ free tier 只給 zonal 抵 |
| Autopilot | 1 controller | $0.10/hr | ✅ 但 control plane 加 $74.40 不抵 |

3 × $74.40 = $223.20 — regional 一個月就燒掉 2/3 額度。學習用沒必要。

## 3.6 GKE control plane 安全姿態

|  | 預設 | 強化選項 |
|---|---|---|
| Control plane IP | public（34.71.54.60）任何 IP 都能 reach :443 | Master Authorized Networks（限定 IP 進） |
| Auth 機制 | OAuth token via gcloud | Workload Identity / cert |
| 我們選擇 | 預設（保留 0.0.0.0/0） | user 決定不加 Authorized Networks |

不加 Authorized Networks 的代價：attacker 能 port scan 34.71.54.60:443，但要進 cluster 仍需要有效 GCP IAM credential（OAuth token），暴力破解不可行。

## 3.7 後續 cost 優化

當所有東西穩定後，可降回 max-nodes=3 省約 $24/月：

```bash
gcloud container clusters update cosmic-void --zone=us-central1-a \
  --enable-autoscaling --min-nodes=2 --max-nodes=3
```
