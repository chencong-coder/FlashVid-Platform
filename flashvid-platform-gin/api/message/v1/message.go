package v1

import "flashvid-platform-gin/internal/model"

// GetConversationsReq 获取会话列表请求
type GetConversationsReq struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// GetConversationsResp 获取会话列表响应
type GetConversationsResp struct {
	List     []model.ConversationInfo `json:"list"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
}

// GetMessagesReq 获取对话消息请求（游标分页）
type GetMessagesReq struct {
	Cursor string `form:"cursor" binding:"omitempty"`
	Count  int    `form:"count" binding:"omitempty,min=1,max=50"`
}

// GetMessagesResp 获取对话消息响应
type GetMessagesResp struct {
	Messages        []model.MessageInfo `json:"messages"`
	NextCursorToken string              `json:"nextCursorToken"`
	HasMore         bool                `json:"hasMore"`
}

// SendMessageReq 发送私信请求
type SendMessageReq struct {
	ToUserID    int64  `json:"toUserId" binding:"required,min=1"`
	MessageType int32  `json:"messageType" binding:"required,oneof=1 2 3"`
	Content     string `json:"content" binding:"omitempty,max=1000"`
	MediaURL    string `json:"mediaUrl" binding:"omitempty"`
}

// SendMessageResp 发送私信响应
type SendMessageResp struct {
	model.MessageInfo
}

// MarkReadResp 标记已读响应
type MarkReadResp struct {
	ReadCount int64 `json:"readCount"`
}

// UnreadCountResp 未读总数响应
type UnreadCountResp struct {
	UnreadCount int64 `json:"unreadCount"`
}
