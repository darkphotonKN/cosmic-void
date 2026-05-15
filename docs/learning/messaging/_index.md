# Messaging 學習路徑

> 開始日期：2026-05-14
> 狀態：learning

## 已寫的筆記
- [[idempotency-inbox-pattern]] — 2026-05-14
- [[outbox-pattern]] — 2026-05-14
- [[outbox-worker-concurrency]] — 2026-05-14

## 學習目標
理解 message broker（RabbitMQ）event-driven 架構下的可靠性 pattern：
- 為什麼 at-least-once delivery 會造成重複事件
- 消費端怎麼擋（inbox pattern）
- 發布端怎麼保證業務 DB 跟事件原子性（outbox pattern）
- outbox worker 在多 replica 下的併發控制

這三個是組合拳，缺一不可。

## 已萃取進 vault 的概念
（由 extract-to-vault 自動回填，這裡先空著）
