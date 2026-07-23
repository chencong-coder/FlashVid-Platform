package user

import (
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/middleware"
	"github.com/gin-gonic/gin"
	"flashvid-platform-gin/internal/service/user"
	v1 "flashvid-platform-gin/api/user/v1"
)

func GetUserFavoritesHandler(c *gin.Context) {
	// 1. 获取用户ID 查看登录用户的收藏列表
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
	// 2. 获取分页参数
	var req v1.GetUserFavoritesReq
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
	// 3. 调用service获取用户收藏列表
	output, resCode, err := user.GetUserFavorites(c, userIdInt64, req.Page, req.PageSize)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 3. 返回响应
	api.ResponseSuccess(c, v1.GetUserFavoritesVideosResp{
		Videos:     output.Videos,
		Pagination: output.Pagination,
	})
}