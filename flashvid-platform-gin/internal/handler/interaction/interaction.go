package interaction

import (
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/interaction/v1"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/internal/service/interaction"
	"strconv"
	"github.com/gin-gonic/gin"
)

// 点赞视频接口
func LikeVideoHandler(c *gin.Context) {
	// 1. 获取登录用户ID
	loginUserId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userIdInt64, ok := loginUserId.(int64)
	if !ok || userIdInt64 <= 0 {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2. 获取视频ID
	videoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3. 调用service进行点赞操作（Redis 优化版）
	resp, resCode, err := interaction.LikeVideo1(c, userIdInt64, videoId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, resp)
}

// 取消点赞视频接口
func UnlikeVideoHandler(c *gin.Context) {
	// 1. 获取登录用户ID
	loginUserId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userIdInt64, ok := loginUserId.(int64)
	if !ok || userIdInt64 <= 0 {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2. 获取视频ID
	videoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3. 调用service进行点赞操作（Redis 优化版）
	resp, resCode, err := interaction.UnlikeVideo1(c, userIdInt64, videoId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, resp)
}

// 收藏视频接口
func FavoriteVideoHandler(c *gin.Context) {
	// 1. 获取登录用户ID
	loginUserId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userIdInt64, ok := loginUserId.(int64)
	if !ok || userIdInt64 <= 0 {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2. 获取视频ID
	videoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3. 调用service进行收藏操作
	resp, resCode, err := interaction.FavoriteVideo1(c, userIdInt64, videoId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, resp)
}

// 取消收藏视频接口
func UnfavoriteVideoHandler(c *gin.Context) {
	// 1. 获取登录用户ID
	loginUserId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userIdInt64, ok := loginUserId.(int64)
	if !ok || userIdInt64 <= 0 {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2. 获取视频ID
	videoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3. 调用service进行取消收藏操作
	resp, resCode, err := interaction.UnfavoriteVideo(c, userIdInt64, videoId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, resp)
}

// 分享视频接口
func ShareVideoHandler(c *gin.Context) {
	// 1. 获取视频ID
	videoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 2. 绑定请求参数
	var req v1.ShareVideoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3. 调用service进行分享操作
	resp, resCode, err := interaction.ShareVideo(c, videoId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, resp)
}