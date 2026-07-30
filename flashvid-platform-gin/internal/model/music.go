package model

// MusicInfo 音乐信息
type MusicInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	CoverURL  string `json:"coverUrl"`
	MusicURL  string `json:"musicUrl"`
	Duration  int32  `json:"duration"`
	UseCount  int32  `json:"useCount"`
	CreatedAt string `json:"createdAt"`
}

// MusicListOutput 音乐列表输出
type MusicListOutput struct {
	List      []MusicInfo `json:"list"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"pageSize"`
}
