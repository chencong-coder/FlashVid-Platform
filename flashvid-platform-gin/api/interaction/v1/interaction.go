package v1

// 点赞视频响应
type LikeVideoResp struct {
	IsLiked   bool  	`json:"isLiked"` // 是否已点赞
	LikeCount int32 	`json:"likeCount"` // 点赞数
}

// 收藏视频响应
type FavoriteVideoResp struct {
	IsFavorited   bool  `json:"isFavorited"`   // 是否已收藏
	FavoriteCount int32 `json:"favoriteCount"` // 收藏数
}

// 分享视频请求
type ShareVideoReq struct {
	Platform string `json:"platform" binding:"required,oneof=wechat qq weibo link"` // 分享平台
}

// 分享视频响应
type ShareVideoResp struct {
	ShareUrl   string `json:"shareUrl"`   // 分享链接
	ShareCount int32  `json:"shareCount"` // 分享数
}