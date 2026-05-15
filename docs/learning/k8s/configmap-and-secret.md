---
title: "K8s ConfigMap + Secret 深入：env / volume mount / reload / Secret 進階"
type: learning-note
project: cosmic-void
topic: k8s
date: 2026-05-08
status: learning
extracted-to-vault: []
related-files:
  - game-server/auth-service/k8s/configmap.yml
  - game-server/auth-service/k8s/secret.yml.example
  - game-server/auth-service/k8s/deployment.yml
tags: [kubernetes, k8s, configmap, secret, env, envfrom, volumemount, subpath, reload, rolling-update, tls, base64, external-secrets]
---

## 學習目標

延續之前實戰寫的 ConfigMap + Secret，深入學「**沒寫過但會碰到**」的場景：env 進階用法、把設定當檔案掛、reload 機制、Secret 進階 type、安全性真相。重點是搞清楚「**改了 ConfigMap/Secret 後，Pod 到底會不會拿到新值**」這個經典踩雷點。

## 對話脈絡

> Q: env 跟 envFrom 可以同時用嗎？衝突誰贏？
> A: 可以同時用，**`env` 比 `envFrom` 優先**。實戰價值：用 envFrom 整批匯入 ConfigMap，再用 env 局部覆寫某個 key（不影響其他 Pod）。記法：「整批先、例外後」。

> Q: 我的 ConfigMap 有 DB_HOST 但程式讀 POSTGRES_HOST，怎麼 rename？
> A: 用 `env.valueFrom.configMapKeyRef`：
>   - `env[].name` = 容器內看到的名字（POSTGRES_HOST）
>   - `configMapKeyRef.key` = ConfigMap 內原始的 key（DB_HOST）
>   兩者不同 = rename。

> Q: 兩個 ConfigMap 都有 DB_HOST 撞名怎麼辦？
> A: `envFrom` 加 `prefix:`，整批匯入時加前綴：
>   ```yaml
>   envFrom:
>   - configMapRef: { name: auth-db-config }
>     prefix: AUTH_      # → AUTH_DB_HOST
>   - configMapRef: { name: audit-db-config }
>     prefix: AUDIT_     # → AUDIT_DB_HOST
>   ```

> Q: optional: true 是什麼？什麼時候用？
> A: 標記 ConfigMap/Secret 是「可選的」，不存在時 Pod 仍會啟動。
> ⚠️ 危險：核心設定（DB / 認證）絕對不要用，否則會出現「Pod 看起來健康但業務全壞」的隱性失敗。只用在「真的可有可無」的設定（feature flags / debug 開關）。

> Q: 把 ConfigMap 當「檔案」掛進 Pod 怎麼用？
> A: ConfigMap 的 key 變成「檔案名」，value（用 `|` 多行字串）變成「檔案內容」。透過 `volumes: configMap` + `volumeMounts: mountPath` 掛到容器內路徑。
>   ```yaml
>   data:
>     nginx.conf: |
>       server { ... }
>   ```
>   容器內：/etc/nginx/conf.d/nginx.conf 是這個內容。

> Q: subPath 的差別？
> A:
>   - **不用 subPath**：mountPath 那個資料夾**整個被取代**，原本容器內的檔案全消失（經典踩雷：nginx 的 mime.types 被覆蓋掉而起不來）
>   - **用 subPath**：只放單一檔案進去，其他原本的檔案還在
>   99% 的設定檔掛載都應該用 subPath。

> Q: env vs volume mount 怎麼選？
> A:
>   - 簡單 key-value（DB_HOST、PORT、LOG_LEVEL）→ envFrom
>   - 多行設定檔（nginx.conf、postgresql.conf）→ volume mount
>   - TLS 證書（多行 PEM）→ volume mount + Secret
>   - JSON 大型設定 → volume mount

