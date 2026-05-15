---
title: "Outbox Worker 併發控制：FOR UPDATE SKIP LOCKED"
type: learning-note
project: cosmic-void
topic: messaging
date: 2026-05-14
status: learning
extracted-to-vault: []
related-files:
  - game-server/common/outbox/repository.go
  - game-server/common/outbox/service.go
  - game-server/common/outbox/worker.go
  - game-server/auth-service/cmd/server/main.go
  - game-server/stats-service/cmd/server/main.go
tags: [messaging, postgresql, concurrency, row-lock, skip-locked, outbox, worker]
---

## 學習目標
搞懂 outbox worker 在多 replica 下會踩什麼雷，以及 Postgres 的 `FOR UPDATE SKIP LOCKED` 在這個場景的應用。

## 對話脈絡

> Q: items-service 有可能 scale，原本的 outbox 程式碼會壞嗎？
> A: 會踩三個坑：(1) 兩個 worker 撈到同一批 row → 重複 publish (2) `LIMIT $1` 參數沒 bind 進 SQL → 一次撈光所有 pending (3) goroutine retry update 在 worker shutdown 時殘留 → 重啟後也會重複 publish

> Q: 都不修會發生什麼？
> A: 三個問題的後果都一樣：「會重複 publish，但下游 inbox 擋得住，所以資料正確」。也就是說，**靠下游 inbox 兜底**，自己這層不健壯。

> Q: 怎麼修？
> A: 把 SELECT 和 UPDATE 包在同一個 tx，SELECT 加 `FOR UPDATE SKIP LOCKED LIMIT $1`，刪掉 retry goroutine。

## 關鍵理解

### 為什麼 `FOR UPDATE` 不夠，要加 `SKIP LOCKED`

| SQL | 行為 |
|---|---|
| `SELECT ...` | 兩個 worker 同時拿到同一批 row |
| `SELECT ... FOR UPDATE` | Worker B 卡住等 Worker A 釋放鎖（變成串行，沒 scale） |
| `SELECT ... FOR UPDATE SKIP LOCKED` | Worker B 跳過已鎖的，拿到下一批（真正並行） |

### 為什麼 SELECT 跟 UPDATE 必須在同一個 tx
`FOR UPDATE` 拿到的 row lock **只在 transaction 期間有效**。
- 如果 SELECT 用 tx、UPDATE 不用 tx → 鎖在 SELECT 結束時就釋放 → 等於沒鎖
- 兩者必須包在同一 tx，UPDATE 完 commit 才釋放鎖

### 為什麼可以刪掉 retry goroutine
原本的 retry goroutine 在解決「UPDATE 失敗時的補償」，改成 tx 後：
- UPDATE 失敗 → continue 下一筆（這筆 published_at 維持 NULL）
- Commit 失敗 → 整批 rollback → 下個 cycle 重撈
- 所有失敗路徑都自然回到「下次 worker 重試」，不需要額外 retry goroutine
- 額外好處：避免 worker shutdown 時 goroutine 殘留導致 update 不完整

### 為什麼 ORDER BY 還要保留
即使 SKIP LOCKED 會打亂跨 worker 之間的順序：
- 單一 worker 內：拿到的 row 還是依 created_at 排序
- 跨 worker 之間：粗略 FIFO（worker B 跳過 1-20，拿 21-40，順序大致對）
- Event 之間本來就沒嚴格順序依賴，這樣足夠了

### 為什麼下游 inbox 即使有了 outbox 修正還是必要
At-least-once 不可能完全消滅，只能逐層降低機率：
- Outbox worker 修正後：publisher 端去重 → 大幅減少
- 但 worker shutdown 時 commit 失敗仍可能重送
- 所以 inbox 還是要在（防最後一道）

## 程式碼

### 修正後的 SQL（核心）
```go
// game-server/common/outbox/repository.go
func (r *repo) GetUnpublishedOutboxItems(ctx context.Context, tx *sqlx.Tx, limit int) ([]*OutboxEvent, error) {
    query := `
    SELECT id, routing_key, exchange, payload, created_at
    FROM outbox
    WHERE published_at IS NULL
    ORDER BY created_at ASC
    LIMIT $1
    FOR UPDATE SKIP LOCKED
    `
    err := tx.SelectContext(ctx, &outboxItem, query, limit)
    ...
}
```
三個關鍵：
- `LIMIT $1` 真的 bind（之前 `*int` 收進來但 SQL 沒寫 LIMIT，等於沒生效）
- `FOR UPDATE SKIP LOCKED` 跨 worker 不卡
- 用 `tx.SelectContext` 不是 `r.db.SelectContext`

