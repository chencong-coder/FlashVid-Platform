package model

import (
	"time"
)

// 视频作者信息
type VideoAuthor struct {
	ID       int64  `json:"id"`       // 用户ID
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像
}

// 视频统计数据
type VideoStats struct {
	ViewCount     int32 `json:"viewCount"`     // 播放量
	LikeCount     int32 `json:"likeCount"`     // 点赞数
	CommentCount  int32 `json:"commentCount"`  // 评论数
	ShareCount    int32 `json:"shareCount"`    // 分享数
	FavoriteCount int32 `json:"favoriteCount"` // 收藏数
}


// 视频信息
type VideoInfo struct {
	ID              int64           `json:"id"`               // 视频ID
	Title           string          `json:"title"`            // 标题
	Description     string          `json:"description"`      // 描述
	CoverUrl        string          `json:"coverUrl"`        // 封面URL
	VideoUrl        string          `json:"videoUrl"`        // 视频URL
	Duration        int32           `json:"duration"`         // 时长（秒）
	Width           int32           `json:"width"`            // 宽度
	Height          int32           `json:"height"`           // 高度
	MusicId       int64          `json:"musicId"`       // 背景音乐ID
	City            string          `json:"city"`             // 城市
	Topics          []string        `json:"topics"`           // 话题标签
	Author          VideoAuthor     `json:"author"`           // 作者信息
	Stats           VideoStats      `json:"stats"`            // 统计数据
	PublishedAt      	time.Time          `json:"publishedAt"`       // 发布时间（ISO8601格式）
}

// 分页信息
type Pagination struct {
	Page       int   `json:"page"`        // 当前页
	PageSize   int   `json:"pageSize"`   // 每页数量
	Total      int64 `json:"total"`       // 总记录数
	TotalPages int   `json:"totalPages"` // 总页数
}

// 视频列表输出
type VideoListOutput struct {
	Videos     []VideoInfo `json:"videos"`     // 视频列表
	Pagination Pagination  `json:"pagination"` // 分页信息
}