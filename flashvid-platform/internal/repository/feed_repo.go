package repository

import (
	"flashvid-platform/internal/dao"
	"flashvid-platform/internal/model"
)

// FeedRepository 视频流数据访问接口
// 用于处理视频流相关的复杂查询（推荐流、关注流、附近流等）
type FeedRepository interface {
	// 在这里添加视频流相关的复杂查询方法
	// 例如：
	// GetFollowingFeed(userID, cursor int64, limit int) ([]*model.Video, error)
	// GetRecommendFeed(userID int64, cursor int64, limit int) ([]*model.Video, error)
}

// feedRepository 实现
type feedRepository struct{}

// NewFeedRepository 创建视频流仓储实例
func NewFeedRepository() FeedRepository {
	return &feedRepository{}
}

// 示例：关注流（复杂联表查询）
// func (r *feedRepository) GetFollowingFeed(userID, cursor int64, limit int) ([]*model.Video, error) {
// 	var videos []*model.Video
//
// 	sql := `
// 		SELECT DISTINCT v.*
// 		FROM videos v
// 		INNER JOIN follows f ON v.user_id = f.following_id
// 		WHERE f.follower_id = ?
// 		  AND f.deleted_at IS NULL
// 		  AND v.status = 2
// 		  AND v.deleted_at IS NULL
// 		  AND v.id < ?
// 		ORDER BY v.published_at DESC
// 		LIMIT ?
// 	`
//
// 	err := dao.DB.Raw(sql, userID, cursor, limit).Scan(&videos).Error
// 	return videos, err
// }
