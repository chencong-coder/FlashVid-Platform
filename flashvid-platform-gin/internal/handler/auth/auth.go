package auth

import (
	"github.com/gin-gonic/gin"
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/auth/v1"
	"flashvid-platform-gin/internal/service/auth"
)

// 用户注册接口
func RegisterHandler(c *gin.Context) {
	// 1. 解析请求参数
	var req v1.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 2. 调用service层
	output, code, err := service.Register(c, &req, c.ClientIP())
	if err != nil {
		api.ResponseError(c, code)
		return
	}
	// 3. 返回响应
	api.ResponseSuccess(c, v1.RegisterResp{
		UserID:   output.UserID,
		Username: output.Username,
	})
}

// 用户登录接口
func LoginHandler(c *gin.Context) {
	
}

// 刷新Token接口
func RefreshHandler(c *gin.Context) {
	
}
