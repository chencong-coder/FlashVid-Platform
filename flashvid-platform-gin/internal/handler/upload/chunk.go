package upload

import (
	"github.com/gin-gonic/gin"

	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/upload/v1"
	uploadSvc "flashvid-platform-gin/internal/service/upload"
)

func InitChunkUploadHandler(c *gin.Context) {
	var req v1.InitChunkUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	resp, code, err := uploadSvc.InitChunkUpload(c, req)
	if err != nil {
		api.ResponseError(c, code)
		return
	}

	api.ResponseSuccess(c, resp)
}

func UploadChunkHandler(c *gin.Context) {
	var req v1.UploadChunkReq
	if err := c.ShouldBind(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	file, err := fh.Open()
	if err != nil {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	defer file.Close()

	resp, code, err := uploadSvc.UploadChunk(c, req.UploadID, req.ChunkIndex, file)
	if err != nil {
		api.ResponseError(c, code)
		return
	}

	api.ResponseSuccess(c, resp)
}

func CompleteChunkUploadHandler(c *gin.Context) {
	var req v1.CompleteChunkUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	resp, code, err := uploadSvc.CompleteChunkUpload(c, req.UploadID)
	if err != nil {
		api.ResponseError(c, code)
		return
	}

	api.ResponseSuccess(c, resp)
}
