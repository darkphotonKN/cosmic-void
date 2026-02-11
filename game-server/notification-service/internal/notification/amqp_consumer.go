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

// ConsumerService defines what the consumer needs from the service
// This interface specifies methods that will be called when processing events
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

func (c *Consumer) consumeItemCreated() {
	msgs, err := c.channel.Consume(
		commonconstants.NotificationItemCreatedQueue, // queue name
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
		retryCount := getRetryCount(msg)
		maxRetries := 3
		slog.Info("Start ItemCreatedEvent")
		var payload pb.ItemCreatedEvent
		if err := proto.Unmarshal(msg.Body, &payload); err != nil {
			slog.Error("Failed to parse event", "error", err)
			msg.Nack(false, false) // RabbitMQ：處理失敗，不要重試
			continue
		}
		slog.Info("Received member signed up event",
			"item_type", payload.ItemType,
			"name", payload.Name,
		)

		ctx := context.Background()

		err := c.service.ProcessItemCreated(ctx, &payload)

		if errors.Is(err, commonconstants.ErrTransient) {
			// 暫時性錯誤（如資料庫連線失敗）
			if retryCount < maxRetries {
				slog.Warn("Transient error, requeue",
					"retry", retryCount,
					"max", maxRetries)

				// 增加重試計數並重新發送
				if err := c.requeueWithRetry(msg, retryCount+1); err != nil {
					slog.Error("Failed to requeue message", "error", err)
					msg.Nack(false, false) // 重新排隊失敗，進入 DLQ
				} else {
					msg.Ack(false) // 原消息 ACK（因為已經重新發送了新的）
				}
			} else {
				slog.Error("Max retries exceeded, sending to DLQ")
				msg.Nack(false, false) // 重試次數用盡 -> DLQ
			}
			continue
		} else if err != nil {
			// 永久性錯誤（如數據格式錯誤）
			slog.Error("Permanent error, sending to DLQ", "error", err)
			msg.Nack(false, false) // 直接進 DLQ
			continue
		}

		// successfully processed (ACK)
		msg.Ack(false)
		slog.Info("Item Created notification created successfully",
			"user_id", payload.UserId,
			"name", payload.Name,
			"type_id", payload.ItemType,
		)
	}
}

func (c *Consumer) consumeMemberSignedUp() {
	msgs, err := c.channel.Consume(
		commonconstants.NotificationMemberSignedUpQueue, // queue name
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
		retryCount := getRetryCount(msg)
		maxRetries := 3

		var payload commonconstants.MemberSignedUpEventPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			slog.Error("Failed to parse event", "error", err)
			msg.Nack(false, false) // 解析失敗 -> DLQ
			continue
		}
		slog.Info("Received member signed up event",
			"user_id", payload.UserID,
			"name", payload.Name,
		)

		ctx := context.Background()
		err := c.service.ProcessMemberSignedUp(ctx, &payload)

		if errors.Is(err, commonconstants.ErrTransient) {
			// 暫時性錯誤
			if retryCount < maxRetries {
				slog.Warn("Transient error, requeue",
					"retry", retryCount,
					"max", maxRetries)

				if err := c.requeueWithRetry(msg, retryCount+1); err != nil {
					slog.Error("Failed to requeue message", "error", err)
					msg.Nack(false, false) // 重新排隊失敗，進入 DLQ
				} else {
					msg.Ack(false) // 原消息 ACK
				}
			} else {
				slog.Error("Max retries exceeded, sending to DLQ")
				msg.Nack(false, false) // 重試次數用盡 -> DLQ
			}
			continue
		} else if err != nil {
			// 永久性錯誤
			slog.Error("Permanent error, sending to DLQ", "error", err)
			msg.Nack(false, false) // 直接進 DLQ
			continue
		}

		// successfully processed (ACK)
		msg.Ack(false)
		slog.Info("Member signed up notification created successfully",
			"user_id", payload.UserID,
		)
	}

}