### 修正後的 worker（單 tx per batch）
```go
// game-server/common/outbox/worker.go
func (w *OutboxWorker) PublishOutboxEvents(ctx context.Context) error {
    tx, err := w.db.BeginTxx(ctx, nil)
    if err != nil { return err }
    defer func() { _ = tx.Rollback() }()

    outboxEvts, err := w.outboxRetriever.GetPendingOutboxItems(ctx, tx, w.batchCount)
    if err != nil { return err }
    if len(outboxEvts) == 0 { return nil }

    for _, evt := range outboxEvts {
        if err := w.publishCh.PublishWithContext(...); err != nil {
            slog.Error("publish failed", "outbox_id", evt.ID)
            continue  // 這筆 published_at 維持 NULL，下次重試
        }
        if err := w.outboxRetriever.UpdateOutboxToPublished(ctx, tx, evt.ID); err != nil {
            continue  // 也是維持 pending，下次重試
        }
    }

    return tx.Commit()
}
```

## 踩過的坑

- 問題：原本以為 `LIMIT *int` 參數會自動接到 SQL
  解法：發現 SQL 字串裡根本沒有 `LIMIT` 子句，参数完全沒生效。要 SQL 寫 `LIMIT $1` + caller 傳 limit 值才會生效
  為什麼：Go 的 named parameter 跟 SQL placeholder 不會自動連動，要手動 bind

- 問題：以為加 `FOR UPDATE` 就夠了
  解法：實際上沒 `SKIP LOCKED` 的話，Worker B 會 block 等 Worker A 釋放鎖，變成串行
  為什麼：`FOR UPDATE` 的預設語意是「我要這些 row，沒拿到就等」，不是「我要這些 row，被別人拿了就放棄」

- 問題：以為 SELECT 用 tx 就夠了，UPDATE 可以用 `r.db`
  解法：那 SELECT 釋放鎖後 UPDATE 才執行，跨 connection 沒 lock 保護
  為什麼：Row lock 跟 tx 綁定，跨 connection（不同 tx）看不到對方的鎖

- 問題：複製貼上時手滑把中文提示「改成：」貼進 .go 檔
  解法：刪掉那行
  為什麼：人工複製貼上時要注意，編譯器會抓出非 ASCII 字元（`U+FF1A '：'`）

- 問題：刪了 retry goroutine 後 `commonhelpers` import 變成 unused
  解法：手動刪 import
  為什麼：Go 對 unused import 是編譯錯誤（不是 warning），改 code 要順手清

## 改了什麼

1. `common/outbox/repository.go`
   - `GetUnpublishedOutboxItems` 改簽名：加 `tx *sqlx.Tx`、`limit *int` 改成 `int`
   - SQL 加 `LIMIT $1 FOR UPDATE SKIP LOCKED`
   - 用 `tx.SelectContext` 而不是 `r.db.SelectContext`
   - `UpdateOutboxToPublished` 同樣改簽名 + 用 `tx.ExecContext`

2. `common/outbox/service.go`
   - `Repository` interface 兩個方法簽名跟著改
   - service wrapper 跟著改

3. `common/outbox/worker.go`
   - `OutboxRetriever` interface 兩個方法簽名跟著改
   - `OutboxWorker` struct 加 `db *sqlx.DB` 欄位
   - `NewOutboxWorker` 簽名加 `db` 參數
   - `PublishOutboxEvents` 整段重寫：BeginTxx → SELECT → loop publish+UPDATE → Commit
   - **刪掉**原本 `go func` retry goroutine 那段
   - 刪掉 unused `commonhelpers` import

4. `auth-service/cmd/server/main.go:178`
   - `NewOutboxWorker(time.Second*5, 20, outboxService, publishCh)` → 中間加 `db`

5. `stats-service/cmd/server/main.go:143`
   - 同樣加 `db`

## 待釐清
- [ ] Postgres `FOR UPDATE SKIP LOCKED` 在不同 isolation level 的行為差異（READ COMMITTED vs REPEATABLE READ）
- [ ] 如果一個 worker batch 跑很久（publish 慢），其他 worker 一直 SKIP，會不會 outbox 累積？
- [ ] outbox row 的 cleanup 策略（published_at 不為 null 的 row 多久刪一次）

## 相關 learning notes
- [[outbox-pattern]] — outbox pattern 的整體設計
- [[idempotency-inbox-pattern]] — 為什麼修完 outbox 還是需要 inbox
