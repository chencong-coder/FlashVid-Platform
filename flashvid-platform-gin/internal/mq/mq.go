package mq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

var (
	conn    *amqp.Connection
	channel *amqp.Channel
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

	channel, err = conn.Channel()
	if err != nil {
		panic(fmt.Errorf("open channel failed: %w", err))
	}
}

// MustDeclareInfrastructure 声明 Exchange、Queue 和 Binding
func MustDeclareInfrastructure() {
	// 1. 通知相关 (Direct 模式)
	if err := channel.ExchangeDeclare(
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

	if _, err := channel.QueueDeclare(
		"notification.queue", // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	); err != nil {
		panic(fmt.Errorf("declare notification queue failed: %w", err))
	}

	if err := channel.QueueBind(
		"notification.queue",    // queue name
		"notification",          // routing key
		"notification.exchange", // exchange
		false,
		nil,
	); err != nil {
		panic(fmt.Errorf("bind notification queue failed: %w", err))
	}

	// 2. 视频发布相关 (Fanout 广播模式)
	if err := channel.ExchangeDeclare(
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

	// 2.1 热度初始化队列
	if _, err := channel.QueueDeclare(
		"video.hotrank.queue", // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	); err != nil {
		panic(fmt.Errorf("declare video.hotrank queue failed: %w", err))
	}

	if err := channel.QueueBind(
		"video.hotrank.queue",    // queue name
		"",                       // routing key (fanout 模式下为空)
		"video.publish.exchange", // exchange
		false,
		nil,
	); err != nil {
		panic(fmt.Errorf("bind video.hotrank queue failed: %w", err))
	}

	// 2.2 Feed 流推送队列
	if _, err := channel.QueueDeclare(
		"video.feed.queue", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	); err != nil {
		panic(fmt.Errorf("declare video.feed queue failed: %w", err))
	}

	if err := channel.QueueBind(
		"video.feed.queue",       // queue name
		"",                       // routing key (fanout 模式下为空)
		"video.publish.exchange", // exchange
		false,
		nil,
	); err != nil {
		panic(fmt.Errorf("bind video.feed queue failed: %w", err))
	}
}

// Close 关闭 RabbitMQ 连接
func Close() error {
	if channel != nil {
		if err := channel.Close(); err != nil {
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
	return channel.PublishWithContext(
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
}

// Consume 消费指定队列的消息
func Consume(queueName string) (<-chan amqp.Delivery, error) {
	return channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (设置为 false，需要手动 ACK)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
}
