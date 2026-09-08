package consumer

import (
	"context"
	"encoding/json"
	"flashvid-platform-gin/internal/dao/query"
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
		// 点赞/收藏事件：更新热度 + 同步 MySQL 计数
		hotrank.UpdateVideoHotScore(ctx, event.VideoID)

		// 从 Redis 读取最新计数并同步到 MySQL
		syncVideoStatsToMySQL(ctx, event.VideoID)

		zap.L().Info("video hot score updated and stats synced",
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

// syncVideoStatsToMySQL 从 Redis 同步视频统计数据到 MySQL
func syncVideoStatsToMySQL(ctx context.Context, videoID int64) {
	statsKey := fmt.Sprintf("video:%d:stats", videoID)

	// 读取 Redis 中的点赞数和收藏数
	likeCount, err := rdb.HGet(ctx, statsKey, "like_count").Int64()
	if err == nil && likeCount >= 0 {
		// 更新 MySQL 点赞数
		_, err := query.Video.WithContext(ctx).
			Where(query.Video.ID.Eq(videoID)).
			UpdateSimple(query.Video.LikeCount.Value(int32(likeCount)))
		if err != nil {
			zap.L().Error("sync like_count to mysql failed",
				zap.Int64("video_id", videoID),
				zap.Error(err))
		} else {
			zap.L().Debug("sync like_count to mysql",
				zap.Int64("video_id", videoID),
				zap.Int64("like_count", likeCount))
		}
	}

	favoriteCount, err := rdb.HGet(ctx, statsKey, "favorite_count").Int64()
	if err == nil && favoriteCount >= 0 {
		// 更新 MySQL 收藏数
		_, err := query.Video.WithContext(ctx).
			Where(query.Video.ID.Eq(videoID)).
			UpdateSimple(query.Video.FavoriteCount.Value(int32(favoriteCount)))
		if err != nil {
			zap.L().Error("sync favorite_count to mysql failed",
				zap.Int64("video_id", videoID),
				zap.Error(err))
		} else {
			zap.L().Debug("sync favorite_count to mysql",
				zap.Int64("video_id", videoID),
				zap.Int64("favorite_count", favoriteCount))
		}
	}
}

