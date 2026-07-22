package v1

import (
	"flashvid-platform-gin/internal/model"
)

// 获取用户作品列表请求（分页参数）
type GetUserVideosReq struct {
	Page     int `form:"page" binding:"omitempty,min=1"`              // 页码，默认1
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"` // 每页数量，默认20
}

// 获取用户作品列表响应
type GetUserVideosResp struct {
	Videos      [] model.VideoInfo `json:"videos"`       // 视频列表
	Pagination model.Pagination  `json:"pagination"` // 分页信息
}