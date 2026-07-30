package server

import (
	"flashvid-platform-gin/internal/handler/auth"
	"flashvid-platform-gin/internal/handler/user"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/pkg/logging"
	"net/http"
	"flashvid-platform-gin/internal/handler/video"
	"flashvid-platform-gin/internal/handler/feed"
	"flashvid-platform-gin/internal/handler/interaction"
	"flashvid-platform-gin/internal/handler/comment"
	"flashvid-platform-gin/internal/handler/topic"
	"flashvid-platform-gin/internal/handler/music"
	"flashvid-platform-gin/internal/handler/upload"
	"flashvid-platform-gin/pkg/storage"
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

	// 静态文件服务：/static 映射到本地存储目录
	r.Static("/static", storage.LocalPath())

	apiV1 := r.Group("/api/v1")
	{
		// 登录注册相关路由组
		authUser := apiV1.Group("/auth")
		{
			authUser.POST("/register", auth.RegisterHandler) // 注册
			authUser.POST("/login", auth.LoginHandler) // 登录
			authUser.POST("/refresh", auth.RefreshHandler) // 刷新Token
		}

		// 用户相关路由组
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

		// 视频相关路由组
		videoR := apiV1.Group("/videos")
		{
			// 公开路由（无需登录）— 静态路由必须在动态路由 /:id 之前注册
			videoR.GET("/search", video.SearchVideosHandler) // 搜索视频
			videoR.GET("/:id", video.GetVideoHandler) // 获取视频详情

			// 需要登录
			videoR.POST("", middleware.Auth(), video.CreateVideoHandler) // 发布视频
			videoR.DELETE("/:id", middleware.Auth(), video.DeleteVideoHandler) // 删除视频
		}

		// 视频流路由组
		feedR := apiV1.Group("/feed")
		feedR.Use(middleware.Auth())
		{
			feedR.GET("recommend", feed.GetFeedRecommendHandler) // 获取推荐视频流
			feedR.GET("follow", feed.GetFeedFollowHandler) // 获取关注视频流
			feedR.GET("friends", feed.GetFeedFriendsHandler) // 获取好友视频流（互相关注）
			feedR.GET("nearby", feed.GetFeedNearbyHandler) // 获取附近视频流
		}

		// 互动相关路由组
		interactionR := apiV1.Group("/videos")
		interactionR.Use(middleware.Auth())
		{
			interactionR.POST("/:id/like", interaction.LikeVideoHandler) // 点赞视频
			interactionR.DELETE("/:id/like", interaction.UnlikeVideoHandler) // 取消点赞视频
			interactionR.POST("/:id/favorite", interaction.FavoriteVideoHandler) // 收藏视频
			interactionR.DELETE("/:id/favorite", interaction.UnfavoriteVideoHandler) // 取消收藏视频
			interactionR.POST("/:id/share", interaction.ShareVideoHandler) // 分享视频
		}

		// 评论相关路由组（公开接口，不强制登录）
		commentR := apiV1.Group("/videos")
		{
			commentR.GET("/:id/comments", comment.GetCommentsHandler) // 获取评论列表
		}

		// 评论写操作路由（需要登录）
		commentWriteR := apiV1.Group("/videos")
		commentWriteR.Use(middleware.Auth())
		{
			commentWriteR.POST("/:id/comments", comment.CreateCommentHandler) // 发表评论
		}

		// 评论回复路由
		replyR := apiV1.Group("/comments")
		{
			replyR.GET("/:id/replies", comment.GetRepliesHandler) // 获取回复列表
		}

		// 评论删除路由（需要登录）
		commentDeleteR := apiV1.Group("/comments")
		commentDeleteR.Use(middleware.Auth())
		{
			commentDeleteR.DELETE("/:id", comment.DeleteCommentHandler) // 删除评论
		}

		// 评论点赞路由（需要登录）
		commentLikeR := apiV1.Group("/comments")
		commentLikeR.Use(middleware.Auth())
		{
			commentLikeR.POST("/:id/like", comment.LikeCommentHandler)    // 点赞评论
			commentLikeR.DELETE("/:id/like", comment.UnlikeCommentHandler) // 取消点赞评论
		}

		// 话题相关路由组
		topicR := apiV1.Group("/topics")
		{
			topicR.GET("", topic.GetTopicsHandler)
		 	topicR.GET("/search", topic.SearchTopicsHandler)
		 	topicR.GET("/:id", topic.GetTopicByIDHandler)
		 	topicR.GET("/:id/videos", topic.GetTopicVideosHandler)
		}

		// 音乐相关路由组
		musicR := apiV1.Group("/music")
		{
			musicR.GET("", music.GetMusicListHandler)           // 获取音乐列表
			musicR.GET("/search", music.SearchMusicHandler)     // 搜索音乐
		}

		// 上传相关路由组（需要登录）
		uploadR := apiV1.Group("/upload")
		uploadR.Use(middleware.Auth())
		{
			uploadR.POST("", upload.UploadFileHandler) // 上传文件（图片/视频/音频）
		}
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
