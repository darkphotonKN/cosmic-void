---
title: "kubectl 指令速查表（按情境分類）"
type: learning-note
project: cosmic-void
topic: k8s
date: 2026-04-30
status: extracted
extracted-to-vault:
  - "[[Kubectl Cheatsheet]]"
  - "[[K8s Apply vs Runtime]]"
extracted-at: 2026-04-30
archive-mirror: vault/創作庫/projects/cosmic-void/learning-archive/k8s/kubectl-cheatsheet.md
related-files:
  - game-server/auth-service/k8s/deployment.yml
  - game-server/auth-service/k8s/service.yml
  - game-server/auth-service/k8s/configmap.yml
tags: [kubernetes, k8s, kubectl, cheatsheet, debug, reference]
note-type: reference
---

## 學習目標

把學 K8s Deployment / Service / ConfigMap 過程中接觸到的 kubectl 指令，按「**實際情境**」分類成隨時可查的速查表。重點不是背指令，而是「**遇到 X 問題就知道用哪個指令**」。

## 對話脈絡

> Q: 希望能把常用的 kubectl 情境指令列給我
> A: 整理出 10 個常見情境 + Pod / Service debug 標準流程 + 偷懶 alias 設定。

（這份是**參考工具型筆記**，不是概念學習筆記，所以對話脈絡較少。指令本身的**「為什麼用」「什麼時機」**才是價值所在。）

## 關鍵理解

### 1. kubectl 三大動詞

```
get      → 列表（看全貌）
describe → 細節（看 events、配置、狀態）
logs     → 程式 log（看 stdout/stderr）
```

90% 的 debug 從這三個開始。

### 2. 「apply 不檢查 runtime」原則

```
apply 階段：只檢查 YAML 語法
runtime 階段：實際能不能通要看：
  - kubectl get endpoints   ← Service 對到 Pod 嗎
  - kubectl logs            ← 程式有沒有正常啟動
  - kubectl exec -- env     ← 環境變數正確注入嗎
  - kubectl exec -- netstat ← 程式真的在監聽 port 嗎
```

寫完 YAML `kubectl apply` 成功只是第一關，runtime 才是真相。

### 3. 標準 debug 流程

| 症狀 | 第一個跑的指令 |
|---|---|
| Pod 起不來 | `kubectl describe pod xxx`（看 Events）|
| Pod 起來了但行為不對 | `kubectl logs xxx`（看程式 log）|
| 服務連不上 | `kubectl get endpoints xxx`（看 Service 對到 Pod 沒）|
| DNS 解析失敗 | `kubectl run --image=busybox:1.28 -- nslookup xxx` |
| 環境變數沒生效 | `kubectl exec xxx -- env \| grep KEY` |

## 程式碼 / 設定

### 情境 1：「我想知道 cluster 裡有什麼」

```bash
# Namespace
kubectl get namespaces                    # 簡寫：kubectl get ns

# 全覽（粗略）
kubectl get all                           # 看當前 namespace
kubectl get all -A                        # 所有 namespace（-A = --all-namespaces）

# 特定種類
kubectl get pods                          # = kubectl get po
kubectl get deployments                   # = kubectl get deploy
kubectl get services                      # = kubectl get svc
kubectl get configmaps                    # = kubectl get cm
kubectl get secrets

# 看更多細節（IP、Node 等）
kubectl get pods -o wide
```

### 情境 2：「我想看某個東西的詳細狀態」

```bash
# describe = 看完整細節（events、設定、狀態）
kubectl describe pod auth-service-xxx-yyy
kubectl describe deployment auth-service
kubectl describe service auth-service
kubectl describe configmap auth-service-config

# 看 YAML 原貌
kubectl get pod auth-service-xxx -o yaml
kubectl get svc auth-service -o yaml
```

⭐ **debug 第一招**：`describe pod xxx` 底下的 **Events** 區塊通常告訴你「為什麼 Pod 起不來」。

### 情境 3：「我想部署 / 更新 / 刪除東西」

```bash
# 套用 YAML（建立或更新都用這個）
kubectl apply -f deployment.yml
kubectl apply -f k8s/                       # 套用整個資料夾的 YAML

# 刪除
kubectl delete -f deployment.yml            # 用 YAML 刪
kubectl delete pod auth-service-xxx         # 直接指名刪
kubectl delete deployment auth-service      # 刪整個 Deployment
kubectl delete service auth-service

# Dry run（先驗證 YAML 不真的跑）
kubectl apply -f deployment.yml --dry-run=client
kubectl apply -f deployment.yml --dry-run=server   # 比較嚴格，會送到 API server 但不寫入

# 重啟 Deployment 的所有 Pod（不改 YAML）
kubectl rollout restart deployment auth-service
```

⭐ **改了 ConfigMap 但 Pod 沒拿到新值**？用 `rollout restart` 強制重建 Pod（envFrom 不會自動 reload）。

### 情境 4：「我的服務壞了，要查看程式 log」

