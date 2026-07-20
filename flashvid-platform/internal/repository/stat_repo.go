package repository

import (
	"flashvid-platform/internal/dao"
)

// StatRepository 统计数据访问接口
// 用于处理各种统计查询（热门话题、用户排行、数据报表等）
type StatRepository interface {
	// 在这里添加统计相关的复杂查询方法
	// 例如：
	// GetHotTopics(days int, limit int) ([]map[string]interface{}, error)
	// GetTopUsers(orderBy string, limit int) ([]map[string]interface{}, error)
	// GetDailyStats(startDate, endDate string) ([]map[string]interface{}, error)
}

// statRepository 实现
type statRepository struct{}

// NewStatRepository 创建统计仓储实例
func NewStatRepository() StatRepository {
	return &statRepository{}
}

// 示例：获取热门话题（聚合查询）
// func (r *statRepository) GetHotTopics(days int, limit int) ([]map[string]interface{}, error) {
// 	var results []map[string]interface{}
//
// 	sql := `
// 		SELECT t.id, t.name, t.cover_url,
// 		       COUNT(DISTINCT v.id) as video_count,
// 		       SUM(v.view_count) as total_views,
// 		       SUM(v.like_count) as total_likes
// 		FROM topics t
// 		JOIN video_topics vt ON t.id = vt.topic_id
// 		JOIN videos v ON vt.video_id = v.id
// 		WHERE v.created_at > DATE_SUB(NOW(), INTERVAL ? DAY)
// 		  AND v.status = 2
// 		  AND v.deleted_at IS NULL
// 		GROUP BY t.id, t.name, t.cover_url
// 		HAVING video_count > 5
// 		ORDER BY total_views DESC
// 		LIMIT ?
// 	`
//
// 	err := dao.DB.Raw(sql, days, limit).Scan(&results).Error
// 	return results, err
// }
