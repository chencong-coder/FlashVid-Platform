package music

import (
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/music/v1"
	"flashvid-platform-gin/internal/service/music"

	"github.com/gin-gonic/gin"
)

// GetMusicListHandler 获取音乐列表接口
func GetMusicListHandler(c *gin.Context) {
	// 1. 获取请求参数
	var req v1.GetMusicListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 2. 设置默认值
	if req.Sort == "" {
		req.Sort = "hot"
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	// 3. 调用 service 获取音乐列表
	output, resCode, err := music.GetMusicList(c, req.Sort, req.Page, req.PageSize)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}

	// 4. 返回响应
	api.ResponseSuccess(c, v1.GetMusicListResp{
		List:     output.List,
		Total:    output.Total,
		Page:     output.Page,
		PageSize: output.PageSize,
	})
}

// CreateMusicHandler 创建音乐接口（需登录）
func CreateMusicHandler(c *gin.Context) {
	var req v1.CreateMusicReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	info, resCode, err := music.CreateMusic(c, req)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}

	api.ResponseSuccess(c, v1.CreateMusicResp{Music: *info})
}

// SearchMusicHandler 搜索音乐接口
func SearchMusicHandler(c *gin.Context) {
	// 1. 获取请求参数
	var req v1.SearchMusicReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 2. 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	// 3. 调用 service 进行音乐搜索
	output, resCode, err := music.SearchMusic(c, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}

	// 4. 返回响应
	api.ResponseSuccess(c, v1.SearchMusicResp{
		List:     output.List,
		Total:    output.Total,
		Page:     output.Page,
		PageSize: output.PageSize,
	})
}
