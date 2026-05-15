---
title: "Inbox Pattern：消費端去重"
type: learning-note
project: cosmic-void
topic: messaging
date: 2026-05-14
status: learning
extracted-to-vault: []
related-files:
  - game-server/notification-service/internal/notification/inbox_repository.go
  - game-server/notification-service/internal/notification/service.go
  - game-server/notification-service/internal/notification/amqp_consumer.go
  - game-server/notification-service/migrations/000004_processed_events.up.sql
  - game-server/common/constants/events.go
  - game-server/common/api/proto/events/game.proto
  - game-server/common/api/proto/events/item.proto
tags: [messaging, rabbitmq, idempotency, inbox, event-driven]
---

## 學習目標
搞懂為什麼 notification-service `insertNotification` 看起來「沒做 row lock」，以及 message broker 場景下重複事件的真正成因和擋法。

## 對話脈絡

> Q: `insertNotification` 沒做 row lock，會不會導致 UserID 寫入到同一個值？
> A: 不會。INSERT 不是讀-改-寫，每個 goroutine 有自己的 `CreateNotification` instance，記憶體不共用。Row lock 是給「先讀再依內容更新」的場景（避免 lost update），純 INSERT 不需要。

> Q: 那真正可能的問題是什麼？
> A: 重複寫入。上游 retry / broker at-least-once delivery / network timeout 都會讓同一個事件被消費 2 次以上，使用者看到兩筆一樣的通知。

> Q: idempotency 該怎麼做？
> A: 不要在 application 層「先 SELECT 再 INSERT」（race condition），讓 DB 的 UNIQUE constraint 保證唯一性。具體做法是 inbox table。

> Q: `CreateNotification` 沒有可以當 key 的欄位，必須在上游 `MatchEndedEvent` 加 event_id 才行？
> A: 對。member.signedup 已經有 EventID（`MemberSignedUpEventPayload`），但 `MatchEndedEvent` 和 `ItemCreatedEvent` 兩個 proto 當時沒加。

## 關鍵理解

### 為什麼會有重複事件
不是 application 邏輯 bug，是 **broker delivery semantics** 的本質：
- RabbitMQ 是 at-least-once，consumer crash / nack / 連線斷 → message 重送
- 上游 publisher retry 也會發新的事件
- 同一個邏輯事件，可能在 broker 上出現 N 次

### Inbox Pattern 怎麼擋
- 維護一張 `processed_events` table，PK = `(event_id, event_type)`
- 處理新事件前，先 INSERT 進 inbox：
  - `ON CONFLICT (event_id, event_type) DO NOTHING`
  - `RowsAffected() == 1` → 新事件，繼續處理
  - `RowsAffected() == 0` → 已處理過，回傳 `ErrAlreadyProcessed`
- 關鍵：inbox INSERT 和 business INSERT **在同一個 tx**，要嘛都成功要嘛都 rollback

### Inbox 跟在 business table 加 idempotency_key 的差異
- inbox 適合「一個事件 → 多張 table 寫入」的場景，一筆 inbox 保護全部
- business table 加 idempotency_key 只能保護自己
- 這個 codebase 選 inbox，跨 business table 的擴展性更好

### Consumer 端要處理 `ErrAlreadyProcessed`
不是 retry、不是 DLQ，是直接 `msg.Ack(false)` 讓 message 離開 queue。重複事件本身不是錯誤。

## 程式碼

### Inbox repository（核心去重邏輯）
```go
// game-server/notification-service/internal/notification/inbox_repository.go
func (r *inboxRepository) MarkEventProcessed(ctx context.Context, tx *sqlx.Tx, eventID uuid.UUID, eventType string) (bool, error) {
    query := `
        INSERT INTO processed_events (event_id, event_type)
        VALUES ($1, $2)
        ON CONFLICT (event_id, event_type) DO NOTHING
    `
    result, err := tx.ExecContext(ctx, query, eventID, eventType)
    if err != nil {
        return false, commonutils.AnalyzeDBErr(err)
    }
    rows, err := result.RowsAffected()
    return rows == 1, err
}
```

### Service 層的 tx wrapper
```go
// game-server/notification-service/internal/notification/service.go:89
func (s *service) runInTx(ctx context.Context, eventID uuid.UUID, eventType string, fn func(tx *sqlx.Tx) error) error {
    return commonutils.ExecTx(ctx, s.db, func(tx *sqlx.Tx) error {
        inserted, err := s.inboxRepo.MarkEventProcessed(ctx, tx, eventID, eventType)
        if err != nil {
            return fmt.Errorf("mark event processed: %w", err)
        }
        if !inserted {
            return commonconstants.ErrAlreadyProcessed
        }
        return fn(tx)
    })
}
```

### Consumer 端處理重複
```go
// game-server/notification-service/internal/notification/amqp_consumer.go:258
if errors.Is(err, commonconstants.ErrAlreadyProcessed) {
    msg.Ack(false) // 不重試、不 DLQ，直接 ack
    return
}
```

### Schema
```sql
-- game-server/notification-service/migrations/000004_processed_events.up.sql
CREATE TABLE processed_events (
    event_id      UUID NOT NULL,
    event_type    TEXT NOT NULL,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, event_type)
);
```

## 踩過的坑

- 問題：一開始懷疑 `insertNotification` 要做 row lock，方向錯了
  解法：搞清楚 INSERT vs UPDATE 的差異 — 純 INSERT 不會 race，是「同一邏輯事件被處理多次」才是真問題
  為什麼：把「並行寫入同一張表」和「同一邏輯操作重複執行」混在一起。前者要 lock，後者要 idempotency key

- 問題：proto 沒 event_id 但 consumer 已有 `ErrAlreadyProcessed` 邏輯
  解法：發現 inbox 是「半成品」— code 已經規劃好但 proto 那層沒同步加欄位
  為什麼：feature 跨 service 開發時，schema 跟 consumer 邏輯不一定同步進度

- 問題：上游用 outbox 的話 event_id 才穩定
  解法：member.signedup 的 publisher（auth-service）有 outbox，所以 event_id 跨 retry 穩定；其他事件沒 outbox 時 retry 會產生新 event_id，inbox 認不出來
  為什麼：inbox 只擋 broker-layer redelivery（同一 message body 重送），擋不住「publisher 重新組事件」

## 改了什麼

1. `MatchEndedEvent` proto 加 `string event_id = 5;`
2. `ItemCreatedEvent` proto 加 `string event_id = 4;`
3. game-service publisher（`MatchEndedEvent` 組成處）加 `EventId: uuid.NewString()`
4. items-service publisher 兩處（`CreateItemTemplate` 和 `publishItemCreatedEvent`）加 `EventId: uuid.NewString()`

## 待釐清
- [ ] item.created / game.ended 的 service 還沒套 `runInTx`，只有 member.signedup 套了
- [ ] game-service publisher 因為沒 outbox，publisher 端 retry 還是會產生新 event_id（這個是 outbox-pattern 的範圍）

## 相關 learning notes
- [[outbox-pattern]] — publisher 端的對偶問題
- [[outbox-worker-concurrency]] — outbox worker 的併發控制
