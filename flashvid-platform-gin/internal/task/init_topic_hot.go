package task

import (
	"context"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/dao/query"
	"github.com/redis/go-redis/v9"
	"strconv"
)

// InitTopicHot 初始化话题热榜数据到 Redis（简单版）
func InitTopicHot(ctx context.Context) error {
	rdb := dao.RedisClient
	if rdb == nil {
		return nil
	}

	topics, err := query.Topic.WithContext(ctx).Find()
	if err != nil {
		return err
	}

	redisKey := "topic:hot"
	for _, t := range topics {
		// 热度公式：view_count × 1 + video_count × 10
		score := float64(t.ViewCount) + float64(t.VideoCount)*10

		rdb.ZAdd(ctx, redisKey, redis.Z{
			Score:  score,
			Member: strconv.FormatInt(t.ID, 10),
		})
	}

	return nil
}
