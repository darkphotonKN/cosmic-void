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
type EventRetryConfig struct {
	EventType  string
	RoutingKey string
	Exchange   string
	WorkQueue  string
	DLQKey     string
}

func GetEventRetryConfigs() []EventRetryConfig {
	return []EventRetryConfig{
		{
			EventType:  "item.created",
			RoutingKey: commonconstants.ItemCreated,
			Exchange:   commonconstants.ItemEventsExchange,
			WorkQueue:  commonconstants.NotificationItemCreatedQueue,
			DLQKey:     commonconstants.NotificationItemCreatedFailed,
		},
		{
			EventType:  "member.signedup",
			RoutingKey: commonconstants.MemberSignedUpEvent,
			Exchange:   commonconstants.AuthEventsExchange,
			WorkQueue:  commonconstants.NotificationMemberSignedUpQueue,
			DLQKey:     commonconstants.NotificationMemberSignedupFailed,
		},
		{
			EventType:  "game.ended",
			RoutingKey: commonconstants.GameMatchEnded,
			Exchange:   commonconstants.GameEventsExchange,
			WorkQueue:  commonconstants.NotificationGameEndQueue,
			DLQKey:     commonconstants.NotificationGameEndFailed,
		},
	}
}

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

var MaxRetries = len(RetryLevels)

// ==========================================
// Consumer
// ==========================================

// ConsumerService defines what the consumer needs from the service
type ConsumerService interface {
	ProcessMemberSignedUp(ctx context.Context, payload *commonconstants.MemberSignedUpEventPayload) error
	ProcessItemCreated(ctx context.Context, payload *pb.ItemCreatedEvent) error
	ProcessGameEnded(ctx context.Context, payload *pb.MatchEndedEvent) error
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
	go c.consumeGameEnded()
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
// Consumer: Game Ended
// ==========================================

func (c *Consumer) consumeGameEnded() {
	msgs, err := c.channel.Consume(
		commonconstants.NotificationGameEndQueue,
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

	slog.Info("Started consuming game.ended events")

	for msg := range msgs {
		c.handleGameEnded(msg)
	}
}

func (c *Consumer) handleGameEnded(msg amqp.Delivery) {
	retryCount := getRetryCount(msg)

	var payload pb.MatchEndedEvent
	if err := proto.Unmarshal(msg.Body, &payload); err != nil {
		slog.Error("Failed to parse MatchEndedEvent", "error", err)
		msg.Nack(false, false)
		return
	}

	slog.Info("Received game ended event",
		"session_id", payload.SessionId,
		"players_count", len(payload.Players),
		"retry_count", retryCount,
	)

	ctx := context.Background()
	err := c.service.ProcessGameEnded(ctx, &payload)

	// Use session_id as identifier
	c.handleResult(msg, err, retryCount, "game.ended", payload.SessionId)
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
	// 1. DLX Exchange + DLQ
	if err := setupDLXExchangeAndQueue(channel); err != nil {
		return err
	}

	// 2. Retry Exchange
	if err := setupRetryExchange(channel); err != nil {
		return err
	}

	// 3. Business Exchanges（自動從 Event Configs 推導）
	if err := setupBusinessExchanges(channel); err != nil {
		return err
	}

	// 4. Retry Queues
	if err := setupRetryQueues(channel); err != nil {
		return err
	}

	// 5. Work Queues
	if err := setupWorkQueues(channel); err != nil {
		return err
	}

	slog.Info("✓ AMQP infrastructure ready")
	return nil
}

// ==========================================
// 1. DLX Exchange + DLQ
// ==========================================

func setupDLXExchangeAndQueue(channel *amqp.Channel) error {
	// Declare DLX Exchange
	if err := channel.ExchangeDeclare(
		commonconstants.DlxEventsExchange,
		"topic", true, false, false, false, nil,
	); err != nil {
		return err
	}

	// Declare DLQ
	if _, err := channel.QueueDeclare(
		commonconstants.NotificationDlqQueue,
		true, false, false, false, nil,
	); err != nil {
		return err
	}

	// Bind DLQ to DLX
	if err := channel.QueueBind(
		commonconstants.NotificationDlqQueue,
		"#",
		commonconstants.DlxEventsExchange,
		false, nil,
	); err != nil {
		return err
	}

	slog.Debug("✓ DLX and DLQ setup")
	return nil
}

// ==========================================
// 2. Retry Exchange
// ==========================================

func setupRetryExchange(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(
		commonconstants.RetryExchange,
		"topic", true, false, false, false, nil,
	); err != nil {
		return err
	}

	slog.Debug("✓ Retry exchange setup")
	return nil
}

// ==========================================
// 3. Business Exchanges（自動推導）
// ==========================================

func setupBusinessExchanges(channel *amqp.Channel) error {
	configs := GetEventRetryConfigs()

	// 收集唯一的 exchanges
	uniqueExchanges := make(map[string]bool)
	for _, config := range configs {
		uniqueExchanges[config.Exchange] = true
	}

	// 聲明所有唯一的 exchanges
	for exchange := range uniqueExchanges {
		if err := channel.ExchangeDeclare(
			exchange,
			"topic", true, false, false, false, nil,
		); err != nil {
			slog.Error("Failed to declare business exchange",
				"exchange", exchange,
				"error", err,
			)
			return err
		}
		slog.Debug("✓ Business exchange", "name", exchange)
	}

	return nil
}

// ==========================================
// 4. Retry Queues
// ==========================================

func setupRetryQueues(channel *amqp.Channel) error {
	configs := GetEventRetryConfigs()

	for _, config := range configs {
		for _, retry := range RetryLevels {
			queueName := "retry." + retry.Level + "." + config.EventType
			bindingKey := retry.Level + "." + config.RoutingKey

			// Declare Retry Queue
			if _, err := channel.QueueDeclare(
				queueName,
				true, false, false, false,
				amqp.Table{
					"x-message-ttl":             retry.TTL,
					"x-dead-letter-exchange":    config.Exchange,
					"x-dead-letter-routing-key": config.RoutingKey,
				},
			); err != nil {
				return err
			}

			// Bind to Retry Exchange
			if err := channel.QueueBind(
				queueName,
				bindingKey,
				commonconstants.RetryExchange,
				false, nil,
			); err != nil {
				return err
			}

			slog.Debug("✓ Retry queue",
				"name", queueName,
				"ttl_seconds", retry.DelaySeconds,
			)
		}
	}

	return nil
}

// ==========================================
// 5. Work Queues
// ==========================================

func setupWorkQueues(channel *amqp.Channel) error {
	configs := GetEventRetryConfigs()

	for _, config := range configs {
		// Declare Work Queue
		if _, err := channel.QueueDeclare(
			config.WorkQueue,
			true, false, false, false,
			amqp.Table{
				"x-dead-letter-exchange":    commonconstants.DlxEventsExchange,
				"x-dead-letter-routing-key": config.DLQKey,
			},
		); err != nil {
			return err
		}

		// Bind to Business Exchange
		if err := channel.QueueBind(
			config.WorkQueue,
			config.RoutingKey,
			config.Exchange,
			false, nil,
		); err != nil {
			return err
		}

		slog.Debug("✓ Work queue", "name", config.WorkQueue)
	}

	return nil
}
