package v1

// GetNotificationsReq 获取通知列表请求
type GetNotificationsReq struct {
	ActionTypes []int32 `form:"action_types"` // 按操作类型筛选，空=全部
	Page        int     `form:"page"          binding:"required,min=1"`
	PageSize    int     `form:"page_size"     binding:"required,min=1,max=100"`
}

// NotificationInfo 单条通知 DTO
type NotificationInfo struct {
	ID          int64  `json:"id"`
	ActorID     int64  `json:"actorId"`
	ActorName   string `json:"actorName"`
	ActorAvatar string `json:"actorAvatar"`
	ActionType  int32  `json:"actionType"` // 1=关注 2=点赞 3=收藏 4=评论 5=回复
	TargetID    int64  `json:"targetId"`
	TargetTitle string `json:"targetTitle"` // 视频标题（点赞/收藏/评论时）
	TargetCover string `json:"targetCover"` // 视频封面
	Content     string `json:"content"`     // 评论内容预览
	IsRead      int32  `json:"isRead"`
	CreatedAt   string `json:"createdAt"`
}

// GetNotificationsResp 获取通知列表响应
type GetNotificationsResp struct {
	List  []NotificationInfo `json:"list"`
	Total int64              `json:"total"`
}

// UnreadCountsResp 各类型未读数响应
type UnreadCountsResp struct {
	Followers    int64 `json:"followers"`    // 粉丝（action_type=1）
	LikesAndFavs int64 `json:"likesAndFavs"` // 赞和收藏（action_type=2,3）
	Mentions     int64 `json:"mentions"`     // @我的（action_type=6）
	Comments     int64 `json:"comments"`     // 评论+回复（action_type=4,5）
}

// MarkAsReadReq 标记已读请求
type MarkAsReadReq struct {
	ActionTypes []int32 `json:"actionTypes"` // 空=标记所有，非空=标记指定类型
}
