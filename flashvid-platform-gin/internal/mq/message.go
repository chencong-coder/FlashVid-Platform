package mq

// VideoPublishMessage 视频发布事件消息
type VideoPublishMessage struct {
	VideoID  int64   `json:"video_id"`
	UserID   int64   `json:"user_id"`
	TopicIDs []int64 `json:"topic_ids"`
}

// HotrankUpdateMessage 热度更新消息
type HotrankUpdateMessage struct {
	Action  string `json:"action"` // "update_hot_score" | "update_topic_view"
	VideoID int64  `json:"video_id,omitempty"`
	TopicID int64  `json:"topic_id,omitempty"`
}
