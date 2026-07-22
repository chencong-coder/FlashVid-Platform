package user

import (
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/internal/service/user"
	v1 "flashvid-platform-gin/api/user/v1"
	"github.com/gin-gonic/gin"
	"strconv"
)

// 获取用户信息接口
func GetUserInfoHandler(c *gin.Context) {
	// 1. 获取用户ID参数
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || userId <= 0 {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2. 调用服务层获取用户信息
	output, rescode, err := user.GetUserInfo(c, userId)
	if err != nil {
		api.ResponseError(c, rescode)
		return
	}
	// 3. 返回用户信息响应
	api.ResponseSuccess(c, &v1.UserInfoResp{
		UserId:         output.UserId,
		Username:       output.Username,
		Nickname:       output.Nickname,
		Avatar:         output.Avatar,
		Bio:            output.Bio,
		City:           output.City,
		FollowersConut: output.FollowersCount,
		FollowingConut: output.FollowingCount,
		VideosCount:    output.VideosCount,
		LikesCount:     output.LikesCount,
		Phone:          output.Phone,
		Gender:         output.Gender,
		Birthday:       output.Birthday.Format("2006-01-02"),
		Email:          output.Email,
		CreatedAt:      output.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// 获取用户发布的视频接口
func GetUserVideosHandler(c *gin.Context) {}

// 更新用户信息接口
func UpdateUserInfoHandler(c *gin.Context) {
	// 1. 从中间件context获取当前登录用户ID
	userId, ok := c.Get(middleware.CtxKeyUserID) // 使用常量，不是字符串
	if !ok {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}
	userIdInt64, ok := userId.(int64)
	if !ok || userIdInt64 <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	// 2. 解析请求体参数
	var req v1.UpdateUserInfoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3. 调用服务层更新用户信息
	output, rescode, err := user.UpdateUserInfo(c, userIdInt64, &req)
	if err != nil {
		api.ResponseError(c, rescode)
		return
	}
	// 4. 返回更新后的用户信息响应
	api.ResponseSuccess(c, &v1.UpdateUserInfoResp{
		UserId:         output.UserId,
		Username:       output.Username,
		Nickname:       output.Nickname,
		Avatar:         output.Avatar,
		Bio:            output.Bio,
		City:           output.City,
		FollowersConut: output.FollowersCount,
		FollowingConut: output.FollowingCount,
		VideosCount:    output.VideosCount,
		LikesCount:     output.LikesCount,
		Phone:          output.Phone,
		Gender:         output.Gender,
		Birthday:       output.Birthday.Format("2006-01-02"),
		Email:          output.Email,
		UpdatedAt:      output.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// 获取用户点赞的视频接口
func GetUserLikesHandler(c *gin.Context) {}

// 获取用户收藏的视频接口
func GetUserFavoritesHandler(c *gin.Context) {}
