package upload

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/upload/v1"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/pkg/storage"

	"github.com/spf13/viper"
)

const (
	defaultChunkSize = 5 * 1024 * 1024  // 5MB
	maxChunkSize     = 10 * 1024 * 1024 // 10MB
)

var rdb = dao.RedisClient
var chunkBasePath string
var chunkExpireHours int

// MustInitChunkUpload 初始化分片上传配置
func MustInitChunkUpload(cfg *viper.Viper) {
	chunkBasePath = cfg.GetString("storage.chunk_path")
	if chunkBasePath == "" {
		chunkBasePath = "./uploads/chunks"
	}
	chunkExpireHours = cfg.GetInt("storage.chunk_expire_hours")
	if chunkExpireHours == 0 {
		chunkExpireHours = 24
	}

	if err := os.MkdirAll(chunkBasePath, 0755); err != nil {
		panic(fmt.Sprintf("create chunk directory failed: %v", err))
	}
}

// InitChunkUpload 初始化分片上传
func InitChunkUpload(ctx context.Context, req v1.InitChunkUploadReq) (*v1.InitChunkUploadResp, api.ResCode, error) {
	// 1. 校验文件大小
	var maxSize int64
	switch req.FileType {
	case "video":
		maxSize = 524288000 // 500MB
	case "audio":
		maxSize = 52428800 // 50MB
	default:
		return nil, api.CodeInvalidParam, fmt.Errorf("unsupported file type")
	}

	if req.FileSize > maxSize {
		return nil, api.CodeVideoTooLarge, fmt.Errorf("file size exceeds limit")
	}

	// 2. 确定分片大小
	chunkSize := req.ChunkSize
	if chunkSize == 0 {
		chunkSize = defaultChunkSize
	}
	if chunkSize > maxChunkSize {
		chunkSize = maxChunkSize
	}

	// 3. 计算总分片数
	totalChunks := int((req.FileSize + chunkSize - 1) / chunkSize)

	// 4. 生成唯一上传 ID
	uploadID := fmt.Sprintf("upload_%d", time.Now().UnixNano())

	// 5. 在 Redis 中初始化元数据
	if rdb != nil {
		metaKey := fmt.Sprintf("upload:meta:%s", uploadID)
		metadata := map[string]interface{}{
			"file_type":    req.FileType,
			"file_name":    req.FileName,
			"file_size":    req.FileSize,
			"chunk_size":   chunkSize,
			"total_chunks": totalChunks,
			"created_at":   time.Now().Unix(),
		}
		if err := rdb.HSet(ctx, metaKey, metadata).Err(); err != nil {
			return nil, api.CodeInternalError, err
		}
		rdb.Expire(ctx, metaKey, time.Duration(chunkExpireHours)*time.Hour)
	}

	// 6. 创建分片存储目录
	uploadDir := filepath.Join(chunkBasePath, uploadID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, api.CodeInternalError, err
	}

	return &v1.InitChunkUploadResp{
		UploadID:    uploadID,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
	}, api.CodeSuccess, nil
}

// UploadChunk 上传单个分片
func UploadChunk(ctx context.Context, uploadID string, chunkIndex int, file io.Reader) (*v1.UploadChunkResp, api.ResCode, error) {
	if rdb == nil {
		return nil, api.CodeInternalError, fmt.Errorf("redis not available")
	}

	// 1. 读取上传元数据
	metaKey := fmt.Sprintf("upload:meta:%s", uploadID)
	metadata, err := rdb.HGetAll(ctx, metaKey).Result()
	if err != nil || len(metadata) == 0 {
		return nil, api.CodeInvalidParam, fmt.Errorf("upload session not found")
	}

	totalChunks, _ := strconv.Atoi(metadata["total_chunks"])
	if chunkIndex < 0 || chunkIndex >= totalChunks {
		return nil, api.CodeInvalidParam, fmt.Errorf("invalid chunk index")
	}

	// 2. 保存分片文件
	uploadDir := filepath.Join(chunkBasePath, uploadID)
	chunkPath := filepath.Join(uploadDir, fmt.Sprintf("%d.part", chunkIndex))

	f, err := os.Create(chunkPath)
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		os.Remove(chunkPath)
		return nil, api.CodeInternalError, err
	}

	// 3. 在 Redis 中标记该分片已上传
	chunksKey := fmt.Sprintf("upload:chunks:%s", uploadID)
	rdb.SAdd(ctx, chunksKey, chunkIndex)
	rdb.Expire(ctx, chunksKey, time.Duration(chunkExpireHours)*time.Hour)

	// 4. 获取已上传的分片列表
	uploadedChunks, err := rdb.SMembers(ctx, chunksKey).Result()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	uploadedIndexes := make([]int, 0, len(uploadedChunks))
	for _, chunk := range uploadedChunks {
		idx, _ := strconv.Atoi(chunk)
		uploadedIndexes = append(uploadedIndexes, idx)
	}

	isComplete := len(uploadedIndexes) == totalChunks

	return &v1.UploadChunkResp{
		ChunkIndex:     chunkIndex,
		UploadedChunks: uploadedIndexes,
		IsComplete:     isComplete,
	}, api.CodeSuccess, nil
}

// CompleteChunkUpload 完成分片上传，合并文件
func CompleteChunkUpload(ctx context.Context, uploadID string) (*v1.CompleteChunkUploadResp, api.ResCode, error) {
	if rdb == nil {
		return nil, api.CodeInternalError, fmt.Errorf("redis not available")
	}

	// 1. 读取元数据
	metaKey := fmt.Sprintf("upload:meta:%s", uploadID)
	metadata, err := rdb.HGetAll(ctx, metaKey).Result()
	if err != nil || len(metadata) == 0 {
		return nil, api.CodeInvalidParam, fmt.Errorf("upload session not found")
	}

	fileType := metadata["file_type"]
	fileName := metadata["file_name"]
	totalChunks, _ := strconv.Atoi(metadata["total_chunks"])

	// 2. 校验所有分片是否已上传
	chunksKey := fmt.Sprintf("upload:chunks:%s", uploadID)
	uploadedCount, err := rdb.SCard(ctx, chunksKey).Result()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if int(uploadedCount) != totalChunks {
		return nil, api.CodeInvalidParam, fmt.Errorf("not all chunks uploaded")
	}

	// 3. 合并分片
	uploadDir := filepath.Join(chunkBasePath, uploadID)
	ext := filepath.Ext(fileName)
	finalFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	finalPath := filepath.Join(storage.LocalPath(), fileType, finalFileName)

	// 确保目标目录存在
	os.MkdirAll(filepath.Dir(finalPath), 0755)

	finalFile, err := os.Create(finalPath)
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	defer finalFile.Close()

	var totalSize int64
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(uploadDir, fmt.Sprintf("%d.part", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return nil, api.CodeInternalError, fmt.Errorf("chunk %d missing", i)
		}

		n, err := io.Copy(finalFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		totalSize += n
	}

	// 4. 生成访问 URL
	baseURL := storage.BaseURL()
	fileURL := fmt.Sprintf("%s/%s/%s", baseURL, fileType, finalFileName)

	// 5. 清理临时文件
	os.RemoveAll(uploadDir)

	// 6. 清理 Redis 元数据
	rdb.Del(ctx, metaKey, chunksKey)

	return &v1.CompleteChunkUploadResp{
		FileURL:  fileURL,
		FileSize: totalSize,
		FileType: fileType,
	}, api.CodeSuccess, nil
}
