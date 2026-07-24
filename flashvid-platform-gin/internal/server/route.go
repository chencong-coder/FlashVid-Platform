package server

import (
	"flashvid-platform-gin/internal/handler/auth"
	"flashvid-platform-gin/internal/handler/user"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/pkg/logging"
	"net/http"
	"flashvid-platform-gin/internal/handler/video"
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

		// 需要登录的路由组
		userR := apiV1.Group("/user")
		userR.Use(middleware.Auth())
		{
			// 静态路由优先（避免和 /:id 冲突）
			userR.PUT("/profile", user.UpdateUserInfoHandler)             // 更新自己的信息
			userR.GET("/profile/likes", user.GetUserLikesHandler)         // 查看自己的点赞列表（私有）
			userR.GET("/profile/favorites", user.GetUserFavoritesHandler) // 查看自己的收藏列表（私有）
			// 动态路由
			userR.GET("/:id", user.GetUserInfoHandler)                    // 查看任意用户主页（公开）
			userR.GET("/:id/videos", user.GetUserVideosHandler)           // 查看用户发布的视频（公开）
			userR.POST("/:id/follow", user.FollowUserHandler)                 // 关注用户（私有）
			userR.DELETE("/:id/follow", user.UnfollowUserHandler) // 取消关注用户（私有）
			userR.GET("/:id/followers", user.GetUserFollowersHandler) // 查看用户的粉丝列表（公开）
			userR.GET("/:id/followings", user.GetUserFollowingHandler) // 查看用户的关注列表（公开）
		}

		videoR := apiV1.Group("/videos")
		{
			// 公开路由（无需登录）— 静态路由必须在动态路由 /:id 之前注册
			//videoR.GET("/search", video.SearchVideosHandler) // 搜索视频
			videoR.GET("/:id", video.GetVideoHandler) // 获取视频详情

			// 需要登录
			videoR.POST("", middleware.Auth(), video.CreateVideoHandler) // 发布视频
			//videoR.DELETE("/:id", middleware.Auth(), video.DeleteVideoHandler) // 删除视频
		}

		
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
