package consumer

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// getRetryCount 获取消息的重试次数（从 x-death header 中提取）
func getRetryCount(msg *amqp.Delivery) int {
	if msg.Headers == nil {
		return 0
	}
	if count, ok := msg.Headers["x-death"]; ok {
		// x-death 是一个数组，第一个元素包含 count 字段
		if deaths, ok := count.([]any); ok && len(deaths) > 0 {
			if death, ok := deaths[0].(amqp.Table); ok {
				if retryCount, ok := death["count"].(int64); ok {
					return int(retryCount)
				}
			}
		}
	}
	return 0
}
