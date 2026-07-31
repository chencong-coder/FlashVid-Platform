package v1

import (
	"flashvid-platform-gin/internal/model"
)

// PlaylistInfo 播放列表信息
type PlaylistInfo struct {
	ID          int64  `json:"id"`          // 播放列表ID
	Title       string `json:"title"`       // 标题
	Description string `json:"description"` // 描述
	CoverUrl    string `json:"coverUrl"`    // 封面
	IsDefault   bool   `json:"isDefault"`   // 是否为默认收藏列表
	VideoCount  int32  `json:"videoCount"`  // 视频数
	CreatedAt   string `json:"createdAt"`   // 创建时间 "2006-01-02 15:04:05"
}

// GetPlaylistsReq 获取播放列表列表请求（可选指定用户，默认当前登录用户）
type GetPlaylistsReq struct {
	Page     int `form:"page" binding:"omitempty,min=1"`             // 页码，默认1
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"` // 每页数量，默认20
}

// GetPlaylistsResp 播放列表列表响应
type GetPlaylistsResp struct {
	Playlists []PlaylistInfo `json:"playlists"` // 播放列表列表
}

// CreatePlaylistReq 创建播放列表请求
type CreatePlaylistReq struct {
	Title       string `json:"title" binding:"required,max=50"` // 标题
	Description string `json:"description" binding:"omitempty,max=200"`
	CoverUrl    string `json:"coverUrl" binding:"omitempty"`
}

// CreatePlaylistResp 创建播放列表响应
type CreatePlaylistResp struct {
	Playlist PlaylistInfo `json:"playlist"`
}

// UpdatePlaylistReq 更新播放列表请求
type UpdatePlaylistReq struct {
	Title       string `json:"title" binding:"omitempty,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
	CoverUrl    string `json:"coverUrl" binding:"omitempty"`
}

// AddVideoToPlaylistReq 添加视频到播放列表请求
type AddVideoToPlaylistReq struct {
	VideoID int64 `json:"videoId" binding:"required"` // 视频ID
}

// GetPlaylistVideosReq 获取播放列表内视频请求
type GetPlaylistVideosReq struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// GetPlaylistVideosResp 获取播放列表内视频响应
type GetPlaylistVideosResp struct {
	Playlist   PlaylistInfo      `json:"playlist"`   // 播放列表信息
	Videos     []model.VideoInfo `json:"videos"`     // 视频列表
	Pagination model.Pagination  `json:"pagination"` // 分页信息
}
