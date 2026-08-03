package model

// 播放列表信息
type PlayListInfo struct {
	ID 		int64  `json:"id"`           // 播放列表ID
	Title 	string `json:"title"`        // 播放列表标题
	Description string `json:"description"` // 播放列表描述
	CoverURL    string `json:"cover_url"`   // 播放列表封面URL
	VideoCount  int32  `json:"video_count"` // 播放列表视频数量
	CreatedAt   string `json:"created_at"`   // 播放列表创建时间
}