```bash
# 看 Pod 的 log
kubectl logs auth-service-xxx-yyy

# 看「即時」log（像 tail -f）
kubectl logs -f auth-service-xxx-yyy

# Pod 有多 container 時指定
kubectl logs auth-service-xxx -c container-name

# 看「上次崩潰前」的 log（Pod 重啟過時很有用）
kubectl logs auth-service-xxx --previous

# 看最近 N 行
kubectl logs auth-service-xxx --tail=100

# 看過去 N 分鐘
kubectl logs auth-service-xxx --since=10m
```

⭐ **CrashLoopBackOff 看不到崩潰原因**？用 `--previous` 看上一次的 log。

### 情境 5：「我要進到 Pod 裡面看」

```bash
# 進 Pod 開 shell（最常用）
kubectl exec -it auth-service-xxx-yyy -- /bin/bash
kubectl exec -it auth-service-xxx-yyy -- /bin/sh   # 沒 bash 時用 sh

# 不進去，直接執行單一指令
kubectl exec auth-service-xxx-yyy -- env           # 看環境變數
kubectl exec auth-service-xxx-yyy -- ls /app
kubectl exec auth-service-xxx-yyy -- cat /etc/hosts
kubectl exec auth-service-xxx-yyy -- netstat -tln  # 看 LISTEN 的 ports

# 跑一個臨時 Pod 來 debug（用完自動刪）
kubectl run -it --rm debug --image=busybox:1.28 --restart=Never -- sh
```

⭐ **「為什麼程式拿不到環境變數？」** → `kubectl exec pod -- env | grep DB_` 直接看實際的環境變數。

### 情境 6：「網路問題，Service / DNS 怎麼 debug」

```bash
# 1. Service 存不存在？
kubectl get svc auth-service

# 2. ⭐ Service 有沒有對到 Pod？（debug 神器）
kubectl get endpoints auth-service
# → 有 IP：selector 對 ✅
# → <none>：selector 對不上 ⚠️

# 3. 從 cluster 內測 DNS 解析
kubectl run -it --rm dns-test --image=busybox:1.28 --restart=Never -- nslookup auth-service

# 4. 從 cluster 內測連線
kubectl run -it --rm curl-test --image=curlimages/curl --restart=Never -- curl http://auth-service:7003

# 5. 直接從某個 Pod 內試
kubectl exec -it some-pod -- nslookup auth-service
kubectl exec -it some-pod -- wget -O- http://auth-service:7003
```

⭐ **記三步走**：`get svc` → `get endpoints` → `nslookup`。

⭐ **busybox 用 1.28 不要用 latest**：最新版的 nslookup 在某些 K8s 環境有 bug。

### 情境 7：「我要看 ConfigMap / Secret 的內容」

```bash
# ConfigMap
kubectl get configmap auth-service-config -o yaml
kubectl describe configmap auth-service-config

# Secret（值會是 base64 編碼）
kubectl get secret auth-service-secrets -o yaml

# 解碼 Secret 的某個 key（看真實值）
kubectl get secret auth-service-secrets -o jsonpath='{.data.DB_PASSWORD}' | base64 -d

# 從指令直接建 ConfigMap（不寫 YAML）
kubectl create configmap myconfig --from-literal=KEY=value
kubectl create configmap myconfig --from-file=app.properties

# 從指令建 Secret
kubectl create secret generic mysecret --from-literal=password=mypw
```

⚠️ **資安**：`-o yaml` 會印出 Secret 內容（即使是 base64），分享 terminal 時注意。

### 情境 8：「我要監控資源使用量」

```bash
# 看 Pod 的 CPU / 記憶體（需要 metrics-server）
kubectl top pods
kubectl top pods -A
kubectl top nodes

# 看 Pod 的詳細狀態變化
kubectl get pods -w                       # -w = watch，持續更新

# 看 Pod 在哪個 Node
kubectl get pods -o wide
```

⭐ **部署時開另一個 terminal**：`kubectl get pods -w` 看 Pod 從 Pending → ContainerCreating → Running 的過程。

### 情境 9：「Rolling Update 相關」

```bash
# 看 Deployment 的 rollout 狀態
kubectl rollout status deployment auth-service

# 看 Deployment 歷史版本
kubectl rollout history deployment auth-service

# 看某個版本的細節
kubectl rollout history deployment auth-service --revision=2

# 回滾到上一版
kubectl rollout undo deployment auth-service

# 回滾到特定版本
kubectl rollout undo deployment auth-service --to-revision=2

# 暫停 rollout（先停下來，準備改設定）
kubectl rollout pause deployment auth-service
kubectl rollout resume deployment auth-service
```

⭐ **生產環境保命招**：發布壞了 → `kubectl rollout undo` 立刻回滾。

### 情境 10：「快速操作 / 偷懶寫法」

