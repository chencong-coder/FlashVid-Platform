package mq

// VideoPublishMessage 视频发布事件消息
type VideoPublishMessage struct {
	VideoID  int64   `json:"video_id"`
	UserID   int64   `json:"user_id"`
	TopicIDs []int64 `json:"topic_ids"`
}

// HotrankUpdateMessage 热度更新消息
type HotrankUpdateMessage struct {
	Action  string `json:"action"` // "update_video_hot" | "update_topic_view " | "update_video_view"
	VideoID int64  `json:"video_id,omitempty"`
	TopicID int64  `json:"topic_id,omitempty"`
}

// NotificationMessage 通知消息（未来扩展）
type NotificationMessage struct {
	UserID     int64 `json:"user_id"`      // 接收通知的用户
	ActorID    int64 `json:"actor_id"`     // 触发行为的用户
	ActionType int32 `json:"action_type"`  // 行为类型：1关注 2点赞 3收藏
	TargetType int32 `json:"target_type"`  // 目标类型：1用户 2视频
	TargetID   int64 `json:"target_id"`    // 目标ID
}
