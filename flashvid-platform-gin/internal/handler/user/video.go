package user

import (
	"flashvid-platform-gin/api"
	"strconv"
	v1 "flashvid-platform-gin/api/user/v1"
	"github.com/gin-gonic/gin"
	"flashvid-platform-gin/internal/service/user"
)

// 获取用户视频列表接口
func GetUserVideosHandler(c *gin.Context) {
	// 1.获取userid
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || userId <= 0 {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2.解析分页参数
	var req v1.GetUserVideosReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	  // 设置默认值
    if req.Page <= 0 {
        req.Page = 1
    }
    if req.PageSize <= 0 {
        req.PageSize = 10
    }
	// 3.调用service获取用户视频列表
	output, resCode, err := user.GetUserVideos(c, userId, req.Page, req.PageSize)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4.返回响应
	api.ResponseSuccess(c, v1.GetUserVideosResp{
		Videos:     output.Videos,
		Pagination: output.Pagination,
	})
}