package task

import (
	"context"
	"fmt"
	"time"

	"flashvid-platform-gin/internal/dao"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// CleanupColdVideosFromRedis 清理 Redis 中 7 天前的冷数据
func CleanupColdVideosFromRedis(cfg *viper.Viper) {
	rdb := dao.RedisClient
	if rdb == nil {
		zap.L().Warn("redis not available, skip cleanup cold videos")
		return
	}

	hotDays := cfg.GetInt("feed.hot_days")
	if hotDays == 0 {
		hotDays = 7
	}

	// 每天凌晨 3 点执行
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// 首次执行：计算到凌晨 3 点的延迟
	now := time.Now()
	next3AM := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	if now.After(next3AM) {
		next3AM = next3AM.Add(24 * time.Hour)
	}
	time.Sleep(time.Until(next3AM))

	// 立即执行一次
	cleanupColdVideos(rdb, hotDays)

	// 定时执行
	for range ticker.C {
		cleanupColdVideos(rdb, hotDays)
	}
}

func cleanupColdVideos(rdb *redis.Client, hotDays int) {
	ctx := context.Background()
	redisKey := "video:hot"

	zap.L().Info("start cleanup cold videos from redis", zap.Int("hot_days", hotDays))

	coldThreshold := time.Now().AddDate(0, 0, -hotDays).Unix()

	removedCount, err := rdb.ZRemRangeByScore(ctx, redisKey, "-inf", fmt.Sprintf("%d", coldThreshold)).Result()
	if err != nil {
		zap.L().Error("cleanup cold videos failed", zap.Error(err))
		return
	}

	zap.L().Info("cleanup cold videos completed",
		zap.Int64("removed_count", removedCount),
		zap.Int("hot_days", hotDays))
}