```bash
# 用 label 批次操作
kubectl get pods -l app=cosmic-void
kubectl delete pods -l component=auth-service

# 強制刪除 Pod（卡住時）
kubectl delete pod xxx --grace-period=0 --force

# 改 replicas 不用編輯 YAML
kubectl scale deployment auth-service --replicas=3

# 即時編輯資源（會跳開編輯器）
kubectl edit deployment auth-service       # 改完存檔自動 apply
kubectl edit svc auth-service

# Port forward（從本機連到 cluster 內服務）
kubectl port-forward svc/auth-service 7003:7003
# 然後本機 localhost:7003 就能連到 Service

# Port forward 到 Pod
kubectl port-forward pod/auth-service-xxx 7003:7003
```

⭐ **超實用**：`kubectl port-forward svc/postgres 5432:5432`，本機 DBeaver 就能連 K8s 內的 Postgres，不用 NodePort。

### 情境 11：Pod 起不來的標準 debug 流程

```bash
# Step 1：看狀態
kubectl get pods
# READY 0/1 + STATUS 是什麼？
# - Pending：排不到 Node（資源不足）
# - ContainerCreating：拉 image 中
# - CrashLoopBackOff：起來又死
# - ImagePullBackOff：image 拉不到
# - Error：直接掛

# Step 2：看細節（events 區塊最有用）
kubectl describe pod auth-service-xxx
# → 底部 "Events:" 通常告訴你原因

# Step 3：看程式 log
kubectl logs auth-service-xxx
kubectl logs auth-service-xxx --previous   # 已崩潰時用這個

# Step 4：進去看
kubectl exec -it auth-service-xxx -- sh
kubectl exec auth-service-xxx -- env       # 看環境變數對不對
kubectl exec auth-service-xxx -- netstat -tln  # 看 port 監聽
```

### 情境 12：服務連不上的標準 debug 流程

```bash
# Step 1：Service 存在？
kubectl get svc auth-service

# Step 2：⭐ Service 對到 Pod？
kubectl get endpoints auth-service
# → 沒 IP：selector 對不上 / Pod 沒通過 readinessProbe

# Step 3：DNS 通？
kubectl run -it --rm dns-test --image=busybox:1.28 --restart=Never -- nslookup auth-service

# Step 4：連線通？
kubectl run -it --rm test --image=curlimages/curl --restart=Never -- curl http://auth-service:7003

# Step 5：Pod 真的開了那個 port 嗎？
kubectl exec auth-service-xxx -- netstat -tln
```

### 偷懶 alias 設定（建議加進 ~/.zshrc）

```bash
alias k=kubectl
alias kgp='kubectl get pods'
alias kgs='kubectl get svc'
alias kgd='kubectl get deployments'
alias kdp='kubectl describe pod'
alias kl='kubectl logs'
alias klf='kubectl logs -f'
alias kex='kubectl exec -it'
```

之後可以：

```bash
k get pods                    # 比 kubectl get pods 省一半字
kgp -A                        # 列所有 namespace 的 pods
klf auth-service-xxx          # 即時看 log
```

## 踩過的坑

- 問題：把「Could not resolve host」debug 用 `kubectl get endpoints`
  解法：先用 `kubectl run -- nslookup` 確認 DNS 解析
  為什麼：「Could not resolve」是 DNS 階段失敗，連 IP 都沒拿到，endpoints 看再多都沒用。要分清楚「DNS 失敗」vs「連線失敗」。

- 問題：以為 `kubectl apply` 成功就 = 服務能通
  解法：apply 後一定要驗證 runtime（get endpoints / logs / exec）
  為什麼：K8s 是宣告式系統，apply 只接受你的宣告，不檢查實際能不能跑。「會 apply」≠「會通」。

- 問題：Pod 在 CrashLoopBackOff 看不到崩潰原因
  解法：用 `kubectl logs xxx --previous` 看上一次的 log
  為什麼：當前 Pod 還沒啟動就崩潰了，logs 是空的。`--previous` 拿前一個 container 實例的 log。

- 問題：改了 ConfigMap 後 Pod 沒拿到新值
  解法：`kubectl rollout restart deployment xxx` 強制重建 Pod
  為什麼：`envFrom` 在 Pod 啟動時注入環境變數，**不會自動 reload**。要重新建 Pod 才會讀到新值。
  進階：要自動 reload 需要 [Reloader](https://github.com/stakater/Reloader) 之類的工具。

## 待釐清

- [ ] kubectl context / config 多 cluster 切換（之後實際部署到雲端會用到）
- [ ] kubectl plugin（krew）是什麼？常用 plugin 有哪些？
- [ ] kustomize 跟 helm 的 kubectl 整合用法
- [ ] kubectl debug（K8s 1.20+ 新功能）vs 傳統 kubectl exec

## 相關專案檔案

- `game-server/auth-service/k8s/deployment.yml`
- `game-server/auth-service/k8s/service.yml`
- `game-server/auth-service/k8s/configmap.yml`
- `game-server/auth-service/k8s/secret.yml.example`

## 相關 learning notes

- [[deployment]] — 學 Deployment 時用到的 kubectl 指令
- [[auth-service-deployment-practice]] — 實戰 debug 經驗
- [[service]] — 學 Service 時的 endpoints / nslookup debug
