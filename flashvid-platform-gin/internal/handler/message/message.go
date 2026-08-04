package message

import (
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/message/v1"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/internal/model"
	svc "flashvid-platform-gin/internal/service/message"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetConversationsHandler 获取会话列表
func GetConversationsHandler(c *gin.Context) {
	loginUserID, ok := mustLogin(c)
	if !ok {
		return
	}
	var req v1.GetConversationsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	output, resCode, err := svc.GetConversations(c, loginUserID, req.Page, req.PageSize)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, v1.GetConversationsResp{
		List:     output.List,
		Total:    output.Total,
		Page:     output.Page,
		PageSize: output.PageSize,
	})
}

// GetConversationMessagesHandler 获取与指定用户的对话消息（游标分页）
func GetConversationMessagesHandler(c *gin.Context) {
	loginUserID, ok := mustLogin(c)
	if !ok {
		return
	}
	targetUserID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil || targetUserID <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	var req v1.GetMessagesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Count < 1 {
		req.Count = 20
	}
	output, resCode, err := svc.GetMessages(c, loginUserID, targetUserID, req.Cursor, req.Count)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, v1.GetMessagesResp{
		Messages:        output.Messages,
		NextCursorToken: output.NextCursorToken,
		HasMore:         output.HasMore,
	})
}

// SendMessageHandler 发送私信
func SendMessageHandler(c *gin.Context) {
	loginUserID, ok := mustLogin(c)
	if !ok {
		return
	}
	var req v1.SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	toUserID, err := strconv.ParseInt(req.ToUserID, 10, 64)
	if err != nil || toUserID <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	info, resCode, err := svc.SendMessage(c, loginUserID, &model.SendMessageInput{
		ToUserID:    toUserID,
		MessageType: req.MessageType,
		Content:     req.Content,
		MediaURL:    req.MediaURL,
	})
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, v1.SendMessageResp{MessageInfo: *info})
}

// MarkConversationReadHandler 标记会话已读
func MarkConversationReadHandler(c *gin.Context) {
	loginUserID, ok := mustLogin(c)
	if !ok {
		return
	}
	targetUserID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil || targetUserID <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	readCount, resCode, err := svc.MarkConversationRead(c, loginUserID, targetUserID)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, v1.MarkReadResp{ReadCount: readCount})
}

// DeleteMessageHandler 删除消息
func DeleteMessageHandler(c *gin.Context) {
	loginUserID, ok := mustLogin(c)
	if !ok {
		return
	}
	messageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || messageID <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	resCode, err := svc.DeleteMessage(c, loginUserID, messageID)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, gin.H{})
}

// GetUnreadCountHandler 获取未读私信总数
func GetUnreadCountHandler(c *gin.Context) {
	loginUserID, ok := mustLogin(c)
	if !ok {
		return
	}
	count, resCode, err := svc.GetUnreadCount(c, loginUserID)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, v1.UnreadCountResp{UnreadCount: count})
}

// mustLogin 提取登录用户 ID，未登录时直接响应并返回 false
func mustLogin(c *gin.Context) (int64, bool) {
	val, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeNeedLogin)
		return 0, false
	}
	id, ok := val.(int64)
	if !ok || id <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return 0, false
	}
	return id, true
}
