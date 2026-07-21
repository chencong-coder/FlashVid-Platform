package server

import (
	"flashvid-platform-gin/internal/handler/auth"
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
			authUser.POST("/register", auth.RegisterHandler)
			authUser.POST("/login", auth.LoginHandler)
			authUser.POST("/refresh", auth.RefreshHandler)
		}
		authUser.Use(middleware.Auth()) // 需要登录 用Token验证身份
		
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
