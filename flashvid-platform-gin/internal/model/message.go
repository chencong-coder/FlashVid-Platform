package model

// MessageUserInfo 私信中对方用户简要信息
type MessageUserInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// LastMessageInfo 会话中最后一条消息的摘要
type LastMessageInfo struct {
	ID          int64  `json:"id"`
	MessageType int32  `json:"messageType"`
	Content     string `json:"content"`
	MediaURL    string `json:"mediaUrl"`
	CreatedAt   string `json:"createdAt"`
}

// ConversationInfo 会话列表条目
type ConversationInfo struct {
	TargetUser  MessageUserInfo `json:"targetUser"`
	LastMessage LastMessageInfo `json:"lastMessage"`
	UnreadCount int32           `json:"unreadCount"`
	UpdatedAt   string          `json:"updatedAt"`
}

// ConversationListOutput 会话列表分页输出
type ConversationListOutput struct {
	List     []ConversationInfo `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

// MessageInfo 单条私信
type MessageInfo struct {
	ID          int64  `json:"id"`
	FromUserID  int64  `json:"fromUserId"`
	ToUserID    int64  `json:"toUserId"`
	MessageType int32  `json:"messageType"`
	Content     string `json:"content"`
	MediaURL    string `json:"mediaUrl"`
	IsRead      bool   `json:"isRead"`
	CreatedAt   string `json:"createdAt"`
}

// MessageListOutput 消息列表游标分页输出
type MessageListOutput struct {
	Messages        []MessageInfo `json:"messages"`
	NextCursorToken string        `json:"nextCursorToken"`
	HasMore         bool          `json:"hasMore"`
}

// SendMessageInput 发送私信的业务入参
type SendMessageInput struct {
	ToUserID    int64
	MessageType int32
	Content     string
	MediaURL    string
}
