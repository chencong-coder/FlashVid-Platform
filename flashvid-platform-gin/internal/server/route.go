package server

import (
	"flashvid-platform-gin/internal/handler/auth"
	"flashvid-platform-gin/internal/handler/user"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/pkg/logging"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func SetupRoutes(cfg *viper.Viper) *gin.Engine {
	r := gin.New()
	r.Use(logging.GinLogger(), logging.GinRecovery(true)) // 日志中间件，记录请求日志
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowHeaders = append(corsCfg.AllowHeaders, "Authorization")
	corsCfg.AllowAllOrigins = true
	r.Use(cors.New(corsCfg)) // CORS 跨域中间件，简单粗暴，直接放行所有跨域请求
	apiV1 := r.Group("/api/v1")
	{
		authUser := apiV1.Group("/auth")
		{	
			authUser.POST("/register", auth.RegisterHandler) // 注册
			authUser.POST("/login", auth.LoginHandler) // 登录
			authUser.POST("/refresh", auth.RefreshHandler) // 刷新Token
		}
		authUser.Use(middleware.Auth()) // 需要登录 用Token验证身份
		userR := apiV1.Group("/user")
		{
			userR.GET("/:id", user.GetUserInfoHandler) // 获取用户信息
			userR.PUT("/:id", user.UpdateUserInfoHandler) // 更新用户信息
			userR.GET("/:id}/videos", user.GetUserVideosHandler) // 获取用户发布的视频
			userR.GET("/:id}/likes", user.GetUserLikesHandler) // 获取用户点赞的视频
			userR.GET("/:id}/favorites", user.GetUserFavoritesHandler) // 获取用户收藏的视频
		}
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
