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