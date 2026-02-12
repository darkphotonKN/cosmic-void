package notification

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

// ==========================================
// Retry Config（統一管理，避免重複定義）
// ==========================================

type RetryLevel struct {
	Level        string
	TTL          int32
	DelaySeconds int
}

var RetryLevels = []RetryLevel{
	{"retry-1", 5000, 5},   // 第1次重試：等 5 秒
	{"retry-2", 15000, 15}, // 第2次重試：等 15 秒
	{"retry-3", 60000, 60}, // 第3次重試：等 60 秒
}

// MaxRetries 從 RetryLevels 推導，不用寫死
var MaxRetries = len(RetryLevels)

// ==========================================
// Consumer
// ==========================================

// ConsumerService defines what the consumer needs from the service
type ConsumerService interface {
	ProcessMemberSignedUp(ctx context.Context, payload *commonconstants.MemberSignedUpEventPayload) error
	ProcessItemCreated(ctx context.Context, payload *pb.ItemCreatedEvent) error
}

type Consumer struct {
	service ConsumerService
	channel *amqp.Channel
}

func NewConsumer(service ConsumerService, channel *amqp.Channel) *Consumer {
	return &Consumer{
		service: service,
		channel: channel,
	}
}

// Listen starts consuming messages from the configured queues
func (c *Consumer) Listen() {
	go c.consumeMemberSignedUp()
	go c.consumeItemCreated()
	slog.Info("Notification consumer listening for events...")
}

// ==========================================
// Consumer: Item Created
// ==========================================

func (c *Consumer) consumeItemCreated() {
	msgs, err := c.channel.Consume(
		commonconstants.NotificationItemCreatedQueue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
		return
	}

	slog.Info("Started consuming item.created events")

	for msg := range msgs {
		c.handleItemCreated(msg)
	}
}

func (c *Consumer) handleItemCreated(msg amqp.Delivery) {
	retryCount := getRetryCount(msg)

	var payload pb.ItemCreatedEvent
	if err := proto.Unmarshal(msg.Body, &payload); err != nil {
		slog.Error("Failed to parse ItemCreatedEvent", "error", err)
		msg.Nack(false, false)
		return
	}

	slog.Info("Received item created event",
		"item_type", payload.ItemType,
		"name", payload.Name,
		"retry_count", retryCount,
	)

	ctx := context.Background()
	err := c.service.ProcessItemCreated(ctx, &payload)

	c.handleResult(msg, err, retryCount, "item.created", payload.UserId)
}

// ==========================================
// Consumer: Member Signed Up
// ==========================================

func (c *Consumer) consumeMemberSignedUp() {
	msgs, err := c.channel.Consume(
		commonconstants.NotificationMemberSignedUpQueue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
		return
	}

	slog.Info("Started consuming member.signedup events")

	for msg := range msgs {
		c.handleMemberSignedUp(msg)
	}
}

func (c *Consumer) handleMemberSignedUp(msg amqp.Delivery) {
	retryCount := getRetryCount(msg)

	var payload commonconstants.MemberSignedUpEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		slog.Error("Failed to parse MemberSignedUpEvent", "error", err)
		msg.Nack(false, false)
		return
	}

	slog.Info("Received member signed up event",
		"user_id", payload.UserID,
		"name", payload.Name,
		"retry_count", retryCount,
	)

	ctx := context.Background()
	err := c.service.ProcessMemberSignedUp(ctx, &payload)

	c.handleResult(msg, err, retryCount, "member.signedup", payload.UserID)
}

// ==========================================
// 統一的錯誤處理邏輯（消除重複）
// ==========================================

