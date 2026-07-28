package v1

import (
	"flashvid-platform-gin/internal/model"
)

// 获取话题列表请求
type GetTopicsReq struct {
	Sort    string `form:"sort" binding:"omitempty,oneof=hot latest"` // 排序方式，hot表示热门，latest表示最新 如果没传，默认按热门排序
	Cursor   string `form:"cursor" binding:"omitempty"` // 游标，首次不传
    Count    int    `form:"count"    binding:"omitempty,max=50"` // 请求数量，最大50，小于10时handler层默认为10
}

// 获取话题列表响应
type GetTopicsResp struct {
	Topics          []model.TopicInfo `json:"topics"` // 话题列表
	NextCursorToken string      `json:"nextCursorToken"` // 下次请求传这个，空字符串表示没有更多
	HasMore         bool        `json:"hasMore"` // 是否还有更多数据
}