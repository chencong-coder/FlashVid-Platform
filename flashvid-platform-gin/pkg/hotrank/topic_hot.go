package hotrank

import (
	"context"
	"strconv"

	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/dao/query"
	"github.com/redis/go-redis/v9"
)

// UpdateTopicViewCount 更新话题浏览量并更新热度到 Redis
func UpdateTopicViewCount(ctx context.Context, topicId int64) {
	rdb := dao.RedisClient
	if rdb == nil {
		return
	}

	// 1. 查询话题创建时间
	topic, err := query.Topic.WithContext(ctx).
		Where(query.Topic.ID.Eq(topicId)).
		Select(query.Topic.ID, query.Topic.ViewCount, query.Topic.CreatedAt).
		First()
	if err != nil {
		return
	}

	// 2. 更新 MySQL 话题浏览量
	query.Topic.WithContext(ctx).
		Where(query.Topic.ID.Eq(topicId)).
		UpdateSimple(query.Topic.ViewCount.Add(1))

	// 3. 计算热度分数（浏览量 + 时间戳）
	// 相同浏览量的话题，最新创建的排在前面
	baseScore := float64(topic.ViewCount + 1) // +1 是因为刚才更新了
	timestamp := float64(topic.CreatedAt.Unix()) / 1e13
	finalScore := baseScore + timestamp

	// 4. 更新 Redis ZSet 热度
	rdb.ZAdd(ctx, "topic:hot", redis.Z{
		Score:  finalScore,
		Member: strconv.FormatInt(topicId, 10),
	})
}
