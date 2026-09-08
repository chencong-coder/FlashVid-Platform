package hotrank

import (
	"context"
	"strconv"

	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/dao/query"
	"github.com/redis/go-redis/v9"
)

// UpdateTopicViewCount 更新话题浏览量到 Redis（不直接更新 MySQL）
func UpdateTopicViewCount(ctx context.Context, topicId int64) {
	rdb := dao.RedisClient
	if rdb == nil {
		return
	}

	// 1. Redis Hash 累积浏览量
	statsKey := "topic:" + strconv.FormatInt(topicId, 10) + ":stats"

	// 先检查 Hash 是否存在，不存在则从 MySQL 初始化
	exists, _ := rdb.Exists(ctx, statsKey).Result()
	if exists == 0 {
		// 首次写入：查询 MySQL 初始化
		topic, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.Eq(topicId)).
			Select(query.Topic.ID, query.Topic.ViewCount, query.Topic.CreatedAt).
			First()
		if err != nil {
			return
		}

		// 初始化 Redis Hash
		rdb.HSet(ctx, statsKey, "view_count", topic.ViewCount+1)
		rdb.HSet(ctx, statsKey, "created_at", topic.CreatedAt.Unix())
	} else {
		// Hash 已存在：直接递增
		rdb.HIncrBy(ctx, statsKey, "view_count", 1)
	}

	// 2. 查询话题创建时间（用于计算热度）
	createdAt, err := rdb.HGet(ctx, statsKey, "created_at").Int64()
	if err != nil {
		// Redis 没有创建时间，从 MySQL 查
		topic, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.Eq(topicId)).
			Select(query.Topic.CreatedAt).
			First()
		if err != nil {
			return
		}
		createdAt = topic.CreatedAt.Unix()
		rdb.HSet(ctx, statsKey, "created_at", createdAt)
	}

	// 3. 读取最新浏览量
	viewCount, _ := rdb.HGet(ctx, statsKey, "view_count").Int64()

	// 4. 计算热度分数（浏览量 + 时间戳）
	baseScore := float64(viewCount)
	timestamp := float64(createdAt) / 1e13
	finalScore := baseScore + timestamp

	// 5. 更新 Redis ZSet 热度
	rdb.ZAdd(ctx, "topic:hot", redis.Z{
		Score:  finalScore,
		Member: strconv.FormatInt(topicId, 10),
	})
}
