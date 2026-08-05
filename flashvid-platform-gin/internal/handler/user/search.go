package user

import (
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/user/v1"
	"flashvid-platform-gin/internal/service/user"
	"github.com/gin-gonic/gin"
)

// SearchUsersHandler 搜索用户接口
// GET /api/v1/users/search?keyword=xxx&count=6
func SearchUsersHandler(c *gin.Context) {
	var req v1.SearchUsersReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Count <= 0 {
		req.Count = 6
	}

	userInfos, resCode, err := user.SearchUsers(c, req.Keyword, req.Count)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}

	items := make([]v1.SearchUserItem, 0, len(userInfos))
	for _, u := range userInfos {
		items = append(items, v1.SearchUserItem{
			UserID:        u.UserId,
			Nickname:      u.Nickname,
			Avatar:        u.Avatar,
			Bio:           u.Bio,
			FollowerCount: u.FollowersCount,
		})
	}
	api.ResponseSuccess(c, &v1.SearchUsersResp{Users: items})
}
