package video

import (
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/middleware"
	"github.com/gin-gonic/gin"
	v1 "flashvid-platform-gin/api/video/v1"
	"flashvid-platform-gin/internal/service/video"
	"strconv"
)

// CreateVideoHandler 创建视频接口
func CreateVideoHandler(c *gin.Context) {
	// 1. 获取登录用户ID
	userId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userIdInt64, ok := userId.(int64)
	if !ok || userIdInt64 <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	// 2. 获取请求参数
	var req v1.CreateVideoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3. 调用service进行视频创建操作
	output, resCode, err := video.CreateVideo(c, userIdInt64, req)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, &v1.CreateVideoResp{
		VideoID: output.VideoID,
		Status:  output.Status,
	})
}

// GetVideoHandler 获取视频详情接口
func GetVideoHandler(c *gin.Context) {
	// 1. 获取视频ID
	videoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || videoId <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 2. 调用service获取视频详情
	output, resCode, err := video.GetVideo(c, videoId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 3. 返回响应
	api.ResponseSuccess(c, &v1.GetVideoResp{
		Video: output.Video,
	})
}

// DeleteVideoHandler 删除视频接口
func DeleteVideoHandler(c *gin.Context) {
	// 1. 获取视频ID
	videoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || videoId <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 2. 获取登录用户ID
	userId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userIdInt64, ok := userId.(int64)
	if !ok || userIdInt64 <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	// 2. 调用service进行视频删除操作
	rescode, err := video.DeleteVideo(c, videoId, userIdInt64)
	if err != nil {
		api.ResponseError(c, rescode)
		return
	}
	// 3. 返回响应
	if rescode == api.CodeSuccess {
		api.ResponseSuccess(c, gin.H{
			"message": "视频删除成功",
		})
	}
}

// SearchVideosHandler 搜索视频接口
func SearchVideosHandler(c *gin.Context) {
	// 1. 获取查询参数
	var req v1.SearchVideosReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	// 2. 调用service进行视频搜索操作
	output, resCode, err := video.SearchVideos(c, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 3. 返回响应
	api.ResponseSuccess(c, &v1.SearchVideosResp{
		Videos:     output.Videos,
		Pagination: output.Pagination,
	})
}