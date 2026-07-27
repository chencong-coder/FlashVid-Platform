package feed

import (
	"flashvid-platform-gin/api"
	"github.com/gin-gonic/gin"
	v1 "flashvid-platform-gin/api/feed/v1"
	"flashvid-platform-gin/internal/service/feed"
)

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