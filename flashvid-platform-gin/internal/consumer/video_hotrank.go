package consumer

import (
	"context"
	"encoding/json"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/mq"
	"strconv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var rdb = dao.RedisClient

// ConsumeVideoHotrank 消费视频发布事件，初始化热度
func ConsumeVideoHotrank() {
	msgs, err := mq.Consume("video.hotrank.queue")
	if err != nil {
		zap.L().Fatal("failed to consume video.hotrank.queue", zap.Error(err))
	}

	zap.L().Info("video.hotrank consumer started")

	for msg := range msgs {
		// 解析消息
		var event mq.VideoPublishMessage
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			zap.L().Error("unmarshal message failed", zap.Error(err))
			msg.Nack(false, false) // 拒绝消息，不重新入队
			continue
		}

		// 处理消息（带重试计数）
		retryCount := getRetryCount(&msg)
		if retryCount >= 3 {
			zap.L().Error("message retry limit exceeded, sending to DLX",
				zap.Int64("video_id", event.VideoID),
				zap.Int("retry_count", retryCount))
			msg.Nack(false, false) // 不重新入队，进入死信队列
			continue
		}

		if err := handleVideoHotrank(event); err != nil {
			zap.L().Error("handle video hotrank failed",
				zap.Int64("video_id", event.VideoID),
				zap.Int("retry_count", retryCount),
				zap.Error(err))
			msg.Nack(false, true) // 失败重新入队
			continue
		}

		// 成功确认
		msg.Ack(false)
	}
}

func handleVideoHotrank(event mq.VideoPublishMessage) error {
	if rdb == nil {
		return nil // Redis 不可用，跳过
	}

	ctx := context.Background()

	// 1. 初始化视频热度到 Redis ZSet（初始分数为 0）
	videoKey := "video:hot"
	if err := rdb.ZAdd(ctx, videoKey, redis.Z{
		Score:  0,
		Member: strconv.FormatInt(event.VideoID, 10),
	}).Err(); err != nil {
		return err
	}

	// 2. 处理话题热度
	if len(event.TopicIDs) > 0 {
		topicKey := "topic:hot"
		for _, topicID := range event.TopicIDs {
			// 2.1 检查话题是否已存在
			_, err := rdb.ZScore(ctx, topicKey, strconv.FormatInt(topicID, 10)).Result()
			if err == redis.Nil {
				// 话题不存在，初始化为 0
				rdb.ZAdd(ctx, topicKey, redis.Z{
					Score:  0,
					Member: strconv.FormatInt(topicID, 10),
				})
			} else if err != nil {
				return err
			}

			// 2.2 更新话题热度（video_count 权重 × 10）
			if err := rdb.ZIncrBy(ctx, topicKey, 10, strconv.FormatInt(topicID, 10)).Err(); err != nil {
				return err
			}
		}
	}

	zap.L().Info("video hotrank initialized",
		zap.Int64("video_id", event.VideoID),
		zap.Int64s("topic_ids", event.TopicIDs))

	return nil
}
