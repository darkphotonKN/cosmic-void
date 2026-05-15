---
topic: gke-deployment
subtopic: db-per-service
date: 2026-05-15
extracted-to-vault: ""
---

# DB 隔離 + Postgres Extension 踩雷

## 起因：第一次 deploy 後一堆 CrashLoopBackOff

```
auth-service log:
  Connected to the database successfully
  Failed to run migrations: no migration found for version 15: read down for version 15 .: file does not exist
```

```
example-service log:
  Failed to run migrations: no migration found for version 15: ...
```

```
stats-service log:
  Could not run migrations: Dirty database version 1. Fix and force version.
```

## 根因：所有 service 共用一個 DB，migrations 撞車

一開始 configmap 全部寫：

```yaml
DB_HOST: "auth-service-db"
DB_NAME: "cosmic_void_auth_service_db"
```

7 個 service 連同一個 DB。每個 service 啟動跑 `migrate.Up()`，把自己的 migrations 套到**同一個 schema_migrations 表**：

| service 啟動順序 | schema_migrations.version |
|---|---|
| auth (假設先) | up to 15 |
| stats（自己 migrations 沒 v1-v14） | 看到 v15 不認識 → "no migration found for version 15" |

互相把對方的 migration 紀錄當成「奇怪的版本號」拒絕跑。

### Stats-service 的 dirty migration

```
Dirty database version 1
```

stats 自己跑了 migration v1（CREATE TABLE），中途用 `uuid_generate_v4()` 函式，但 PG 預設沒裝 uuid-ossp 擴充，function 不存在 → CREATE TABLE 失敗 → schema_migrations 標 dirty。

`Dirty database version N` 是 golang-migrate 的設計：N migration 跑到一半失敗就標記，**之後再啟動會拒絕 retry**，要 operator 介入。

## 修法 1：每個 service 自己的 DB

```bash
# 同一個 Postgres pod 內建 7 個 DB
for db in cosmic_void_items_service_db \
          cosmic_void_stats_service_db \
          cosmic_void_notification_service_db \
          cosmic_void_payment_service_db \
          cosmic_void_example_service_db \
          cosmic_void_game_service_db; do
  kubectl -n cosmic-void exec auth-service-db-0 -- \
    psql -U user -d cosmic_void_auth_service_db -c "CREATE DATABASE $db;"
done
```

每個 service configmap 改成自己的 DB：

```diff
- DB_NAME: "cosmic_void_auth_service_db"
+ DB_NAME: "cosmic_void_<svc>_service_db"
```

| Service | DB Name |
|---|---|
| auth-service | cosmic_void_auth_service_db |
| items-service | cosmic_void_items_service_db |
| stats-service | cosmic_void_stats_service_db |
| notification-service | cosmic_void_notification_service_db |
| payment-service | cosmic_void_payment_service_db |
| example-service | cosmic_void_example_service_db |
| game-service | 用 `STATS_DB_*`，連 cosmic_void_stats_service_db |

**game-service 特例**：它的 `config.InitStatsServiceDB()` 讀 `STATS_DB_*` 環境變數而不是 `DB_*`，且連的是 stats 的 DB（負責讀玩家統計）。

## 修法 2：在所有 7 個 DB 裝 uuid-ossp + pgcrypto extension

```bash
for db in cosmic_void_auth_service_db \
          cosmic_void_items_service_db \
          cosmic_void_stats_service_db \
          cosmic_void_notification_service_db \
          cosmic_void_payment_service_db \
          cosmic_void_example_service_db \
          cosmic_void_game_service_db; do
  kubectl -n cosmic-void exec auth-service-db-0 -- psql -U user -d "$db" \
    -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"; CREATE EXTENSION IF NOT EXISTS pgcrypto;"
done
```

| Function | Extension |
|---|---|
| `uuid_generate_v4()` | uuid-ossp（PG 沒內建） |
| `gen_random_uuid()` | pgcrypto（PG 13+ 已 built-in 但裝 extension 不痛） |

哪些 service 用什麼，可以 grep migrations 看：

```bash
grep -l "uuid_generate_v4" */migrations/*.up.sql
# api-gateway, auth-service, items-service, notification-service, stats-service
grep -l "gen_random_uuid" */migrations/*.up.sql
# auth-service, payment-service, stats-service
```

## 修法 3：清掉 dirty migration（drop & recreate stats DB）

stats DB 已標 dirty，重啟也不會自動清。最簡單：drop 後重建。

```bash
# 先 kick 出 connections
kubectl -n cosmic-void exec auth-service-db-0 -- psql -U user -d postgres -c "
  SELECT pg_terminate_backend(pid) FROM pg_stat_activity
  WHERE datname = 'cosmic_void_stats_service_db' AND pid <> pg_backend_pid();
"

kubectl -n cosmic-void exec auth-service-db-0 -- psql -U user -d postgres \
  -c "DROP DATABASE cosmic_void_stats_service_db;"
kubectl -n cosmic-void exec auth-service-db-0 -- psql -U user -d postgres \
  -c "CREATE DATABASE cosmic_void_stats_service_db;"

# Install extension on fresh DB
kubectl -n cosmic-void exec auth-service-db-0 -- psql -U user -d cosmic_void_stats_service_db \
  -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"; CREATE EXTENSION IF NOT EXISTS pgcrypto;"

# Restart service，這次 migration 會在乾淨 DB 上重跑
kubectl -n cosmic-void rollout restart deployment stats-service
```

對 auth-service DB 也要做（它的 DB 之前被別人髒過）。

## 教訓

1. **微服務的 DB 從第一天就要隔離** — 不要先用一個 DB 想著「之後再分」。Migration 撞車的 debug 比一開始就分多花 10×時間。
2. **PG 沒有預設 extension**。`uuid_generate_v4()` 寫了不裝會 silent 死掉。
3. **golang-migrate 的 dirty 狀態是 fail-safe 不是 fail-fast**，要 operator 介入清。
4. **dropdb 前要 terminate 連線**，不然會卡。

## 長期架構建議

學習階段這樣可以。Production 該：

- 每個 service 自己的 PostgreSQL instance（StatefulSet 各自獨立），不共用 Pod
- Migration 用 init container 跑（不要 service code 自己跑）
- 或用 Cloud SQL + 各 service 自己的 DB instance
