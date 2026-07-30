package v1

import "flashvid-platform-gin/internal/model"

// GetMusicListReq 获取音乐列表请求
type GetMusicListReq struct {
	Sort     string `form:"sort" binding:"omitempty,oneof=hot latest"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,max=100"`
}

// GetMusicListResp 获取音乐列表响应
type GetMusicListResp struct {
	List     []model.MusicInfo `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// SearchMusicReq 搜索音乐请求
type SearchMusicReq struct {
	Keyword  string `form:"keyword" binding:"required,min=1"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,max=100"`
}

// SearchMusicResp 搜索音乐响应
type SearchMusicResp struct {
	List     []model.MusicInfo `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}
