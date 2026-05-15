# Kubernetes 學習路徑

> 開始日期：2026-04-27
> 狀態：learning

## 已寫的筆記

- [[deployment]] — 2026-04-30：完整學會 deployment.yml 從頭到尾（YAML 結構 → metadata → labels/selector → strategy → containers → probes / env / volumes）
- [[auth-service-deployment-practice]] — 2026-04-30：第一次實戰寫 K8s YAML（deployment + configmap + secret），含 envFrom 用法、K8s vs docker-compose port 差異、.gitignore 保護真實 secret
- [[service]] — 2026-04-30：Service 入門 + 實戰寫 auth-service service.yml（四種 type、port 三層、DNS / FQDN、子集匹配規則、apply vs runtime）
- [[kubectl-cheatsheet]] — 2026-04-30：kubectl 指令速查表（12 種情境分類 + Pod/Service debug 流程 + alias 設定）
- [[configmap-and-secret]] — 2026-05-08：ConfigMap + Secret 深入（env 進階用法、volume mount + subPath、reload 機制、Secret 4 種 type、base64 真相）
- [[ingress]] — 2026-05-08：Ingress 入門 + 實戰寫 cosmic-void HTTPS 路由（Ingress vs Controller、L7 vs L4、pathType 路徑層級匹配、TLS termination、cert-manager + Let's Encrypt、整套 K8s 物件契約對齊）

## 進行中的實作計劃

- [[_plan-deploy-auth-service]] — 2026-05-08 起：本機 K8s → EKS 四階段部署
  - 範圍：auth-service + api-gateway
  - 狀態：Phase 1 進行中（本機 Docker Desktop K8s 跑通）
  - 完成 phase 後在這裡同步打勾：Phase 1 ⬜ / Phase 2 ⬜ / Phase 3 ⬜ / Phase 4 ⬜

## 學習目標

短期（這個月）：
- [x] 看懂任意 K8s deployment.yml 每一個欄位
- [x] 為 cosmic-void 的某個微服務（auth-service 優先）寫一份完整 deployment.yml ✅ 2026-04-30
- [x] 學會 Service（讓 Pod 可被連線）✅ 2026-04-30
- [x] 學會 ConfigMap + Secret（設定外部化）✅ 2026-04-30
- [x] kubectl 速查表整理 ✅ 2026-04-30
- [x] 學會 Ingress（HTTP 路由）✅ 2026-05-08

中期（之後幾個月）：
- [ ] PersistentVolume / PersistentVolumeClaim（持久化儲存）
- [ ] HPA（自動擴縮）
- [ ] Helm（多份 YAML 模板化）

## 學習風格偏好

- 一個一個 YAML 欄位拆開講解（不要一次倒一堆）
- 每段結尾用三個小問題驗證理解
- 例子要連結到 cosmic-void 真實架構（auth-service / notification-service / outbox 等）

## 已萃取進 vault 的概念

### 第一輪：由 [[deployment]] 萃取於 2026-04-30
- `wiki/entities/cloud/Kubernetes Deployment.md`
- `wiki/entities/cloud/K8s RollingUpdate Strategy.md`
- `wiki/entities/cloud/K8s Labels and Selector.md`
- `wiki/entities/cloud/K8s Liveness vs Readiness Probe.md`
- `wiki/entities/cloud/K8s Resources Requests Limits.md`
- `wiki/entities/cloud/K8s Pod Volume.md`

Source 摘要：`wiki/sources/source-cosmic-void-k8s-deployment-2026-04-30.md`

### 第二輪：由 [[auth-service-deployment-practice]] 萃取於 2026-04-30
新建：
- `wiki/entities/cloud/K8s ConfigMap.md`
- `wiki/entities/cloud/K8s Secret.md`
- `wiki/entities/cloud/K8s envFrom vs env.md`
- `wiki/entities/cloud/Docker Image vs Container vs Compose.md`
- `wiki/entities/cloud/K8s Service Discovery.md`（stub，待 Service 章節完成補完）
- `wiki/concepts/Gitignore Patterns.md`

補充：
- `wiki/entities/cloud/K8s Labels and Selector.md`（多重 labels 模式）
- `wiki/entities/cloud/K8s Liveness vs Readiness Probe.md`（tcpSocket vs httpGet 實戰）

Source 摘要：`wiki/sources/source-cosmic-void-k8s-auth-service-2026-04-30.md`

### 第三輪：由 [[service]] + [[kubectl-cheatsheet]] 萃取於 2026-04-30
新建：
- `wiki/entities/cloud/Kubernetes Service.md`
- `wiki/entities/cloud/K8s Service Port Layers.md`
- `wiki/entities/cloud/Kubectl Cheatsheet.md`
- `wiki/concepts/K8s Apply vs Runtime.md`

升級（從 stub 到 stable）：
- `wiki/entities/cloud/K8s Service Discovery.md`（CoreDNS 完整機制、FQDN、search domain、debug）

補充：
- `wiki/entities/cloud/K8s Labels and Selector.md`（Service selector 子集匹配風險、Service vs Deployment 語法差異）

Source 摘要：`wiki/sources/source-cosmic-void-k8s-service-2026-04-30.md`

歸檔位置：vault/創作庫/projects/cosmic-void/learning-archive/k8s/
