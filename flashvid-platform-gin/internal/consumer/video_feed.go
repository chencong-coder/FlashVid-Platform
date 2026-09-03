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

		// 处理消息（带重试计数）
		retryCount := getRetryCount(&msg)
		if retryCount >= 3 {
			zap.L().Error("message retry limit exceeded, sending to DLX",
				zap.Int64("video_id", event.VideoID),
				zap.Int64("user_id", event.UserID),
				zap.Int("retry_count", retryCount))
			msg.Nack(false, false) // 不重新入队，进入死信队列
			continue
		}

		if err := handleVideoFeed(event); err != nil {
			zap.L().Error("handle video feed failed",
				zap.Int64("video_id", event.VideoID),
				zap.Int64("user_id", event.UserID),
				zap.Int("retry_count", retryCount),
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

	// 2. 查询该用户关注的所有人（用于计算互相关注）
	var followingIDs []int64
	err = db.WithContext(ctx).
		Table("user_follows").
		Where("follower_id = ? AND deleted_at IS NULL", event.UserID).
		Pluck("followed_id", &followingIDs).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// 3. 计算互相关注的用户（朋友）
	followingSet := make(map[int64]struct{}, len(followingIDs))
	for _, id := range followingIDs {
		followingSet[id] = struct{}{}
	}
	var friendIDs []int64
	for _, followerID := range followerIDs {
		if _, ok := followingSet[followerID]; ok {
			friendIDs = append(friendIDs, followerID)
		}
	}

	// 4. 大V判断：粉丝 >= 10000，跳过推模式（走拉模式）
	if len(followerIDs) >= 10000 {
		zap.L().Info("skip push mode for big V",
			zap.Int64("user_id", event.UserID),
			zap.Int("follower_count", len(followerIDs)))
		return nil
	}

	// 5. 推送到每个粉丝的 Feed 流（Redis List）
	if rdb == nil {
		return nil // Redis 不可用，跳过
	}

	videoIDStr := strconv.FormatInt(event.VideoID, 10)

	// 5.1 推送到关注流（所有粉丝）
	for _, followerID := range followerIDs {
		feedKey := "feed:follow:" + strconv.FormatInt(followerID, 10)

		if err := rdb.LPush(ctx, feedKey, videoIDStr).Err(); err != nil {
			zap.L().Error("lpush follow feed failed",
				zap.Int64("follower_id", followerID),
				zap.Int64("video_id", event.VideoID),
				zap.Error(err))
			continue
		}

		if err := rdb.LTrim(ctx, feedKey, 0, 999).Err(); err != nil {
			zap.L().Error("ltrim follow feed failed",
				zap.Int64("follower_id", followerID),
				zap.Error(err))
		}

		if err := rdb.Expire(ctx, feedKey, 7*24*time.Hour).Err(); err != nil {
			zap.L().Error("expire follow feed failed",
				zap.Int64("follower_id", followerID),
				zap.Error(err))
		}
	}

	// 5.2 推送到朋友流（互相关注的用户）
	for _, friendID := range friendIDs {
		friendFeedKey := "feed:friends:" + strconv.FormatInt(friendID, 10)

		if err := rdb.LPush(ctx, friendFeedKey, videoIDStr).Err(); err != nil {
			zap.L().Error("lpush friends feed failed",
				zap.Int64("friend_id", friendID),
				zap.Int64("video_id", event.VideoID),
				zap.Error(err))
			continue
		}

		if err := rdb.LTrim(ctx, friendFeedKey, 0, 999).Err(); err != nil {
			zap.L().Error("ltrim friends feed failed",
				zap.Int64("friend_id", friendID),
				zap.Error(err))
		}

		if err := rdb.Expire(ctx, friendFeedKey, 7*24*time.Hour).Err(); err != nil {
			zap.L().Error("expire friends feed failed",
				zap.Int64("friend_id", friendID),
				zap.Error(err))
		}
	}

	zap.L().Info("video feed pushed",
		zap.Int64("video_id", event.VideoID),
		zap.Int64("user_id", event.UserID),
		zap.Int("follower_count", len(followerIDs)),
		zap.Int("friend_count", len(friendIDs)))

	return nil
}
