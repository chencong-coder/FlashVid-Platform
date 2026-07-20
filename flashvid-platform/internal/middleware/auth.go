package middleware

import (
	"strings"

	"flashvid-platform/api"

	"flashvid-platform/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	tokenPrefix = "Bearer "

	CtxKeyUserID = "userId" // 用户ID上下文 key

)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 token
		authorizationValue := c.GetHeader("Authorization")
		if len(authorizationValue) == 0 || !strings.HasPrefix(authorizationValue, tokenPrefix) {
			api.ResponseError(c, api.CodeNeedLogin)
			c.Abort()
			return
		}
		// 判断token是否合法
		if len(authorizationValue) <= 7 || !strings.HasPrefix(authorizationValue, tokenPrefix) {
			api.ResponseError(c, api.CodeInvalidToken)
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authorizationValue, tokenPrefix)
		// 解析token，获取claims
		claims, err := jwt.ParseAccessToken(tokenString)
		if err != nil {
			zap.L().Sugar().Debugf("parse access token error: %v", err)
			api.ResponseError(c, api.CodeInvalidToken)
			c.Abort()
			return
		}
		c.Set(CtxKeyUserID, claims.UserId)
		c.Next()
	}
}
