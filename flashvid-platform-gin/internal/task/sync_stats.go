package task

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/dao/query"

	"github.com/redis/go-redis/v9"
)

// SyncVideoStatsFromRedis 定时任务：从 Redis 同步视频统计数据到 MySQL
// 每 10 秒执行一次，批量更新点赞数/播放数到数据库
func SyncVideoStatsFromRedis(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := syncOnce(ctx); err != nil {
				// 日志记录错误（这里简化处理，生产环境应接入日志库）
				fmt.Printf("[SyncVideoStats] sync failed: %v\n", err)
			}
		}
	}
}

func syncOnce(ctx context.Context) error {
	rdb := dao.RedisClient
	if rdb == nil {
		return fmt.Errorf("redis client is nil")
	}

	// 1. 扫描所有 video:*:stats 的 key（生产环境应改用 SCAN 游标遍历）
	keys, err := rdb.Keys(ctx, "video:*:stats").Result()
	if err != nil {
		return fmt.Errorf("redis keys failed: %w", err)
	}
	if len(keys) == 0 {
		return nil // 无数据，跳过
	}

	// 2. 批量读取并更新到 MySQL
	successCount := 0
	for _, key := range keys {
		// 解析 video_id：video:123:stats -> 123
		var videoID int64
		if _, err := fmt.Sscanf(key, "video:%d:stats", &videoID); err != nil {
			continue // 跳过格式不匹配的 key
		}

		// 读取 Redis 中的计数（只处理 like_count，播放数等类似逻辑）
		likeCountStr, err := rdb.HGet(ctx, key, "like_count").Result()
		if err == redis.Nil {
			continue // 该字段不存在，跳过
		}
		if err != nil {
			fmt.Printf("[SyncVideoStats] hget %s like_count failed: %v\n", key, err)
			continue
		}

		likeCount, err := strconv.ParseInt(likeCountStr, 10, 64)
		if err != nil || likeCount < 0 {
			continue
		}

		// 更新 MySQL（直接覆盖，因为 Redis 已是最新值）
		_, err = query.Video.WithContext(ctx).
			Where(query.Video.ID.Eq(videoID)).
			UpdateSimple(query.Video.LikeCount.Value(likeCount))
		if err != nil {
			fmt.Printf("[SyncVideoStats] update video %d failed: %v\n", videoID, err)
			continue
		}

		successCount++
	}

	if successCount > 0 {
		fmt.Printf("[SyncVideoStats] synced %d videos at %s\n", successCount, time.Now().Format("15:04:05"))
	}
	return nil
}
