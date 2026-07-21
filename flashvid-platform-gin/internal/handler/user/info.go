package user

import (
	"flashvid-platform-gin/api"
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
	output, rescode, err := user.GetUserInfoService(c, userId)
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
		Birthday:       output.Birthday,
		Email:          output.Email,
		CreatedAt:      output.CreatedAt,
	})
}

// 获取用户发布的视频接口
func GetUserVideosHandler(c *gin.Context) {}

// 更新用户信息接口
func UpdateUserInfoHandler(c *gin.Context) {}

// 获取用户点赞的视频接口
func GetUserLikesHandler(c *gin.Context) {}

// 获取用户收藏的视频接口
func GetUserFavoritesHandler(c *gin.Context) {}
