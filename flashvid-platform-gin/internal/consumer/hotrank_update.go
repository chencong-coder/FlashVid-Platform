package consumer

import (
	"context"
	"encoding/json"
	"flashvid-platform-gin/internal/mq"
	"flashvid-platform-gin/pkg/hotrank"
	"fmt"

	"go.uber.org/zap"
)

// ConsumeHotrankUpdate 消费热度更新事件
func ConsumeHotrankUpdate() {
	msgs, err := mq.Consume("hotrank.update.queue")
	if err != nil {
		zap.L().Fatal("failed to consume hotrank.queue", zap.Error(err))
	}

	zap.L().Info("hotrank update consumer started")

	for msg := range msgs {
		// 解析消息
		var event mq.HotrankUpdateMessage
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			zap.L().Error("unmarshal message failed", zap.Error(err))
			msg.Nack(false, false) // 拒绝消息，不重新入队
			continue
		}

		// 处理消息（带重试计数）
		retryCount := getRetryCount(&msg)
		if retryCount >= 3 {
			// 重试次数超过 3 次，拒绝消息（进入死信队列）
			zap.L().Error("message retry limit exceeded, sending to DLX",
				zap.String("action", event.Action),
				zap.Int64("video_id", event.VideoID),
				zap.Int64("topic_id", event.TopicID),
				zap.Int("retry_count", retryCount))
			msg.Nack(false, false) // 不重新入队，进入死信队列
			continue
		}

		if err := handleHotrankUpdate(event); err != nil {
			zap.L().Error("handle hotrank update failed",
				zap.String("action", event.Action),
				zap.Int64("video_id", event.VideoID),
				zap.Int64("topic_id", event.TopicID),
				zap.Int("retry_count", retryCount),
				zap.Error(err))
			msg.Nack(false, true) // 失败重新入队
			continue
		}

		// 成功确认
		msg.Ack(false)
	}
}

func handleHotrankUpdate(event mq.HotrankUpdateMessage) error {
	if rdb == nil {
		return nil // Redis 不可用，跳过
	}

	ctx := context.Background()

	switch event.Action {
	case "update_video_view":
		// 播放事件：Redis 播放量 +1 + 更新热度
		statsKey := fmt.Sprintf("video:%d:stats", event.VideoID)
		rdb.HIncrBy(ctx, statsKey, "view_count", 1)

		hotrank.UpdateVideoHotScore(ctx, event.VideoID)
		zap.L().Info("video view count and hot score updated",
			zap.Int64("video_id", event.VideoID))

	case "update_video_hot":
		// 点赞/收藏事件：只更新热度（Redis 计数已在 interaction 中更新）
		hotrank.UpdateVideoHotScore(ctx, event.VideoID)
		zap.L().Info("video hot score updated",
			zap.Int64("video_id", event.VideoID))

	case "update_video_comment":
		// 评论事件：Redis 评论数 +1 + 更新热度（权重 +10）
		statsKey := fmt.Sprintf("video:%d:stats", event.VideoID)
		rdb.HIncrBy(ctx, statsKey, "comment_count", 1)

		hotrank.UpdateVideoHotScore(ctx, event.VideoID)
		zap.L().Info("video comment count and hot score updated",
			zap.Int64("video_id", event.VideoID))

	case "update_topic_view":
		// 更新话题浏览量热度
		hotrank.UpdateTopicViewCount(ctx, event.TopicID)
		zap.L().Info("topic view count updated",
			zap.Int64("topic_id", event.TopicID))

	default:
		zap.L().Warn("unknown action type", zap.String("action", event.Action))
	}

	return nil
}