func (c *Consumer) handleResult(msg amqp.Delivery, err error, retryCount int, eventType string, identifier string) {
	if err == nil {
		msg.Ack(false)
		slog.Info("Event processed successfully",
			"event_type", eventType,
			"identifier", identifier,
		)
		return
	}

	if errors.Is(err, commonconstants.ErrTransient) {
		if retryCount < MaxRetries {
			slog.Warn("Transient error, requeueing with delay",
				"event_type", eventType,
				"retry", retryCount,
				"max", MaxRetries,
				"error", err,
			)

			if requeueErr := c.requeueWithRetry(msg, retryCount+1); requeueErr != nil {
				slog.Error("Failed to requeue message", "error", requeueErr)
				msg.Nack(false, false) // 重新排隊失敗 -> DLQ
			} else {
				msg.Ack(false) // 原消息 ACK（已重新發送新的到 retry queue）
			}
		} else {
			slog.Error("Max retries exceeded, sending to DLQ",
				"event_type", eventType,
				"retry_count", retryCount,
				"error", err,
			)
			msg.Nack(false, false) // 重試次數用盡 -> DLQ
		}
		return
	}

	// 永久性錯誤
	slog.Error("Permanent error, sending to DLQ",
		"event_type", eventType,
		"error", err,
	)
	msg.Nack(false, false)
}

// ==========================================
// Retry 相關工具函數
// ==========================================

// getRetryCount 從消息 header 獲取重試次數
func getRetryCount(msg amqp.Delivery) int {
	if msg.Headers == nil {
		return 0
	}
	if count, ok := msg.Headers["x-retry-count"].(int32); ok {
		return int(count)
	}
	return 0
}

// requeueWithRetry 發送消息到 retry exchange，實現延遲重試
func (c *Consumer) requeueWithRetry(msg amqp.Delivery, newRetryCount int) error {
	// 複製 headers
	headers := amqp.Table{}

	if msg.Headers != nil {
		for k, v := range msg.Headers {
			headers[k] = v
		}
	}

	// 更新重試計數
	headers["x-retry-count"] = int32(newRetryCount)

	// 清理 RabbitMQ 自動加的 x-death headers，避免隨重試次數不斷膨脹
	delete(headers, "x-death")
	delete(headers, "x-first-death-exchange")
	delete(headers, "x-first-death-queue")
	delete(headers, "x-first-death-reason")

	// 根據重試次數選擇延遲級別
	idx := newRetryCount - 1
	if idx >= len(RetryLevels) {
		idx = len(RetryLevels) - 1
	}

	// 組合 retry routing key: e.g. "retry-1.item.created"
	retryRoutingKey := RetryLevels[idx].Level + "." + msg.RoutingKey

	slog.Info("Requeueing message with delay",
		"retry_count", newRetryCount,
		"delay_seconds", RetryLevels[idx].DelaySeconds,
		"retry_routing_key", retryRoutingKey,
	)

	// 發送到 retry exchange（不是原始 exchange）
	return c.channel.Publish(
		commonconstants.RetryExchange, // 關鍵：送到 retry exchange
		retryRoutingKey,               // e.g. "retry-1.item.created"
		false,                         // mandatory
		false,                         // immediate
		amqp.Publishing{
			ContentType:  msg.ContentType,
			Body:         msg.Body,
			DeliveryMode: msg.DeliveryMode,
			Headers:      headers,
		},
	)
}

// ==========================================
// AMQP Infrastructure Setup
// ==========================================

