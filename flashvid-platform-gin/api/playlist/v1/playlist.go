package v1

import (
	"flashvid-platform-gin/internal/model"
)

// 获取播放列表响应
type GetPlayListsResp struct {
	PlayLists []model.PlayListInfo `json:"playlists"` // 播放列表信息
}

// 创建播放列表请求
type CreatePlaylistReq struct {
	Title       string `json:"title" binding:"required"` // 播放列表标题
	Description string `json:"description"`              // 播放列表描述
	CoverURL    string `json:"cover_url"`                // 播放列表封面URL
}

// 创建播放列表响应
type CreatePlaylistResp struct {
	Playlist model.PlayListInfo `json:"playlist"` // 播放列表信息
}

// 更新播放列表请求
type UpdatePlaylistReq struct {
	ID          int64   `uri:"id"  binding:"required"` // 播放列表ID（路径参数）
	Title       *string `json:"title"`                 // 播放列表标题
	Description *string `json:"description"`           // 播放列表描述
	CoverURL    *string `json:"cover_url"`             // 播放列表封面URL
}

// 删除播放列表请求
type DeletePlaylistReq struct {
	ID int64 `uri:"id" binding:"required"` // 播放列表ID（路径参数）
}

// 获取播放列表内的视频请求
type GetPlaylistVideosReq struct {
	ID     int64  `uri:"id" binding:"required"` // 播放列表ID（路径参数）
	Cursor   string `form:"cursor" binding:"omitempty"` // 游标，首次不传
    Count    int    `form:"count"    binding:"omitempty,max=50"` // 请求数量，最大50，小于10时handler层默认为10
}

// 获取播放列表内的视频响应
type GetPlaylistVideosResp struct {
	Playlist   model.PlayListInfo `json:"playlist"`   // 播放列表信息
	Videos     []model.VideoInfo  `json:"videos"`     // 视频列表
	NextCursor string             `json:"nextCursor"` // 下一页游标，空字符串表示没有更多
	HasMore    bool               `json:"hasMore"`    // 是否还有更多数据
}

// 添加视频到播放列表请求
type AddVideoToPlaylistReq struct {
	PlaylistID int64 `uri:"id" binding:"required"`      // 播放列表ID（路径参数）
	VideoID    int64 `json:"videoId" binding:"required"` // 视频ID
}

// 从播放列表移除视频请求
type RemoveVideoFromPlaylistReq struct {
	PlaylistID int64 `uri:"id" binding:"required"`     // 播放列表ID（路径参数）
	VideoID    int64 `uri:"videoId" binding:"required"` // 视频ID（路径参数）
}
