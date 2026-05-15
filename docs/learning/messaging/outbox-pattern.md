---
title: "Outbox Pattern：發布端原子性"
type: learning-note
project: cosmic-void
topic: messaging
date: 2026-05-14
status: learning
extracted-to-vault: []
related-files:
  - game-server/common/outbox/model.go
  - game-server/common/outbox/repository.go
  - game-server/common/outbox/service.go
  - game-server/common/outbox/worker.go
  - game-server/auth-service/internal/member/service.go
  - game-server/auth-service/migrations/000009_create_outbox.up.sql
  - game-server/auth-service/cmd/server/main.go
  - game-server/items-service/internal/items/service.go
tags: [messaging, rabbitmq, outbox, transactional-outbox, event-driven]
---

## 學習目標
搞懂為什麼 inbox（消費端去重）不夠 — 還需要 outbox（發布端保證業務 DB 跟事件原子性），以及為什麼不是所有 service 都需要 outbox。

## 對話脈絡

> Q: items-service 真的需要 outbox 嗎？
> A: 不一定。Outbox 是工具，要評估「事件遺失的業務成本 vs outbox 的複雜度成本」。items-service 的 ItemCreated 是 admin 內部通知，掉了影響極小 → over-engineering。auth-service 的 member.signedup 是用戶導向通知，丟了用戶會抱怨 → 必要。

> Q: 那為什麼 game-service 的 MatchEnded 也不套 outbox？
> A: 它不是 DB-driven。MatchEnded 是 in-memory 跑 match logic 結束才發，沒有「業務 DB 寫成功 + publish 失敗」這種原子性問題，套上去只是把問題搬家。

> Q: 不套 outbox 會發生什麼？
> A: 場景 C — DB 寫成功但 publish 失敗 → 業務狀態跟事件狀態不一致 → 下游永遠收不到通知。Inbox 救不了這個（事件根本沒發出來）。

## 關鍵理解

### Outbox 解決什麼問題
**「DB 寫入」和「事件 publish」的跨系統原子性**。
- DB 和 message broker 是兩個獨立系統，沒有跨系統 atomic tx
- 但「DB 業務寫入 + DB outbox 寫入」可以在同一個 tx 內保證 atomic
- 然後背景 worker 慢慢把 outbox row publish 到 broker

### 三段式流程
1. **Service 端**：業務 INSERT + outbox INSERT 在同一個 tx 內，commit
2. **Outbox table**：作為「已決定要發、還沒發出去」的緩衝
3. **Worker**：定期 poll pending row → publish → 標記 published_at

### 為什麼 event_id 用 outbox row PK 是好設計
Outbox row 在 DB 裡是穩定的，worker 重試只是 publish 同一筆。
- 同一筆 outbox row 對應同一個 event_id（永遠不變）
- 下游 inbox 看到同一個 event_id → 認得出來是重複
- 解決了「publisher retry 產生新 event_id」的問題（[[idempotency-inbox-pattern]] 提到的）

### Outbox vs Inbox 的職責分離
| | Outbox | Inbox |
|---|---|---|
| 位置 | publisher 端 | consumer 端 |
| 解決 | DB-and-broker 原子性 | 重複消費 |
| 形式 | INSERT 緩衝 + 背景 worker | UNIQUE constraint + dedup |
| 必要性 | 看業務 critical 程度 | 永遠需要（at-least-once delivery） |

### 不是所有 service 都該套 outbox
這次最大的學習：**「能用」不等於「該用」**。
判斷標準：「事件遺失成本 > outbox 的複雜度成本」才該用。
- auth-service member.signedup：✅ 用戶導向，必要
- items-service ItemCreated：❌ admin 內部通知，over-engineering
- game-service MatchEnded：❌ 不是 DB-driven，套了沒意義

## 程式碼

### Outbox 寫入（service 層，跟業務寫入同 tx）
```go
// game-server/auth-service/internal/member/service.go:124
// Build payload up-front so marshal failure aborts before tx
payload := commonconstants.MemberSignedUpEventPayload{
    EventID: uuid.NewString(),
    ...
}

tx, _ := s.db.BeginTxx(ctx, nil)
defer func() { _ = tx.Rollback() }()

memberId, _ := s.Repo.CreateTx(ctx, tx, ...)   // 業務 INSERT
payload.UserID = memberId.String()
marshalledPayload, _ := json.Marshal(payload)

s.outboxPublisher.CreateOutboxTx(ctx, tx, commonoutbox.OutboxParams{
    RoutingKey: commonconstants.MemberSignedUpEvent,
    Exchange:   commonconstants.AuthEventsExchange,
    Payload:    marshalledPayload,
})  // outbox INSERT（同 tx）

tx.Commit()  // 兩個 INSERT 一起落盤
```

### Outbox table schema
```sql
-- game-server/auth-service/migrations/000009_create_outbox.up.sql
CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    routing_key VARCHAR(255) NOT NULL,
    exchange VARCHAR(255) NOT NULL,
    payload BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMPTZ NULL  -- null=pending, not null=published
);

-- Partial index：worker 的 hot path 只看 pending row
CREATE INDEX idx_outbox_pending ON outbox (created_at)
    WHERE published_at IS NULL;
```

### Worker 部署（in-process goroutine）
```go
// game-server/auth-service/cmd/server/main.go:172
outboxRepo := commonoutbox.NewRepo(db)
outboxService := commonoutbox.NewService(outboxRepo)

memberService := member.NewService(db, memberRepo, publishCh, cacheService, outboxService)

outboxWorker := commonoutbox.NewOutboxWorker(time.Second*5, 20, db, outboxService, publishCh)
go outboxWorker.Run(workerCtx)
```

## 踩過的坑

- 問題：一開始想對 items-service 也套 outbox
  解法：認真評估後發現是 over-engineering，撤掉計畫，改修 common/outbox 併發問題
  為什麼：被「pattern 看起來很酷、auth-service 已經有範本」吸引，沒先想「真的需要嗎」。學習導向跟工程必要性是兩件事

- 問題：以為 outbox 也能解決「客戶端 retry 造成業務操作重複」
  解法：發現 outbox 只解決「業務 DB 寫了但 publish 失敗」的不一致，「業務操作本身重複」要靠 API 層 idempotency key
  為什麼：outbox 是 publisher 內部的原子性，不是 publisher 對外的去重

- 問題：把 outbox 和 inbox 想成「互相替代」
  解法：實際是「互補」— outbox 保業務跟事件原子，inbox 保消費端不重複
  為什麼：兩個都是針對 at-least-once 的不同面向

## 改了什麼

這次練習的本意是要在 items-service 套 outbox，但**最後決定不做**（理由見「踩過的坑」）。所以這次 outbox 本身沒新增，只是讀 auth-service 既有實作搞懂 pattern。

實際做的事：
1. trace auth-service outbox 完整實作（migration、common 套件、service 整合、worker 部署）
2. 找出 common/outbox 的併發問題（見 [[outbox-worker-concurrency]]）

## 待釐清
- [ ] outbox row 累積大量 published_at != null 的死 row，何時 cleanup？目前沒看到 cleanup job
- [ ] 如果 outbox INSERT 成功但業務 INSERT 失敗會怎樣？（理論上 tx rollback 兩邊都不存在，但要驗證）
- [ ] 為什麼 auth-service 是 5 秒 cycle、stats-service 是 1 分鐘 cycle？延遲跟負載的權衡邏輯

## 相關 learning notes
- [[idempotency-inbox-pattern]] — 消費端的對偶問題
- [[outbox-worker-concurrency]] — 多 worker scale 時的併發控制
