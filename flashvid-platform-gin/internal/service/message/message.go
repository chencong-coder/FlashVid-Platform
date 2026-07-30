package message

import (
	"context"
	"errors"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	repo "flashvid-platform-gin/internal/repository/message"
	"time"

	"gorm.io/gorm"
)

// GetConversations 获取当前用户的会话列表（offset 分页，按最后消息时间倒序）
func GetConversations(ctx context.Context, userID int64, page, pageSize int) (*model.ConversationListOutput, api.ResCode, error) {
	total, err := repo.CountConversations(ctx, userID)
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	convs, err := repo.ListConversations(ctx, userID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if len(convs) == 0 {
		return &model.ConversationListOutput{
			List:     []model.ConversationInfo{},
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}, api.CodeSuccess, nil
	}

	// 收集需要批量查询的 ID
	var lastMsgIDs, targetUserIDs []int64
	for _, c := range convs {
		if c.LastMessageID > 0 {
			lastMsgIDs = append(lastMsgIDs, c.LastMessageID)
		}
		if userID == c.User1ID {
			targetUserIDs = append(targetUserIDs, c.User2ID)
		} else {
			targetUserIDs = append(targetUserIDs, c.User1ID)
		}
	}

	// 批量拉取最后一条消息 & 对方用户信息
	msgs, _ := repo.BatchGetMessagesByIDs(ctx, lastMsgIDs)
	users, _ := repo.BatchGetUsersByIDs(ctx, targetUserIDs)

	msgMap := make(map[int64]*model.Message, len(msgs))
	for _, m := range msgs {
		msgMap[m.ID] = m
	}
	userMap := make(map[int64]*model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	// 组装响应
	list := make([]model.ConversationInfo, 0, len(convs))
	for _, c := range convs {
		targetID := c.User2ID
		unreadCount := c.UnreadCount2
		if userID == c.User2ID {
			targetID = c.User1ID
			unreadCount = c.UnreadCount1
		}
		tu := model.MessageUserInfo{}
		if u, ok := userMap[targetID]; ok {
			tu = model.MessageUserInfo{ID: u.ID, Username: u.Username, Nickname: u.Nickname, Avatar: u.Avatar}
		}
		lm := model.LastMessageInfo{}
		if m, ok := msgMap[c.LastMessageID]; ok {
			lm = model.LastMessageInfo{
				ID: m.ID, MessageType: m.MessageType,
				Content: m.Content, MediaURL: m.MediaURL,
				CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}
		list = append(list, model.ConversationInfo{
			TargetUser:  tu,
			LastMessage: lm,
			UnreadCount: unreadCount,
			UpdatedAt:   c.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &model.ConversationListOutput{List: list, Total: total, Page: page, PageSize: pageSize}, api.CodeSuccess, nil
}

// GetMessages 获取与指定用户的私信列表（游标分页，按 created_at 倒序）
func GetMessages(ctx context.Context, loginUserID, targetUserID int64, cursor string, count int) (*model.MessageListOutput, api.ResCode, error) {
	// 验证对方用户存在
	_, err := query.User.WithContext(ctx).Where(query.User.ID.Eq(targetUserID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeUserNotExist, err
		}
		return nil, api.CodeInternalError, err
	}

	var cursorTime *time.Time
	if cursor != "" {
		t, err := time.Parse("2006-01-02 15:04:05", cursor)
		if err != nil {
			return nil, api.CodeInvalidParam, errors.New("cursor 格式错误，应为 2006-01-02 15:04:05")
		}
		cursorTime = &t
	}

	msgs, err := repo.ListMessagesByConversation(ctx, loginUserID, targetUserID, cursorTime, count)
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	hasMore := len(msgs) > count
	if hasMore {
		msgs = msgs[:count]
	}
	var nextCursor string
	if hasMore && len(msgs) > 0 {
		nextCursor = msgs[len(msgs)-1].CreatedAt.Format("2006-01-02 15:04:05")
	}

	list := make([]model.MessageInfo, 0, len(msgs))
	for _, m := range msgs {
		list = append(list, model.MessageInfo{
			ID: m.ID, FromUserID: m.FromUserID, ToUserID: m.ToUserID,
			MessageType: m.MessageType, Content: m.Content, MediaURL: m.MediaURL,
			IsRead: m.IsRead == 1, CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &model.MessageListOutput{Messages: list, NextCursorToken: nextCursor, HasMore: hasMore}, api.CodeSuccess, nil
}

// SendMessage 发送私信（业务校验 + 写消息 + upsert 会话）
func SendMessage(ctx context.Context, fromUserID int64, input *model.SendMessageInput) (*model.MessageInfo, api.ResCode, error) {
	if fromUserID == input.ToUserID {
		return nil, api.CodeCannotSendToSelf, errors.New("不能给自己发消息")
	}
	if input.MessageType == 1 && input.Content == "" {
		return nil, api.CodeInvalidParam, errors.New("文本消息 content 不能为空")
	}
	if (input.MessageType == 2 || input.MessageType == 3) && input.MediaURL == "" {
		return nil, api.CodeInvalidParam, errors.New("图片/视频消息 mediaUrl 不能为空")
	}
	// 验证接收方存在
	_, err := query.User.WithContext(ctx).Where(query.User.ID.Eq(input.ToUserID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeUserNotExist, err
		}
		return nil, api.CodeInternalError, err
	}

	msg := &model.Message{
		FromUserID: fromUserID, ToUserID: input.ToUserID,
		MessageType: input.MessageType, Content: input.Content,
		MediaURL: input.MediaURL, IsRead: 0,
	}
	if err := repo.CreateMessageWithConversation(ctx, msg); err != nil {
		return nil, api.CodeInternalError, err
	}
	return &model.MessageInfo{
		ID: msg.ID, FromUserID: msg.FromUserID, ToUserID: msg.ToUserID,
		MessageType: msg.MessageType, Content: msg.Content, MediaURL: msg.MediaURL,
		IsRead: false, CreatedAt: msg.CreatedAt.Format("2006-01-02 15:04:05"),
	}, api.CodeSuccess, nil
}

// MarkConversationRead 标记会话中对方发给自己的消息为已读，重置未读计数
func MarkConversationRead(ctx context.Context, loginUserID, targetUserID int64) (int64, api.ResCode, error) {
	readCount, err := repo.MarkMessagesRead(ctx, loginUserID, targetUserID)
	if err != nil {
		return 0, api.CodeInternalError, err
	}
	// 忽略会话行不存在时的错误（不影响主流程）
	_ = repo.ResetConversationUnread(ctx, loginUserID, targetUserID)
	return readCount, api.CodeSuccess, nil
}

// DeleteMessage 删除消息（软删除，仅发送方或接收方可操作）
func DeleteMessage(ctx context.Context, loginUserID, messageID int64) (api.ResCode, error) {
	msg, err := repo.FindMessageByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.CodeMessageNotExist, err
		}
		return api.CodeInternalError, err
	}
	if msg.FromUserID != loginUserID && msg.ToUserID != loginUserID {
		return api.CodePermissionDenied, errors.New("无权限删除该消息")
	}
	if err := repo.SoftDeleteMessage(ctx, messageID); err != nil {
		return api.CodeInternalError, err
	}
	return api.CodeSuccess, nil
}

// GetUnreadCount 获取当前用户所有会话的未读私信总数
func GetUnreadCount(ctx context.Context, userID int64) (int64, api.ResCode, error) {
	count, err := repo.SumUnreadCount(ctx, userID)
	if err != nil {
		return 0, api.CodeInternalError, err
	}
	return count, api.CodeSuccess, nil
}
