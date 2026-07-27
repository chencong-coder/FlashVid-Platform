package feed

import (
	"flashvid-platform-gin/api"
	"github.com/gin-gonic/gin"
	v1 "flashvid-platform-gin/api/feed/v1"
	"flashvid-platform-gin/internal/service/feed"
)

// GetFeedRecommendHandler 获取推荐视频流接口
func GetFeedRecommendHandler(c *gin.Context) {
	// 1. 获取游标和分页参数
	var req v1.RecommendFeedReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Count <= 0 {
		req.Count = 10
	}
	// 2. 调用service获取推荐视频流
	output, resCode, err := feed.GetFeedRecommend(c, req.Cursor, req.Count)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 3. 返回响应
	api.ResponseSuccess(c, v1.FeedResp{
		Videos:     output.Videos,
		NextCursorToken: output.NextCursorToken,
		HasMore:    output.HasMore,
	})
}

// GetFeedFollowHandler 获取关注视频流接口
func GetFeedFollowHandler(c *gin.Context) {
	// 1. 获取用户ID 查看登录用户的关注视频流
	userId, exists := c.Get("userID")
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userIdInt64, ok := userId.(int64)
	if !ok || userIdInt64 <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	// 2. 获取游标和分页参数
	var req v1.FollowFeedReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Count <= 0 {
		req.Count = 10
	}
	// 3. 调用service获取推荐视频流
	output, resCode, err := feed.GetFeedFollow(c, userIdInt64, req.Cursor, req.Count)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, v1.FeedResp{
		Videos:     output.Videos,
		NextCursorToken: output.NextCursorToken,
		HasMore:    output.HasMore,
	})
}