func SetupAMQPInfrastructure(channel *amqp.Channel) error {
	// ==========================================
	// 1. Dead Letter Exchange + DLQ
	// ==========================================
	if err := channel.ExchangeDeclare(
		commonconstants.DlxEventsExchange,
		"topic",
		true, false, false, false, nil,
	); err != nil {
		return err
	}

	if _, err := channel.QueueDeclare(
		commonconstants.NotificationDlqQueue,
		true, false, false, false, nil,
	); err != nil {
		return err
	}

	if err := channel.QueueBind(
		commonconstants.NotificationDlqQueue,
		"#", // 接收所有 DLX 消息
		commonconstants.DlxEventsExchange,
		false, nil,
	); err != nil {
		return err
	}

	// ==========================================
	// 2. 原始 Exchanges（先宣告，因為 retry queue 的 DLX 指向它們）
	// ==========================================
	if err := channel.ExchangeDeclare(
		commonconstants.AuthEventsExchange,
		"topic",
		true, false, false, false, nil,
	); err != nil {
		slog.Error("Failed to declare auth exchange", "error", err)
		return err
	}

	if err := channel.ExchangeDeclare(
		commonconstants.ItemEventsExchange,
		"topic",
		true, false, false, false, nil,
	); err != nil {
		slog.Error("Failed to declare item exchange", "error", err)
		return err
	}

	// ==========================================
	// 3. Retry Exchange + Retry Queues
	// ==========================================
	if err := channel.ExchangeDeclare(
		commonconstants.RetryExchange,
		"topic",
		true, false, false, false, nil,
	); err != nil {
		return err
	}

	// 每種 event type 對應的 retry 設定
	type eventRetryConfig struct {
		eventType  string // e.g. "item.created"
		routingKey string // 原始 routing key
		exchange   string // TTL 到期後送回的 exchange
	}

	eventConfigs := []eventRetryConfig{
		{
			eventType:  "item.created",
			routingKey: commonconstants.ItemCreated,
			exchange:   commonconstants.ItemEventsExchange,
		},
		{
			eventType:  "member.signedup",
			routingKey: commonconstants.MemberSignedUpEvent,
			exchange:   commonconstants.AuthEventsExchange,
		},
	}

	// 為每種 event type 創建不同延遲級別的 retry queue
	for _, config := range eventConfigs {
		for _, retry := range RetryLevels {
			queueName := "retry." + retry.Level + "." + config.eventType
			bindingKey := retry.Level + "." + config.eventType

			if _, err := channel.QueueDeclare(
				queueName,
				true, false, false, false,
				amqp.Table{
					"x-message-ttl":             retry.TTL,         // 消息在此 queue 等待的時間
					"x-dead-letter-exchange":    config.exchange,   // TTL 到期後送回原始 exchange
					"x-dead-letter-routing-key": config.routingKey, // 使用原始 routing key
				},
			); err != nil {
				return err
			}

			if err := channel.QueueBind(
				queueName,
				bindingKey,
				commonconstants.RetryExchange,
				false, nil,
			); err != nil {
				return err
			}

			slog.Info("Retry queue created",
				"queue", queueName,
				"ttl_ms", retry.TTL,
				"binding_key", bindingKey,
			)
		}
	}

	// ==========================================
	// 4. 原始工作 Queues + Bindings
	// ==========================================

	// Member Signed Up Queue
	if _, err := channel.QueueDeclare(
		commonconstants.NotificationMemberSignedUpQueue,
		true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange":    commonconstants.DlxEventsExchange,
			"x-dead-letter-routing-key": commonconstants.NotificationMemberSignedupFailed,
		},
	); err != nil {
		slog.Error("Failed to declare member signedup queue", "error", err)
		return err
	}

	if err := channel.QueueBind(
		commonconstants.NotificationMemberSignedUpQueue,
		commonconstants.MemberSignedUpEvent,
		commonconstants.AuthEventsExchange,
		false, nil,
	); err != nil {
		slog.Error("Failed to bind member signedup queue", "error", err)
		return err
	}

	// Item Created Queue
	if _, err := channel.QueueDeclare(
		commonconstants.NotificationItemCreatedQueue,
		true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange":    commonconstants.DlxEventsExchange,
			"x-dead-letter-routing-key": commonconstants.NotificationItemCreatedFailed,
		},
	); err != nil {
		slog.Error("Failed to declare item created queue", "error", err)
		return err
	}

	if err := channel.QueueBind(
		commonconstants.NotificationItemCreatedQueue,
		commonconstants.ItemCreated,
		commonconstants.ItemEventsExchange,
		false, nil,
	); err != nil {
		slog.Error("Failed to bind item created queue", "error", err)
		return err
	}

	slog.Info("Notification AMQP infrastructure setup completed")
	return nil
}

