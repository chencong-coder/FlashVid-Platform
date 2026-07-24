package v1

import (
	"flashvid-platform-gin/internal/model"
)

//发布视频请求
type CreateVideoReq struct {
    Title       *string   `json:"title"       binding:"required,max=100"` // 视频标题，必填，最大长度100
    Description *string   `json:"description" binding:"omitempty,max=500"` // 视频描述，可选，最大长度500
    CoverUrl    *string   `json:"coverUrl"    binding:"required"` // 视频封面URL，必填
    VideoUrl    *string   `json:"videoUrl"    binding:"required"` // 视频文件URL，必填
    Duration    *int32    `json:"duration"    binding:"required,min=1"` // 视频时长，单位为秒，必填，最小值1
    Width       *int32    `json:"width"       binding:"omitempty,min=1"` // 视频宽度，可选，最小值1
    Height      *int32    `json:"height"      binding:"omitempty,min=1"` // 视频高度，可选，最小值1
    MusicId     *int64    `json:"musicId"     binding:"omitempty"` // 背景音乐ID，可选
    City        *string   `json:"city"        binding:"omitempty"` // 视频拍摄城市，可选
	Location    *string   `json:"location"    binding:"omitempty"` // 视频拍摄地点，可选
	Latitude    *float64  `json:"latitude"    binding:"omitempty,gte=-90,lte=90"` // 视频拍摄地点纬度，可选，范围[-90, 90]
	Longitude   *float64  `json:"longitude"   binding:"omitempty,gte=-180,lte=180"` // 视频拍摄地点经度，可选，范围[-180, 180]
    Topics      *[]int64  `json:"topics"      binding:"omitempty,max=5"` // 视频话题标签列表，可选，最多5个
}

// 发布视频响应
type CreateVideoResp struct {
	VideoID int64 `json:"video_id"` // 发布成功的视频ID
	Status  int   `json:"status"`   // 发布状态，1表示成功，0表示失败
}

// 获取视频详情响应
type GetVideoResp struct {
	Video model.VideoInfo `json:"video"` // 视频详情
}

// 搜索视频请求
type SearchVideosReq struct {
    Keyword  string `form:"keyword"  binding:"required,min=1"` // 搜索关键词，必填，最小长度1
    Page     int    `form:"page"     binding:"omitempty,min=1"` // 页码，可选，最小值1
    PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=50"` // 每页数量，可选，最小值1，最大值50
}

// 搜索视频响应
type SearchVideosResp struct {
	Videos     []model.VideoInfo `json:"videos"`     // 视频列表
	Pagination model.Pagination  `json:"pagination"` // 分页信息
}
