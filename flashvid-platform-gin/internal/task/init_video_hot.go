package task

import (
	"context"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/pkg/hotrank"
	"github.com/redis/go-redis/v9"
	"strconv"
)

// InitVideoHot 初始化视频热榜数据到 Redis（带时间衰减）
func InitVideoHot(ctx context.Context) error {
	rdb := dao.RedisClient
	if rdb == nil {
		return nil
	}

	videos, err := query.Video.WithContext(ctx).Find()
	if err != nil {
		return err
	}

	redisKey := "video:hot"
	for _, v := range videos {
		// 使用公共函数计算带时间衰减的热度分数
		score := hotrank.CalculateVideoHotScore(v)

		rdb.ZAdd(ctx, redisKey, redis.Z{
			Score:  score,
			Member: strconv.FormatInt(v.ID, 10),
		})
	}

	return nil
}
