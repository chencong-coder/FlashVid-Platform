package consumer

import (
	"context"
	"encoding/json"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/mq"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ConsumeVideoFeed 消费视频发布事件，推送到粉丝 Feed 流
func ConsumeVideoFeed() {
	msgs, err := mq.Consume("video.feed.queue")
	if err != nil {
		zap.L().Fatal("failed to consume video.feed.queue", zap.Error(err))
	}

	zap.L().Info("video.feed consumer started")

	for msg := range msgs {
		// 解析消息
		var event mq.VideoPublishMessage
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			zap.L().Error("unmarshal message failed", zap.Error(err))
			msg.Nack(false, false) // 拒绝消息，不重新入队
			continue
		}

		// 处理消息
		if err := handleVideoFeed(event); err != nil {
			zap.L().Error("handle video feed failed",
				zap.Int64("video_id", event.VideoID),
				zap.Int64("user_id", event.UserID),
				zap.Error(err))
			msg.Nack(false, true) // 失败重新入队
			continue
		}

		// 成功确认
		msg.Ack(false)
	}
}

func handleVideoFeed(event mq.VideoPublishMessage) error {
	ctx := context.Background()

	// 1. 从 MySQL 查询该用户的所有粉丝
	db := dao.DB
	if db == nil {
		return nil // 数据库不可用，跳过
	}

	var followerIDs []int64
	err := db.WithContext(ctx).
		Table("user_follows").
		Where("followed_id = ? AND deleted_at IS NULL", event.UserID).
		Pluck("follower_id", &followerIDs).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if len(followerIDs) == 0 {
		zap.L().Debug("no followers for user", zap.Int64("user_id", event.UserID))
		return nil
	}

	// 2. 推送到每个粉丝的 Feed 流（Redis List）
	if rdb == nil {
		return nil // Redis 不可用，跳过
	}

	videoIDStr := strconv.FormatInt(event.VideoID, 10)

	for _, followerID := range followerIDs {
		feedKey := "feed:follow:" + strconv.FormatInt(followerID, 10)

		// 2.1 将视频 ID 推到 List 头部（最新的在前面）
		if err := rdb.LPush(ctx, feedKey, videoIDStr).Err(); err != nil {
			zap.L().Error("lpush feed failed",
				zap.Int64("follower_id", followerID),
				zap.Int64("video_id", event.VideoID),
				zap.Error(err))
			continue
		}

		// 2.2 限制 Feed 长度，只保留最新 1000 条
		if err := rdb.LTrim(ctx, feedKey, 0, 999).Err(); err != nil {
			zap.L().Error("ltrim feed failed",
				zap.Int64("follower_id", followerID),
				zap.Error(err))
		}

		// 2.3 设置过期时间（7 天）
		if err := rdb.Expire(ctx, feedKey, 7*24*time.Hour).Err(); err != nil {
			zap.L().Error("expire feed failed",
				zap.Int64("follower_id", followerID),
				zap.Error(err))
		}
	}

	zap.L().Info("video feed pushed",
		zap.Int64("video_id", event.VideoID),
		zap.Int64("user_id", event.UserID),
		zap.Int("follower_count", len(followerIDs)))

	return nil
}
