package user

import (
	"github.com/gin-gonic/gin"
)
// 获取用户信息接口
func GetUserInfoHandler(c *gin.Context) {}

// 获取用户发布的视频接口
func GetUserVideosHandler(c *gin.Context) {}

// 更新用户信息接口
func UpdateUserInfoHandler(c *gin.Context) {}

// 获取用户点赞的视频接口
func GetUserLikesHandler(c *gin.Context) {}

// 获取用户收藏的视频接口
func GetUserFavoritesHandler(c *gin.Context) {}
