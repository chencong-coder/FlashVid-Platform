package topic

import (
	"github.com/gin-gonic/gin"
	v1 "flashvid-platform-gin/api/topic/v1"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/service/topic"
	"strconv"
)

// 获取话题列表接口
func GetTopicsHandler(c *gin.Context) {
	// 1. 获取请求参数
	var req v1.GetTopicsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Count < 10 {
		req.Count = 10
	}
	if req.Sort == "" {
		req.Sort = "hot"
	}
	// 2. 调用service获取话题列表
	output, resCode, err := topic.GetTopics(c, req.Sort, req.Cursor, req.Count)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 3. 返回响应
	api.ResponseSuccess(c, v1.GetTopicsResp{
		Topics:          output.Topics,
		NextCursorToken: output.NextCursorToken,
		HasMore:         output.HasMore,
	})
}

// 根据话题ID获取话题详情接口
func GetTopicByIDHandler(c *gin.Context) {
	// 1. 获取话题ID
	topicId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 2. 调用service获取话题详情
	output, resCode, err := topic.GetTopicByID(c, topicId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 3. 返回响应
	api.ResponseSuccess(c, &v1.GetTopicByIDResp{
		Topic: output.Topic,
	})
}

// 根据话题ID获取话题下的视频列表接口
func GetTopicVideosHandler(c *gin.Context) {
	// 1. 获取话题ID
	topicId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 2. 获取分页参数
	var req v1.GetTopicVideosReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Count < 10 {
		req.Count = 10
	}
	if req.Sort == "" {
		req.Sort = "latest"
	}
	// 3. 调用service获取话题下的视频列表	
	output, resCode, err := topic.GetTopicVideos(c, topicId, req.Sort, req.Cursor, req.Count)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, v1.GetTopicVideosResp{
		Videos:          output.Videos,
		NextCursorToken: output.NextCursorToken,
		HasMore:         output.HasMore,
	})
}

// 搜索话题接口
func SearchTopicsHandler(c *gin.Context) {
	// 1. 获取请求参数
	var req v1.SearchTopicsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Count < 10 {
		req.Count = 10
	}
	// 2. 调用service进行话题搜索
	output, resCode, err := topic.SearchTopics(c, req.Keyword, req.Cursor, req.Count)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 3. 返回响应
	api.ResponseSuccess(c, v1.GetTopicsResp{
		Topics:          output.Topics,
		NextCursorToken: output.NextCursorToken,
		HasMore:         output.HasMore,
	})
}