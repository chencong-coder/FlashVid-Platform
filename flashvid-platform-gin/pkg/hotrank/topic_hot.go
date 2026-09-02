package hotrank

import (
	"context"
	"strconv"

	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/dao/query"
)

// UpdateTopicViewCount 更新话题浏览量并更新热度到 Redis
func UpdateTopicViewCount(ctx context.Context, topicId int64) {
	rdb := dao.RedisClient
	if rdb == nil {
		return
	}

	// 1. 更新 MySQL 话题浏览量
	query.Topic.WithContext(ctx).
		Where(query.Topic.ID.Eq(topicId)).
		UpdateSimple(query.Topic.ViewCount.Add(1))

	// 2. 更新 Redis ZSet 热度 (+1)
	rdb.ZIncrBy(ctx, "topic:hot", 1, strconv.FormatInt(topicId, 10))
}
