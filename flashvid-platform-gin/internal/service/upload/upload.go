package upload

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"flashvid-platform-gin/api"
	"flashvid-platform-gin/pkg/storage"
)

// allowedExts 各文件类型允许的扩展名
var allowedExts = map[string]map[string]bool{
	"image": {".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true},
	"video": {".mp4": true, ".mov": true, ".avi": true, ".mkv": true},
	"audio": {".mp3": true, ".wav": true, ".aac": true, ".ogg": true, ".m4a": true},
}

// UploadOutput 上传结果
type UploadOutput struct {
	FileURL  string
	FileSize int64
	FileType string
	Duration int32
}

// UploadFile 校验并保存文件到本地磁盘
func UploadFile(_ context.Context, fileType string, fh *multipart.FileHeader) (*UploadOutput, api.ResCode, error) {
	// 1. 检查文件大小
	if fh.Size > storage.MaxSize(fileType) {
		return nil, api.CodeVideoTooLarge, fmt.Errorf("文件大小 %d 超出限制 %d", fh.Size, storage.MaxSize(fileType))
	}

	// 2. 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedExts[fileType][ext] {
		return nil, api.CodeVideoFormatInvalid, fmt.Errorf("不支持的文件格式: %s", ext)
	}

	// 3. 打开文件句柄
	f, err := fh.Open()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	defer f.Close()

	// 4. 写入本地磁盘
	url, size, err := storage.Save(fileType, fh.Filename, f)
	if err != nil {
		return nil, api.CodeUploadFailed, err
	}

	return &UploadOutput{
		FileURL:  url,
		FileSize: size,
		FileType: fileType,
		Duration: 0, // 本地存储暂不解析时长
	}, api.CodeSuccess, nil
}
