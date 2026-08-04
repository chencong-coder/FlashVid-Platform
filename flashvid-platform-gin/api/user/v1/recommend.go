package v1

// RecommendUserItem 推荐用户条目
type RecommendUserItem struct {
	UserID        int64  `json:"userId,string"` // 用户ID（string序列化，避免JS精度丢失）
	Nickname      string `json:"nickname"`      // 昵称
	Avatar        string `json:"avatar"`        // 头像
	Bio           string `json:"bio"`           // 个人简介
	FollowerCount int32  `json:"followerCount"` // 粉丝数
}

// GetRecommendUsersResp 推荐用户响应
type GetRecommendUsersResp struct {
	Users []RecommendUserItem `json:"users"` // 推荐用户列表
}
