package task

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// CleanupExpiredChunks 清理过期的分片上传临时文件
func CleanupExpiredChunks(chunkBasePath string, expireHours int) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		zap.L().Info("start cleanup expired chunks")

		expireTime := time.Now().Add(-time.Duration(expireHours) * time.Hour)

		entries, err := os.ReadDir(chunkBasePath)
		if err != nil {
			zap.L().Error("read chunk directory failed", zap.Error(err))
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(expireTime) {
				dirPath := filepath.Join(chunkBasePath, entry.Name())
				if err := os.RemoveAll(dirPath); err != nil {
					zap.L().Error("remove expired chunk dir failed",
						zap.String("dir", dirPath),
						zap.Error(err))
				} else {
					zap.L().Info("removed expired chunk dir",
						zap.String("dir", entry.Name()))
				}
			}
		}
	}
}
