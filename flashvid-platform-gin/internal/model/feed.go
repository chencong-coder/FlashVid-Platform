package model

// 视频流output
type FeedOutput struct {
	Videos  []VideoInfo `json:"video"` // 视频详情
	NextCursorToken string      `json:"nextCursorToken"` // 下次请求传这个，空字符串表示没有更多
	HasMore         bool        `json:"hasMore"` // 是否还有更多数据
}