func SetupAMQPInfrastructure(channel *amqp.Channel) error {
	// Dead Letter Exchange 和 DLQ
	err := channel.ExchangeDeclare(
		commonconstants.DlxEventsExchange, // DLX 名稱
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	_, err = channel.QueueDeclare(
		commonconstants.NotificationDlqQueue, // DLQ 名稱
		true,                                 // durable
		false,                                // delete when unused
		false,                                // exclusive
		false,                                // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	err = channel.QueueBind(
		commonconstants.NotificationDlqQueue, // queue
		"#",                                  // routing key (接收所有消息)
		commonconstants.DlxEventsExchange,    // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Declare auth.events exchange
	err = channel.ExchangeDeclare(
		commonconstants.AuthEventsExchange, // exchange name
		"topic",                            // exchange type
		true,                               // durable
		false,                              // auto-deleted
		false,                              // internal
		false,                              // no-wait
		nil,                                // arguments
	)
	if err != nil {
		slog.Error("Failed to declare auth exchange", "error", err)
		return err
	}

	// Declare item.events exchange
	err = channel.ExchangeDeclare(
		commonconstants.ItemEventsExchange, // exchange name
		"topic",                            // exchange type
		true,                               // durable
		false,                              // auto-deleted
		false,                              // internal
		false,                              // no-wait
		nil,                                // arguments
	)
	if err != nil {
		slog.Error("Failed to declare item exchange", "error", err)
		return err
	}

	// Declare notification.auth.member.signedup queue
	_, err = channel.QueueDeclare(
		commonconstants.NotificationMemberSignedUpQueue, // queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			// 重點：設定此 Queue 的 Dead Letter Exchange
			"x-dead-letter-exchange":    commonconstants.DlxEventsExchange,
			"x-dead-letter-routing-key": commonconstants.NotificationMemberSignedupFailed, // 使用正確的 routing key
		},
	)
	if err != nil {
		slog.Error("Failed to declare member signedup queue", "error", err)
		return err
	}

	// Bind member.signedup queue to auth.events exchange
	err = channel.QueueBind(
		commonconstants.NotificationMemberSignedUpQueue, // queue name
		commonconstants.MemberSignedUpEvent,             // routing key
		commonconstants.AuthEventsExchange,              // exchange
		false,                                           // no-wait
		nil,                                             // args
	)
	if err != nil {
		slog.Error("Failed to bind member signedup queue", "error", err)
		return err
	}

	// Declare notification.item.item.created queue
	_, err = channel.QueueDeclare(
		commonconstants.NotificationItemCreatedQueue, // queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			// 設定此 Queue 的 Dead Letter Exchange
			"x-dead-letter-exchange":    commonconstants.DlxEventsExchange,
			"x-dead-letter-routing-key": commonconstants.NotificationItemCreatedFailed,
		},
	)
	if err != nil {
		slog.Error("Failed to declare item created queue", "error", err)
		return err
	}

	// Bind item.created queue to item.events exchange
	err = channel.QueueBind(
		commonconstants.NotificationItemCreatedQueue, // queue name
		commonconstants.ItemCreated,                  // routing key
		commonconstants.ItemEventsExchange,           // exchange
		false,                                        // no-wait
		nil,                                          // args
	)
	if err != nil {
		slog.Error("Failed to bind item created queue", "error", err)
		return err
	}

	slog.Info("Notification AMQP infrastructure setup completed")
	return nil
}

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

// requeueWithRetry 重新發送消息，並增加重試計數
func (c *Consumer) requeueWithRetry(msg amqp.Delivery, newRetryCount int) error {
	// 複製消息的 headers
	headers := amqp.Table{}
	if msg.Headers != nil {
		for k, v := range msg.Headers {
			headers[k] = v
		}
	}

	// 更新重試計數
	headers["x-retry-count"] = int32(newRetryCount)
	headers["x-original-routing-key"] = msg.RoutingKey

	// 重新發送到原始 exchange
	err := c.channel.Publish(
		msg.Exchange,   // exchange
		msg.RoutingKey, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  msg.ContentType,
			Body:         msg.Body,
			DeliveryMode: msg.DeliveryMode,
			Priority:     msg.Priority,
			Headers:      headers,
		},
	)

	if err != nil {
		return err
	}

	slog.Info("Message requeued with incremented retry count",
		"retry_count", newRetryCount,
		"routing_key", msg.RoutingKey)

	return nil
}
