package v1

import (
	"flashvid-platform-gin/internal/model"
)

// 推荐流请求
type RecommendFeedReq struct {
    Cursor   string `form:"cursor" binding:"omitempty"` // 游标，首次不传
    Count    int    `form:"count"    binding:"omitempty,min=1,max=50"` // 请求数量，最小1，最大50
}

// 关注流请求
type FollowFeedReq struct {
    Cursor   string `form:"cursor" binding:"omitempty"` // 游标，首次不传
    Count    int    `form:"count"    binding:"omitempty,min=1,max=50"` // 请求数量，最小1，最大50
}

// 附近流请求
type NearbyFeedReq struct {
    Latitude  float64 `form:"latitude"  binding:"required，gte=-90,lte=90"` // 纬度
    Longitude float64 `form:"longitude" binding:"required，gt=-180,lte=180"` // 经度
    Distance  int`form:"distance"  binding:"omitempty,min=1,max=100"` // 距离，单位米，最小1，最大100
    Cursor    string  `form:"cursor"    binding:"omitempty"` // 游标，首次不传
    Count     int     `form:"count"     binding:"omitempty,min=1,max=50"` // 请求数量，最小1，最大50
}

// 视频流响应（3个接口共用）
type FeedResp struct {
    Videos 	[]model.VideoInfo `json:"videos"` // 视频列表
    NextCursorToken string            `json:"nextCursorToken"` // 下次请求传这个，空字符串表示没有更多
    HasMore         bool              `json:"hasMore"` // 是否还有更多数据
}
