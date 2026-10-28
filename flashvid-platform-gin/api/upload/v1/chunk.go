package v1

type InitChunkUploadReq struct {
	FileType  string `json:"file_type" binding:"required,oneof=video audio"`
	FileName  string `json:"file_name" binding:"required"`
	FileSize  int64  `json:"file_size" binding:"required,gt=0"`
	ChunkSize int64  `json:"chunk_size"` // 可选，默认 5MB
}

type InitChunkUploadResp struct {
	UploadID    string `json:"upload_id"`
	ChunkSize   int64  `json:"chunk_size"`
	TotalChunks int    `json:"total_chunks"`
}

type UploadChunkReq struct {
	UploadID   string `form:"upload_id" binding:"required"`
	ChunkIndex int    `form:"chunk_index" binding:"required,gte=0"`
}

type UploadChunkResp struct {
	ChunkIndex     int   `json:"chunk_index"`
	UploadedChunks []int `json:"uploaded_chunks"`
	IsComplete     bool  `json:"is_complete"`
}

type CompleteChunkUploadReq struct {
	UploadID string `json:"upload_id" binding:"required"`
}

type CompleteChunkUploadResp struct {
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
}
