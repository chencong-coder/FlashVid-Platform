package hotrank

import (
	"context"
	"strconv"
	"time"

	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"github.com/redis/go-redis/v9"
)

// CalculateVideoHotScore 计算视频带时间衰减的热度分数
// 公式: (播放×1 + 点赞×5 + 收藏×8 + 评论×10) / (发布天数 + 2)
func CalculateVideoHotScore(video *model.Video) float64 {
	baseScore := float64(video.ViewCount)*1 +
		float64(video.LikeCount)*5 +
		float64(video.FavoriteCount)*8 +
		float64(video.CommentCount)*10

	// 计算发布后经过的天数
	daysSincePublish := time.Since(video.PublishedAt).Hours() / 24

	// 除以 (天数 + 2)，避免除以0，给新视频缓冲期
	return baseScore / (daysSincePublish + 2)
}

// UpdateVideoHotScore 更新视频的热度分数到 Redis（带时间衰减）
func UpdateVideoHotScore(ctx context.Context, videoId int64) {
	rdb := dao.RedisClient
	if rdb == nil {
		return
	}

	// 查询视频最新数据
	video, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoId)).
		First()
	if err != nil {
		return
	}

	// 计算带时间衰减的热度分数
	score := CalculateVideoHotScore(video)

	// 更新 Redis ZSet
	rdb.ZAdd(ctx, "video:hot", redis.Z{Score: score, Member: strconv.FormatInt(videoId, 10)})
}