> Q: env 注入的本質 vs volume mount 的本質？
> A: 我自己抓到的關鍵：「env 是 config，volume mount 是容器」。精確版：
>   - env 注入 = **啟動時拍快照**：值複製到 Pod 環境變數，從此跟 ConfigMap 脫鉤
>   - volume mount = **持續同步**：K8s kubelet 持續把 ConfigMap 反映到容器內檔案（約 1 分鐘 sync 一次）
>   比喻：env 是拍照📸（照片不會變），volume mount 是直播📡（畫面持續更新）。

> Q: 改了 ConfigMap 後，volume mount 自動更新檔案，程式會自動用新設定嗎？
> A: **不會**。檔案更新 ≠ 程式收到通知。
>   - nginx：要 `nginx -s reload`（發 SIGHUP）
>   - postgres：要 `SELECT pg_reload_conf();`
>   - Go 程式：預設不會 reload，要自己用 fsnotify 監聽
>   注意 K8s ConfigMap volume 不是「直接寫檔」，是 symlink 切換 → fsnotify 要監聽 Remove + Create 不是 Write。
>   業界懶人解：用 [Reloader](https://github.com/stakater/Reloader) 工具，ConfigMap 變更時自動 trigger rollout restart。

> Q: 為什麼 env 不能 reload？
> A: Linux 限制。process 啟動時讀環境變數 → 複製到自己的記憶體。沒有任何 OS 機制能「從外部更新另一個 process 的環境變數」。K8s 不能違反這個。所以 env 注入永遠是啟動快照。

> Q: 改了 ConfigMap 然後 rollout restart，舊 Pod 還沒被殺、新 Pod 還沒 ready 之前，舊 Pod 用的是哪個值？
> A: **舊值**（這題我答錯了）。Rolling update 是漸進的，不是瞬間切換。過渡期內舊 Pod 還活著，繼續用舊 ConfigMap 的快照值。Service 同時把流量導給新舊 Pod。
> 教訓：**rolling update 是漸進的，不是瞬間切換**。應用程式設計時要考慮過渡期（向前/向後相容）。

> Q: Secret 4 種 type 差在哪？
> A:
>   - **Opaque**：通用（99% 用這個），K8s 不檢查 key 名
>   - **kubernetes.io/tls**：強制必須有 `tls.crt` + `tls.key`，給 Ingress / cert-manager 用
>   - **kubernetes.io/dockerconfigjson**：私有 registry 認證
>   - **kubernetes.io/service-account-token**：K8s 自己管，不用碰

> Q: TLS Secret 用了 `cert` 跟 `key` 的 key 名（不是 tls.crt / tls.key），會怎樣？
> A: **K8s 直接擋 apply**（這題我答錯了）。type 有強制 schema，key 名不對 Secret 連建立都失敗，不是「建得起來但用不了」。錯誤訊息：`data[tls.crt]: Required value`。

> Q: Secret 真的安全嗎？
> A: **base64 不是加密，只是編碼**。`echo "bXlwYXNzd29yZA==" | base64 -d` → mypassword（5 秒解）。
> Secret 的安全靠：
>   1. RBAC 權限（誰能 kubectl get secret）
>   2. etcd 加密（cluster 層級設定，1.13+ 可選）
>   3. 不出現在 log / events
>   4. External Secrets + 雲端 KMS（最安全）

> Q: kubectl get secret -o yaml 印出來看到 bXlwYXNzd29yZA== 該擔心嗎？
> A: **該擔心**（這題我跑題答 git 了）。任何拿到這串 base64 的人 5 秒就能解碼成明文。所以：
>   - 不要備份 secret 到本機檔案（base64 = 明文外流）
>   - 不要在共用 terminal / 螢幕分享時 kubectl get secret
>   - 真的要 backup 用雲端 KMS

## 關鍵理解

### 1. env 注入 = 啟動快照（靜態）

```
T0 Pod 啟動時讀 ConfigMap → 值複製到 process 環境變數
之後 ConfigMap 怎麼改都不影響 Pod
唯一更新方式：rollout restart（重建 Pod）
```

**Linux process 的環境變數無法從外部修改**，這是 OS 級限制，不是 K8s 設計問題。

### 2. volume mount = 持續同步（動態，但有延遲）

```
T0 Pod 啟動，K8s 把 ConfigMap 變成虛擬檔案系統
ConfigMap 改了之後 → kubelet 偵測（約 1 分鐘）→ 容器內檔案自動更新
但程式不會自動 reload，要主動發 signal 或重讀
```

**檔案更新 ≠ 程式收到通知**，這是兩件事。

### 3. K8s 更新 ConfigMap volume 是 symlink 切換

K8s 不是「edit 檔案」，是「建新資料夾 + 切 symlink + 刪舊資料夾」。對 fsnotify 程式的影響：要監聽 Remove + Create，不是 Write。

### 4. env vs volume mount 決策樹

```
這是「簡單 key-value」嗎？
  ├── 是 → envFrom
  └── 否 → volume mount

是「敏感資料」嗎？
  ├── 是 → Secret（不管 env 還是 volume）
  └── 否 → ConfigMap

需要熱更新嗎？
  ├── 需要 → 必須 volume mount
  └── 不需要 → 兩種都行（env 較簡單）
```

### 5. subPath 是 99% 設定檔掛載的必要寫法

```yaml
# ❌ 沒 subPath：整個 /etc/nginx/ 被取代，nginx image 預設的 mime.types 等都消失
volumeMounts:
- mountPath: /etc/nginx
  configMap: { name: my-nginx-config }

# ✅ 用 subPath：只放單一檔案進去
volumeMounts:
- mountPath: /etc/nginx/nginx.conf
  subPath: nginx.conf
```

不用 subPath 等於「**整個資料夾被取代**」是 K8s 設定檔掛載最經典的踩雷。

### 6. env 跟 envFrom 同時用：env 永遠贏

```yaml
envFrom:
- configMapRef: { name: auth-service-config }   # 內含 LOG_LEVEL=info
env:
- name: LOG_LEVEL
  value: "debug"   # ← 這個贏
```

順序不影響結果（K8s 不在乎 env / envFrom 寫的先後）。但**閱讀慣例**是「整批先（envFrom）、例外後（env）」，從一般到特殊。

### 7. optional: true 的隱性風險

```yaml
envFrom:
- secretRef:
    name: auth-service-secrets
    optional: true   # ⚠️ 危險
```

Secret 不存在時 Pod 仍會啟動，但環境變數沒注入 → 程式拿到 zero value（空字串）→ DB 認證失敗 → 業務全壞但 Pod 看起來 Running ✅。

「Pod 直接壞掉」反而比「Pod 看起來健康但業務壞」好 debug。**核心設定絕對不要 optional**。

### 8. Rolling update 是漸進的，過渡期混合新舊 Pod

```
T0：舊舊舊舊（全部用舊 ConfigMap）
T1：舊舊舊新（過渡期開始）
T2：舊舊新新（過渡期）
T3：舊新新新（過渡期結束）
T4：新新新新（完成）
```

**過渡期間 Service 同時把流量導給新舊 Pod**。應用程式設計要考慮這個過渡期：
- 改 DB schema 時必須向前/向後相容
- feature flag 切換不會「全部用戶同時切」
- DB 遷移要走「雙寫」過渡期，不能直接切換

### 9. Secret type 不是裝飾，是 schema 契約

| type | 強制 schema |
|---|---|
| Opaque | 無（自由）|
| kubernetes.io/tls | 必須 `tls.crt` + `tls.key` |
| kubernetes.io/dockerconfigjson | 必須 `.dockerconfigjson` |

**K8s 在 apply 階段就擋 schema 錯誤**，沒有「建得起來但用不了」中間狀態。會擋的原因：其他元件（Ingress / cert-manager）寫死去找特定 key 名。

### 10. base64 ≠ 加密

```
編碼 (encoding)：可逆，只是換格式 → base64 屬於這類
加密 (encryption)：需密鑰才能還原 → AES / RSA

bXlwYXNzd29yZA== → echo | base64 -d → mypassword（5 秒解）
```

**Secret 在 etcd 是 base64 儲存**（除非設了 etcd encryption）。安全性靠 RBAC + etcd 加密 + 不外流，**不是靠 base64**。

### 11. 安全層級階梯（從最不安全到最安全）

```
ConfigMap 寫密碼            ← 完全暴露（任何人能讀）
↓
Secret（沒 etcd 加密）       ← base64，懂技術 5 秒解
↓
Secret + etcd 加密           ← 真加密，但 kubectl get 還是會印 base64
↓
Secret + RBAC 嚴格限制       ← 限制誰能 kubectl get
↓
External Secrets + 雲端 KMS  ← 密碼根本不在 K8s 裡，autoreload + 版本控管
```

cosmic-void 目前在第 2 層，生產環境部署到雲端時值得升到第 5 層。

## 程式碼 / 設定

### env 進階用法整合範例

```yaml
spec:
  containers:
  - name: auth-service
    # 1. envFrom 整批匯入（基礎設定）
    envFrom:
    - configMapRef:
        name: auth-service-config
    - secretRef:
        name: auth-service-secrets
    # 2. env 局部覆寫 + rename
    env:
    - name: LOG_LEVEL
      value: "debug"                # 覆寫 ConfigMap 的同名 key（env 贏）
    - name: POSTGRES_HOST            # rename：容器內叫 POSTGRES_HOST
      valueFrom:
        configMapKeyRef:
          name: auth-service-config
          key: DB_HOST                # 但 ConfigMap 裡叫 DB_HOST
```

### volume mount 完整範例（給未來 nginx 用）

```yaml
# ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-config
data:
  nginx.conf: |
    server {
      listen 80;
      location / {
        proxy_pass http://auth-service:7003;
      }
    }

---
# Deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx
        volumeMounts:
        - name: config-volume
          mountPath: /etc/nginx/nginx.conf   # ← 完整檔案路徑
          subPath: nginx.conf                 # ← ⭐ 必須用，避免覆蓋整個資料夾
          readOnly: true
      volumes:
      - name: config-volume
        configMap:
          name: nginx-config
```

### Secret 4 種 type 範例

```yaml
# Opaque（你已經會的）
type: Opaque
stringData:
  ANY_KEY: "..."

# TLS（給 Ingress 用）
type: kubernetes.io/tls
data:
  tls.crt: <base64>          # ← 強制 key 名
  tls.key: <base64>          # ← 強制 key 名

# Docker registry（拉私有 image 用）
# 用指令建比寫 YAML 容易：
# kubectl create secret docker-registry my-registry \
#   --docker-server=... --docker-username=... --docker-password=...

# 在 Pod 引用 docker registry secret
spec:
  imagePullSecrets:
  - name: my-registry
  containers:
  - image: my-private-registry.com/auth-service:v1
```

### immutable Secret（大型 cluster 優化）

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secrets
type: Opaque
immutable: true              # ← 唯讀，K8s 不需 watch，提升 API server 效能
stringData:
  DB_PASSWORD: "..."
```

cosmic-void 規模還不需要，但知道有這個。

### External Secrets 範例（未來部署到雲端時的方向）

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: auth-service-secrets
spec:
  refreshInterval: 1h          # 自動 sync
  secretStoreRef:
    name: gcp-secret-store
    kind: SecretStore
  target:
    name: auth-service-secrets  # K8s Secret 名（給 deployment 用）
  data:
  - secretKey: DB_PASSWORD
    remoteRef:
      key: prod/auth/db-password   # GCP Secret Manager 路徑
```

## 踩過的坑

- 問題：誤以為「volume mount 是重啟時才重新載入」
  解法：搞清楚正好相反 —— **env 才是啟動快照（重啟才能更新），volume mount 是持續同步（不重啟也會更新）**
  為什麼：直覺上會把「快照」跟「啟動」連在一起反過來想。記法：env 拍照📸（不會變），volume mount 直播📡（持續更新）。

- 問題：誤以為「TLS Secret 用 cert/key 當 key 名會建立成功但連線失敗」
  解法：K8s 在 apply 階段就擋 schema 錯誤
  為什麼：第二題答錯。`type: kubernetes.io/tls` 是強制 schema，key 名不對 K8s 直接拒絕建立 Secret。**沒有「建得起來但用不了」這個中間狀態**。

- 問題：誤以為「rollout restart 後所有 Pod 立刻拿到新 ConfigMap」
  解法：理解 rolling update 是「漸進」不是「瞬間」
  為什麼：第三題答錯。過渡期間舊 Pod 還在用舊快照值，Service 同時把流量導給新舊 Pod。設計應用時要考慮過渡期 → 向前/向後相容。

- 問題：被問「kubectl get secret -o yaml 印出 base64 該擔心嗎」答到 git 去了
  解法：核心是 base64 不是加密
  為什麼：把 Q3「base64 外流」跟之前學的「Secret 不進 git」混在一起了。兩個都是 Secret 安全的問題，但 Q3 問的是「**base64 字串本身是否安全**」（答案：不安全，5 秒能解碼）。

- 問題：mountPath 直接掛資料夾沒用 subPath
  解法：99% 場景都該用 subPath
  為什麼：不用 subPath 會整個資料夾被取代，nginx image 預設的 mime.types / fastcgi_params 等檔案全消失，nginx 起不來。**這是 K8s 設定檔掛載最經典的雷**。

- 問題：對 cosmic-void 的 envFrom 是否該改成 volume mount 猶豫
  解法：不該改
  為什麼：cosmic-void 的設定全是 key-value 短字串（不是檔案）+ Go 程式用 os.Getenv() + 連線池在啟動時建好（reload 也沒用）。env 注入是這種場景的最佳選擇。**volume mount 是給多行設定 / 證書 / 真的需要熱更新的場景**。

- 問題：optional: true 的隱性危險
  解法：核心設定絕對不要 optional，否則「Pod 看起來健康但業務全壞」
  為什麼：Secret 不存在時 Pod 仍會啟動，但環境變數沒注入 → 程式拿到 zero value 連 DB → 認證失敗。**Pod 直接壞掉反而比「看起來健康但業務壞」好 debug**。只用在「真的可有可無」的設定（feature flags / debug 開關）。

## 待釐清

- [ ] auth-service 的密碼輪換流程：實際操作一次（改 Postgres 密碼 + 改 Secret + rollout restart + 過渡期雙密碼處理）
- [ ] cosmic-void 部署到雲端時何時導入 External Secrets（GCP Secret Manager）
- [ ] etcd 加密怎麼設？是 cluster admin 設的還是個別開發者能控的？
- [ ] Reloader 工具是否值得導入 cosmic-void（或自己控 rollout 即可）
- [ ] LOG_LEVEL 想做動態切換（不重啟 Pod 也能改 log level）的話，Go 程式需要怎麼設計？fsnotify 在 K8s ConfigMap volume 的正確用法
- [ ] cert-manager 自動發 / 續 TLS 證書（之後做 Ingress HTTPS 時會用）

## 相關專案檔案

- `game-server/auth-service/k8s/configmap.yml` ← 簡單 key-value envFrom 用法
- `game-server/auth-service/k8s/secret.yml.example` ← Opaque type 範本
- `game-server/auth-service/k8s/deployment.yml` ← envFrom 引用兩者

## 相關 learning notes

- [[deployment]] — Deployment 概念（Pod 生命週期、rolling update）
- [[auth-service-deployment-practice]] — 第一次寫 ConfigMap + Secret 的實戰
- [[service]] — Service 提供穩定 hostname（envFrom 注入的 DB_HOST 都是 Service 名）
- [[kubectl-cheatsheet]] — rollout restart / get secret 等指令
- 未來會有 [[ingress]] — Ingress 會引用 TLS Secret 做 HTTPS
