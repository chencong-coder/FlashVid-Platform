package interaction

import (
	"flashvid-platform-gin/api"
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
	// 3. 调用service进行点赞操作
	resp, resCode, err := interaction.LikeVideo(c, userIdInt64, videoId)
	if err != nil {
		if resCode == api.CodeAlreadyLiked {
			api.ResponseErrorWithMsg(c, resCode, "已经点赞过该视频")
			return
		}
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
	// 3. 调用service进行点赞操作
	resp, resCode, err := interaction.UnlikeVideo(c, userIdInt64, videoId)
	if err != nil {
		if resCode == api.CodeAlreadyUnliked {
			api.ResponseErrorWithMsg(c, resCode, "已经取消点赞过该视频")
			return
		}
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, resp)
}