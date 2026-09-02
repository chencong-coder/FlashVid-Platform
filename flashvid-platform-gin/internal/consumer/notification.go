package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"flashvid-platform-gin/internal/mq"
	"go.uber.org/zap"
)

// ConsumeNotification 消费通知创建事件
func ConsumeNotification() {
	msgs, err := mq.Consume("notification.queue")
	if err != nil {
		zap.L().Fatal("failed to consume notification.queue", zap.Error(err))
	}

	zap.L().Info("notification consumer started")

	for msg := range msgs {
		// 解析消息
		var event mq.NotificationMessage
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			zap.L().Error("unmarshal notification message failed", zap.Error(err))
			msg.Nack(false, false) // 格式错误，拒绝消息不重新入队
			continue
		}

		// 处理消息（带重试计数）
		retryCount := getRetryCount(&msg)
		if retryCount >= 3 {
			zap.L().Error("message retry limit exceeded, sending to DLX",
				zap.Int64("user_id", event.UserID),
				zap.Int64("actor_id", event.ActorID),
				zap.Int32("action_type", event.ActionType),
				zap.Int("retry_count", retryCount))
			msg.Nack(false, false) // 不重新入队，进入死信队列
			continue
		}

		if err := handleNotificationCreate(event); err != nil {
			zap.L().Error("handle notification create failed",
				zap.Int64("user_id", event.UserID),
				zap.Int64("actor_id", event.ActorID),
				zap.Int32("action_type", event.ActionType),
				zap.Int("retry_count", retryCount),
				zap.Error(err))
			msg.Nack(false, true) // 失败重新入队
			continue
		}

		// 成功确认
		msg.Ack(false)
	}
}

func handleNotificationCreate(event mq.NotificationMessage) error {
	ctx := context.Background()

	// 创建通知记录
	notification := &model.Notification{
		UserID:     event.UserID,
		ActorID:    event.ActorID,
		ActionType: event.ActionType,
		TargetType: event.TargetType,
		TargetID:   event.TargetID,
	}

	if err := query.Q.Notification.WithContext(ctx).Create(notification); err != nil {
		return fmt.Errorf("create notification failed: %w", err)
	}

	// 更新 Redis 未读数（如果 Redis 可用）
	if rdb != nil {
		key := fmt.Sprintf("user:%d:unread", event.UserID)
		field := fmt.Sprintf("%d", event.ActionType)
		if err := rdb.HIncrBy(ctx, key, field, 1).Err(); err != nil {
			zap.L().Warn("update redis unread count failed",
				zap.Int64("user_id", event.UserID),
				zap.Error(err))
			// Redis 失败不影响主流程，继续
		}
	}

	zap.L().Info("notification created",
		zap.Int64("user_id", event.UserID),
		zap.Int64("actor_id", event.ActorID),
		zap.Int32("action_type", event.ActionType))

	return nil
}
