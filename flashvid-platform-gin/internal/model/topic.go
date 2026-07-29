package model

// 话题信息
type TopicInfo struct {
	ID 		int64  `json:"id"`          // 话题ID
	Name 	string `json:"name"`        // 话题名称
	Description string `json:"description"` // 话题描述
	CoverURL    string `json:"coverUrl"`   // 话题封面
	ViewCount   int64  `json:"viewCount"`   // 浏览量
	VideoCount  int32  `json:"videoCount"`  // 视频数
	CreatedAt   string `json:"createdAt"`   // 创建时间
}

// 话题列表输出
type TopicListOutput struct {
	Topics          []TopicInfo `json:"topics"`          // 话题列表
	NextCursorToken string      `json:"nextCursorToken"` // 下次请求传这个，空字符串表示没有更多
	HasMore         bool        `json:"hasMore"`         // 是否还有更多数据
}

// 根据话题ID获取话题下的视频列表输出
type GetTopicVideosOutput struct {
	Videos          []VideoInfo `json:"videos"`          // 视频列表
	NextCursorToken string      `json:"nextCursorToken"` // 下次请求传这个，空字符串表示没有更多
	HasMore         bool        `json:"hasMore"`         // 是否还有更多数据
}