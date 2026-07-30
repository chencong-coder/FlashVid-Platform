package upload

import (
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/upload/v1"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/internal/service/upload"

	"github.com/gin-gonic/gin"
)

// UploadFileHandler 上传文件接口
func UploadFileHandler(c *gin.Context) {
	// 1. 验证登录态
	_, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	// 2. 绑定 file_type 参数
	var req v1.UploadFileReq
	if err := c.ShouldBind(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 3. 获取上传文件
	fh, err := c.FormFile("file")
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 4. 调用 service 处理上传
	output, resCode, err := upload.UploadFile(c, req.FileType, fh)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}

	// 5. 返回响应
	api.ResponseSuccess(c, v1.UploadFileResp{
		FileURL:  output.FileURL,
		FileSize: output.FileSize,
		FileType: output.FileType,
		Duration: output.Duration,
	})
}
