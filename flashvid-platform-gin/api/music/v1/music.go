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

// CreateMusicReq 创建音乐请求（需登录，上传本地文件后调用）
type CreateMusicReq struct {
	Name     string  `json:"name"     binding:"required,min=1,max=100"`
	Artist   string  `json:"artist"   binding:"omitempty,max=100"`
	Album    string  `json:"album"    binding:"omitempty,max=100"`
	CoverURL string  `json:"coverUrl" binding:"omitempty,url"`
	MusicURL string  `json:"musicUrl" binding:"required,url"`
	Duration int32   `json:"duration" binding:"omitempty,min=0"`
}

// SearchMusicResp 搜索音乐响应
type SearchMusicResp struct {
	List     []model.MusicInfo `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// CreateMusicResp 创建音乐响应
type CreateMusicResp struct {
	Music model.MusicInfo `json:"music"`
}
