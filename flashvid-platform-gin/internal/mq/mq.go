package mq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

var (
	conn          *amqp.Connection
	globalChannel *amqp.Channel
)

// MustInitRabbitMQ 初始化 RabbitMQ 连接和 Channel
func MustInitRabbitMQ(cfg *viper.Viper) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.GetString("rabbitmq.user"),
		cfg.GetString("rabbitmq.password"),
		cfg.GetString("rabbitmq.host"),
		cfg.GetInt("rabbitmq.port"),
	)

	var err error
	conn, err = amqp.Dial(url)
	if err != nil {
		panic(fmt.Errorf("connect rabbitmq failed: %w", err))
	}

	globalChannel, err = conn.Channel()
	if err != nil {
		panic(fmt.Errorf("open channel failed: %w", err))
	}
}

// MustDeclareInfrastructure 声明 Exchange、Queue 和 Binding
func MustDeclareInfrastructure() {
	// 0. 声明死信队列
	declareDLX()

	// 1. 事件处理相关 (Direct 模式)
	if err := globalChannel.ExchangeDeclare(
		"notification.exchange", // name
		"direct",                // type
		true,                    // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	); err != nil {
		panic(fmt.Errorf("declare notification exchange failed: %w", err))
	}

	// 1.1 热度更新队列（带死信队列配置）
	if _, err := globalChannel.QueueDeclare(
		"hotrank.update.queue", // name
		true,                   // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "dlx.exchange",
			"x-dead-letter-routing-key": "dlx",
			"x-message-ttl":             600000, // 10分钟消息过期时间
		},
	); err != nil {
		panic(fmt.Errorf("declare hotrank.update queue failed: %w", err))
	}

	if err := globalChannel.QueueBind(
		"hotrank.update.queue",  // queue name
		"hotrank",               // routing key
		"notification.exchange", // exchange
		false,
		nil,
	); err != nil {
		panic(fmt.Errorf("bind hotrank.update queue failed: %w", err))
	}

	// 1.2 通知队列（带死信队列配置）
	if _, err := globalChannel.QueueDeclare(
		"notification.queue", // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "dlx.exchange",
			"x-dead-letter-routing-key": "dlx",
			"x-message-ttl":             600000, // 10分钟消息过期时间
		},
	); err != nil {
		panic(fmt.Errorf("declare notification queue failed: %w", err))
	}

	if err := globalChannel.QueueBind(
		"notification.queue",    // queue name
		"notification",          // routing key
		"notification.exchange", // exchange
		false,
		nil,
	); err != nil {
		panic(fmt.Errorf("bind notification queue failed: %w", err))
	}

	// 2. 视频发布相关 (Fanout 广播模式)
	if err := globalChannel.ExchangeDeclare(
		"video.publish.exchange", // name
		"fanout",                 // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	); err != nil {
		panic(fmt.Errorf("declare video.publish exchange failed: %w", err))
	}

	// 2.1 热度初始化队列（带死信队列配置）
	if _, err := globalChannel.QueueDeclare(
		"video.hotrank.queue", // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "dlx.exchange",
			"x-dead-letter-routing-key": "dlx",
			"x-message-ttl":             600000, // 10分钟消息过期时间
		},
	); err != nil {
		panic(fmt.Errorf("declare video.hotrank queue failed: %w", err))
	}

	if err := globalChannel.QueueBind(
		"video.hotrank.queue",    // queue name
		"",                       // routing key (fanout 模式下为空)
		"video.publish.exchange", // exchange
		false,
		nil,
	); err != nil {
		panic(fmt.Errorf("bind video.hotrank queue failed: %w", err))
	}

	// 2.2 Feed 流推送队列（带死信队列配置）
	if _, err := globalChannel.QueueDeclare(
		"video.feed.queue", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "dlx.exchange",
			"x-dead-letter-routing-key": "dlx",
			"x-message-ttl":             600000, // 10分钟消息过期时间
		},
	); err != nil {
		panic(fmt.Errorf("declare video.feed queue failed: %w", err))
	}

	if err := globalChannel.QueueBind(
		"video.feed.queue",       // queue name
		"",                       // routing key (fanout 模式下为空)
		"video.publish.exchange", // exchange
		false,
		nil,
	); err != nil {
		panic(fmt.Errorf("bind video.feed queue failed: %w", err))
	}
}

// declareDLX 声明死信队列（Dead Letter Exchange & Queue）
func declareDLX() {
	// 1. 声明死信交换机
	if err := globalChannel.ExchangeDeclare(
		"dlx.exchange", // name
		"direct",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	); err != nil {
		panic(fmt.Errorf("declare dlx exchange failed: %w", err))
	}

	// 2. 声明死信队列
	if _, err := globalChannel.QueueDeclare(
		"dlx.queue", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	); err != nil {
		panic(fmt.Errorf("declare dlx queue failed: %w", err))
	}

	// 3. 绑定死信交换机和死信队列
	if err := globalChannel.QueueBind(
		"dlx.queue",    // queue name
		"dlx",          // routing key
		"dlx.exchange", // exchange
		false,
		nil,
	); err != nil {
		panic(fmt.Errorf("bind dlx queue failed: %w", err))
	}
}

// Close 关闭 RabbitMQ 连接
func Close() error {
	if globalChannel != nil {
		if err := globalChannel.Close(); err != nil {
			return err
		}
	}
	if conn != nil {
		if err := conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Publish 发布消息到指定 Exchange
func Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	if globalChannel == nil {
		return fmt.Errorf("rabbitmq channel not initialized")
	}

	// 发送消息（不等待 Confirm，提升性能）
	err := globalChannel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 持久化消息
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("publish message failed: %w", err)
	}

	return nil
}

// Consume 消费指定队列的消息
func Consume(queueName string) (<-chan amqp.Delivery, error) {
	return globalChannel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (设置为 false，需要手动 ACK)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
}
