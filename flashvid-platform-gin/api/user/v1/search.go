package v1

// SearchUsersReq 搜索用户请求
type SearchUsersReq struct {
	Keyword string `form:"keyword" binding:"required,min=1"` // 搜索关键词
	Count   int    `form:"count"   binding:"omitempty,min=1,max=20"` // 返回数量，最大20，默认6
}

// SearchUserItem 搜索用户条目
type SearchUserItem struct {
	UserID        int64  `json:"userId,string"` // 用户ID（string避免JS精度丢失）
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Bio           string `json:"bio"`
	FollowerCount int32  `json:"followerCount"`
}

// SearchUsersResp 搜索用户响应
type SearchUsersResp struct {
	Users []SearchUserItem `json:"users"`
}
