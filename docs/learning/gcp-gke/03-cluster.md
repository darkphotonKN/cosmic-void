---
date: 2026-05-15
topic: gcp-gke
subtopic: 03-cluster
extracted-to-vault:
---

# 03 — GKE Standard zonal cluster + IAM

## 為什麼 zonal 不 regional

- **GKE Free Tier = $74.40/月 抵免**，**只給 zonal Standard cluster** 或 Autopilot 用
- regional cluster 直接收 $73/月 不抵 → 90 天破預算

```bash
ZONE="us-central1-a"   # 注意是 zone 不是 region

gcloud container clusters create cosmic-void \
  --zone="$ZONE" \
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

### 每個 flag 為什麼

| Flag | 為什麼 |
|---|---|
| `--zone us-central1-a` | Zonal 才享 $74.40/月 free tier |
| `--num-nodes=3` | 8 service + middleware + ingress 估算 ~1.3 vCPU、~3.5GB requests，3 nodes 配 e2-medium 充裕 |
| `--machine-type=e2-medium` | 2 vCPU / 4GB，shared-core，$24.46/月，學習用足夠 |
| `--disk-type=pd-balanced` | 比 pd-standard 略快但便宜，PVC 默認也用這個 |
| `--disk-size=20` | **省 $24/月**。預設 100GB × 3 nodes 多收 $24/月，學習用 20GB 足夠 |
| `--enable-autoscaling --min-nodes=2 --max-nodes=4` | autoscale 區間。max=4 是上限，超過會 Pending |
| `--enable-autorepair --enable-autoupgrade` | GKE 自動處理節點維運 |
| `--release-channel=regular` | 自動拿穩定的 master 升級 |
| `--enable-ip-alias` | VPC-native cluster，**NEG mode 必要**（container-native LB）|
| `--addons=...HttpLoadBalancing` | 安裝 GCE Ingress controller，預設 Ingress 用 GCP L7 LB |

完成後：
```bash
gcloud container clusters get-credentials cosmic-void --zone="$ZONE"
kubectl get nodes  # 3 Ready
```

時間：建 cluster 約 5-8 分鐘。

## 為什麼 max-nodes 後來考慮降回 3

autoscaler 在 `rolling restart` 時為了讓新 pod 上線而 surge 一個節點到 4，但用完不縮回去（縮 timer 是 10 分鐘）。學習用流量小，3 nodes 永遠夠 → 待辦：`--max-nodes=3`。

## ⚠ 新版 GCP 隱形雷：Default Compute SA 沒有 `roles/editor`

部署 8 個 service 後，pod 全部 `ImagePullBackOff`：
```
failed to pull and unpack image "...": failed to authorize:
failed to fetch oauth token: ... 403 Forbidden
```

**根因**：GKE 節點 default 用 `<PROJECT_NUMBER>-compute@developer.gserviceaccount.com` SA。舊版 GCP 這個 SA 自動有 `roles/editor`（涵蓋 artifactregistry.reader）。**新版 GCP（2024+）取消這個自動授權**，新 project 的 default Compute SA 完全沒任何 role。

驗證：
```bash
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
DEFAULT_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

gcloud projects get-iam-policy "$PROJECT_ID" \
  --flatten="bindings[].members" \
  --filter="bindings.members:$DEFAULT_SA" \
  --format="value(bindings.role)"
# 空 = 沒有任何 role
```

**修法**：明確賦予 Artifact Registry reader：
```bash
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$DEFAULT_SA" \
  --role="roles/artifactregistry.reader" \
  --condition=None
```

加完後 `kubectl rollout restart deployment <name>` 重拉 image 就會通。

## 為什麼 cluster autoscaler 會把 pods 集中在 1 個 node

第一次部署觀察到：8 個 service 中 5 個 Pending，剩 3 個在同一個 node 跑（雖然有 3 個 node）。

**原因**：
1. 第一個 pod 上 node A，autoscaler 看其他 node 不需要 → 不擴
2. 第二個 pod 上 node A（pack 到既有節點），同理
3. 等 node A 沒 CPU 才開始排隊
4. 後面 pod Pending → autoscaler 才 surge 第 4 個 node
5. 但 max-nodes=4 撞頂

整個過程：image pull fail 期間，pod 一直 retry，scheduler 在等 image 成功；同時 autoscaler 看待排程 pod 數量決定擴/縮。混亂時 pod 集中在某個 node 是 expected behavior。

**強化建議**：給每個 service 加 `topologySpreadConstraints` 強制散到不同節點。本次未做（學習階段可接受集中）。

## kubeconfig context 切換

```bash
kubectl config set-context --current --namespace=cosmic-void
# 之後 kubectl 不用每次 -n cosmic-void
```

## 驗證

```bash
kubectl get nodes -o wide                                       # 3 Ready
gcloud container clusters describe cosmic-void --zone=$ZONE \
  --format="value(currentMasterVersion,locations)"
```

## 為什麼這個選擇 vs 其他

| 替代 | 為何沒選 |
|---|---|
| Autopilot | $74.40/月控制平面 + per-pod billing，估算超預算 |
| Regional cluster | 不抵 free tier，多 $73/月 |
| Private cluster（節點無 public IP）| 需要 Cloud NAT 處理 egress = +$32/月，邊緣超預算 |
| e2-small | 2 vCPU shared / 2GB RAM，太小排不下 13 個 pod |
| e2-standard-2 | 8GB RAM 有餘但 $48.92/月 × 3 = $147，預算撐不到 90 天 |
| **e2-medium × 3** | 甜蜜點 |
