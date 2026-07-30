package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	localPath string
	baseURL   string
	maxSizes  map[string]int64
)

// MustInit 初始化本地存储，启动时调用
func MustInit(cfg *viper.Viper) {
	localPath = cfg.GetString("storage.local_path")
	baseURL = strings.TrimRight(cfg.GetString("storage.base_url"), "/")
	maxSizes = map[string]int64{
		"image": cfg.GetInt64("storage.max_image_size"),
		"video": cfg.GetInt64("storage.max_video_size"),
		"audio": cfg.GetInt64("storage.max_audio_size"),
	}
	// 预建子目录
	for _, dir := range []string{"image", "video", "audio"} {
		if err := os.MkdirAll(filepath.Join(localPath, dir), 0755); err != nil {
			panic(fmt.Sprintf("storage: 创建目录失败 %s: %v", dir, err))
		}
	}
}

// MaxSize 返回指定文件类型的最大允许大小（字节）
func MaxSize(fileType string) int64 {
	if size, ok := maxSizes[fileType]; ok && size > 0 {
		return size
	}
	return 10 * 1024 * 1024 // 默认 10MB
}

// Save 将 src 写入本地磁盘，返回可访问的 URL 和文件大小
func Save(fileType, originalName string, src io.Reader) (url string, size int64, err error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	// 用纳秒时间戳保证文件名唯一
	unique := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(localPath, fileType, unique)

	f, err := os.Create(dst)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	size, err = io.Copy(f, src)
	if err != nil {
		os.Remove(dst)
		return "", 0, err
	}

	url = fmt.Sprintf("%s/%s/%s", baseURL, fileType, unique)
	return url, size, nil
}

// LocalPath 返回存储根目录（用于 gin 静态文件服务）
func LocalPath() string {
	return localPath
}
