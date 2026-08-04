package user

import (
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/user/v1"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/internal/service/user"
	"github.com/gin-gonic/gin"
	"strconv"
)

// GetRecommendUsersHandler 获取推荐用户列表（按粉丝数降序）
// GET /api/v1/user/recommend?count=5
func GetRecommendUsersHandler(c *gin.Context) {
	// 1. 获取可选登录用户ID（OptionalAuth：未登录时不在context中）
	var loginUserID int64
	if uid, ok := c.Get(middleware.CtxKeyUserID); ok {
		if id, ok := uid.(int64); ok {
			loginUserID = id
		}
	}

	// 2. 解析 count 参数，默认 5，上限 20
	count := 5
	if countStr := c.Query("count"); countStr != "" {
		if n, err := strconv.Atoi(countStr); err == nil && n > 0 && n <= 20 {
			count = n
		}
	}

	// 3. 调用服务层
	userInfos, resCode, err := user.GetRecommendUsers(c, loginUserID, count)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}

	// 4. 组装响应
	items := make([]v1.RecommendUserItem, 0, len(userInfos))
	for _, u := range userInfos {
		items = append(items, v1.RecommendUserItem{
			UserID:        u.UserId,
			Nickname:      u.Nickname,
			Avatar:        u.Avatar,
			Bio:           u.Bio,
			FollowerCount: u.FollowersCount,
		})
	}
	api.ResponseSuccess(c, &v1.GetRecommendUsersResp{Users: items})
}
