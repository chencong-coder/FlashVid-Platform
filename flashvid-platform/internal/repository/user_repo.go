package repository

import (
	"flashvid-platform/internal/dao"
	"flashvid-platform/internal/model"
)

// UserRepository 用户数据访问接口
// 用于处理复杂的用户相关查询
type UserRepository interface {
	// 在这里添加复杂的自定义查询方法
	// 例如：
	// GetUserStatistics(userID int64) (map[string]interface{}, error)
	// GetActiveUsers(days int, limit int) ([]*model.User, error)
}

// userRepository 实现
type userRepository struct{}

// NewUserRepository 创建用户仓储实例
func NewUserRepository() UserRepository {
	return &userRepository{}
}

// 示例：获取用户统计数据（复杂查询）
// func (r *userRepository) GetUserStatistics(userID int64) (map[string]interface{}, error) {
// 	var result map[string]interface{}
//
// 	sql := `
// 		SELECT
// 			COUNT(DISTINCT v.id) as video_count,
// 			COALESCE(SUM(v.view_count), 0) as total_views,
// 			COALESCE(SUM(v.like_count), 0) as total_likes
// 		FROM videos v
// 		WHERE v.user_id = ? AND v.deleted_at IS NULL
// 	`
//
// 	err := dao.DB.Raw(sql, userID).Scan(&result).Error
// 	return result, err
// }
