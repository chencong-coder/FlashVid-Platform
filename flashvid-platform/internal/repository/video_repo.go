package repository

import (
	"flashvid-platform/internal/dao"
	"flashvid-platform/internal/model"
)

// VideoRepository 视频数据访问接口
// 用于处理复杂的视频相关查询（推荐算法、统计、联表等）
type VideoRepository interface {
	// 在这里添加复杂的自定义查询方法
	// 例如：
	// GetRecommendVideos(userID int64, limit int) ([]*model.Video, error)
	// GetHotVideos(days int, limit int) ([]*model.Video, error)
	// GetNearbyVideos(lat, lng float64, radiusKM int, limit int) ([]*model.Video, error)
}

// videoRepository 实现
type videoRepository struct{}

// NewVideoRepository 创建视频仓储实例
func NewVideoRepository() VideoRepository {
	return &videoRepository{}
}

// 示例：基于用户兴趣的推荐视频（复杂查询）
// func (r *videoRepository) GetRecommendVideos(userID int64, limit int) ([]*model.Video, error) {
// 	var videos []*model.Video
//
// 	sql := `
// 		SELECT v.*,
// 		       (v.like_count * 0.3 + v.view_count * 0.2 + v.comment_count * 0.5) as score
// 		FROM videos v
// 		LEFT JOIN video_topics vt ON v.id = vt.video_id
// 		LEFT JOIN (
// 		    SELECT DISTINCT topic_id
// 		    FROM video_topics vt2
// 		    JOIN likes l ON vt2.video_id = l.target_id
// 		    WHERE l.user_id = ? AND l.target_type = 1 AND l.deleted_at IS NULL
// 		) user_topics ON vt.topic_id = user_topics.topic_id
// 		WHERE v.status = 2
// 		  AND v.deleted_at IS NULL
// 		  AND v.user_id != ?
// 		  AND v.created_at > DATE_SUB(NOW(), INTERVAL 7 DAY)
// 		GROUP BY v.id
// 		ORDER BY user_topics.topic_id IS NOT NULL DESC, score DESC, v.published_at DESC
// 		LIMIT ?
// 	`
//
// 	err := dao.DB.Raw(sql, userID, userID, limit).Scan(&videos).Error
// 	return videos, err
// }

// 示例：附近的视频（地理位置查询）
// func (r *videoRepository) GetNearbyVideos(lat, lng float64, radiusKM int, limit int) ([]*model.Video, error) {
// 	var videos []*model.Video
//
// 	// 使用 Haversine 公式计算距离
// 	sql := `
// 		SELECT v.*,
// 		       (6371 * acos(
// 		           cos(radians(?)) * cos(radians(latitude)) *
// 		           cos(radians(longitude) - radians(?)) +
// 		           sin(radians(?)) * sin(radians(latitude))
// 		       )) AS distance
// 		FROM videos v
// 		WHERE v.latitude IS NOT NULL
// 		  AND v.longitude IS NOT NULL
// 		  AND v.status = 2
// 		  AND v.deleted_at IS NULL
// 		HAVING distance < ?
// 		ORDER BY distance ASC, v.published_at DESC
// 		LIMIT ?
// 	`
//
// 	err := dao.DB.Raw(sql, lat, lng, lat, radiusKM, limit).Scan(&videos).Error
// 	return videos, err
// }
