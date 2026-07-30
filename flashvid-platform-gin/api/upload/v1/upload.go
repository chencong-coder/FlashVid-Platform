package v1

// UploadFileReq 上传文件请求
type UploadFileReq struct {
	FileType string `form:"file_type" binding:"required,oneof=image video audio"`
}

// UploadFileResp 上传文件响应
type UploadFileResp struct {
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
	Duration int32  `json:"duration"` // 视频/音频时长（秒），本地存储暂不解析，返回 0
}
