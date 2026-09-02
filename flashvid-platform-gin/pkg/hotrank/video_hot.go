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

	// 查询视频基础信息（published_at）
	video, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoId)).
		Select(query.Video.ID, query.Video.PublishedAt).
		First()
	if err != nil {
		return
	}

	// 从 Redis 读取实时统计数据（优先级高于 MySQL）
	statsKey := "video:" + strconv.FormatInt(videoId, 10) + ":stats"
	statsMap := rdb.HGetAll(ctx, statsKey).Val()

	// 解析 Redis 中的统计数据
	if len(statsMap) > 0 {
		// Redis 有数据，用实时数据
		video.ViewCount = int32(parseInt64(statsMap["view_count"]))
		video.LikeCount = int32(parseInt64(statsMap["like_count"]))
		video.FavoriteCount = int32(parseInt64(statsMap["favorite_count"]))
		video.CommentCount = int32(parseInt64(statsMap["comment_count"]))
	} else {
		// Redis 无数据，回退到 MySQL（新视频或缓存失效）
		fullVideo, err := query.Video.WithContext(ctx).
			Where(query.Video.ID.Eq(videoId)).
			First()
		if err != nil {
			return
		}
		video = fullVideo
	}

	// 计算带时间衰减的热度分数
	score := CalculateVideoHotScore(video)

	// 更新 Redis ZSet
	rdb.ZAdd(ctx, "video:hot", redis.Z{Score: score, Member: strconv.FormatInt(videoId, 10)})
}

// parseInt64 安全解析字符串为 int64
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}
