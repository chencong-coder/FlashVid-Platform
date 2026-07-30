package message

import (
	"context"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"time"
)

// CountConversations 统计用户的会话总数
func CountConversations(ctx context.Context, userID int64) (int64, error) {
	return query.Conversation.WithContext(ctx).
		Where(query.Conversation.User1ID.Eq(userID)).
		Or(query.Conversation.User2ID.Eq(userID)).
		Count()
}

// ListConversations 分页查询用户的会话（按 updated_at 倒序）
func ListConversations(ctx context.Context, userID int64, offset, limit int) ([]*model.Conversation, error) {
	return query.Conversation.WithContext(ctx).
		Where(query.Conversation.User1ID.Eq(userID)).
		Or(query.Conversation.User2ID.Eq(userID)).
		Order(query.Conversation.UpdatedAt.Desc()).
		Offset(offset).Limit(limit).
		Find()
}

// BatchGetMessagesByIDs 批量按 ID 查消息（用于会话列表取最后一条消息摘要）
func BatchGetMessagesByIDs(ctx context.Context, ids []int64) ([]*model.Message, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return query.Message.WithContext(ctx).
		Where(query.Message.ID.In(ids...)).
		Find()
}

// BatchGetUsersByIDs 批量按 ID 查用户（用于会话列表取对方用户信息）
func BatchGetUsersByIDs(ctx context.Context, ids []int64) ([]*model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return query.User.WithContext(ctx).
		Where(query.User.ID.In(ids...)).
		Find()
}

// ListMessagesByConversation 游标分页查两人之间的消息（按 created_at 倒序）
// cursorTime 为 nil 时查第一页；多取 1 条用于判断 hasMore
func ListMessagesByConversation(ctx context.Context, loginUserID, targetUserID int64, cursorTime *time.Time, count int) ([]*model.Message, error) {
	q := query.Message.WithContext(ctx).
		Where(
			query.Message.FromUserID.Eq(loginUserID),
			query.Message.ToUserID.Eq(targetUserID),
		).Or(
		query.Message.FromUserID.Eq(targetUserID),
		query.Message.ToUserID.Eq(loginUserID),
	).Order(query.Message.CreatedAt.Desc(), query.Message.ID.Desc()).
		Limit(count + 1)
	if cursorTime != nil {
		q = q.Where(query.Message.CreatedAt.Lt(*cursorTime))
	}
	return q.Find()
}

// FindMessageByID 按主键查单条消息（未找到返回 gorm.ErrRecordNotFound）
func FindMessageByID(ctx context.Context, id int64) (*model.Message, error) {
	return query.Message.WithContext(ctx).Where(query.Message.ID.Eq(id)).First()
}

// SumUnreadCount 汇总当前用户在所有会话中的未读消息总数
func SumUnreadCount(ctx context.Context, userID int64) (int64, error) {
	var total int64
	rawSQL := `SELECT COALESCE(SUM(CASE WHEN user1_id = ? THEN unread_count1 ELSE unread_count2 END), 0)
FROM conversations WHERE user1_id = ? OR user2_id = ?`
	err := query.Conversation.WithContext(ctx).UnderlyingDB().
		Raw(rawSQL, userID, userID, userID).Scan(&total).Error
	return total, err
}

// CreateMessageWithConversation 事务：写消息 + upsert 会话（更新最后消息 & 接收方未读数）
func CreateMessageWithConversation(ctx context.Context, msg *model.Message) error {
	fromUserID := msg.FromUserID
	toUserID := msg.ToUserID
	return query.Q.Transaction(func(tx *query.Query) error {
		if err := tx.Message.WithContext(ctx).Create(msg); err != nil {
			return err
		}
		user1ID := min(fromUserID, toUserID)
		user2ID := max(fromUserID, toUserID)
		var rawSQL string
		if fromUserID < toUserID {
			// 发送方是 user1，接收方是 user2 → unread_count2+1
			rawSQL = `INSERT INTO conversations (user1_id, user2_id, last_message_id, unread_count2, updated_at, created_at)
VALUES (?, ?, ?, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE last_message_id = VALUES(last_message_id), unread_count2 = unread_count2 + 1, updated_at = NOW()`
		} else {
			// 发送方是 user2，接收方是 user1 → unread_count1+1
			rawSQL = `INSERT INTO conversations (user1_id, user2_id, last_message_id, unread_count1, updated_at, created_at)
VALUES (?, ?, ?, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE last_message_id = VALUES(last_message_id), unread_count1 = unread_count1 + 1, updated_at = NOW()`
		}
		return tx.Message.WithContext(ctx).UnderlyingDB().Exec(rawSQL, user1ID, user2ID, msg.ID).Error
	})
}

// MarkMessagesRead 将指定会话中对方发给自己的未读消息全部标为已读，返回更新行数
func MarkMessagesRead(ctx context.Context, toUserID, fromUserID int64) (int64, error) {
	result, err := query.Message.WithContext(ctx).
		Where(query.Message.ToUserID.Eq(toUserID)).
		Where(query.Message.FromUserID.Eq(fromUserID)).
		Where(query.Message.IsRead.Eq(0)).
		UpdateSimple(query.Message.IsRead.Value(1))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

// ResetConversationUnread 将当前用户在指定会话中的未读计数清零
func ResetConversationUnread(ctx context.Context, loginUserID, targetUserID int64) error {
	user1ID := min(loginUserID, targetUserID)
	user2ID := max(loginUserID, targetUserID)
	var rawSQL string
	if loginUserID < targetUserID {
		rawSQL = `UPDATE conversations SET unread_count1 = 0 WHERE user1_id = ? AND user2_id = ?`
	} else {
		rawSQL = `UPDATE conversations SET unread_count2 = 0 WHERE user1_id = ? AND user2_id = ?`
	}
	return query.Conversation.WithContext(ctx).UnderlyingDB().Exec(rawSQL, user1ID, user2ID).Error
}

// SoftDeleteMessage 软删除消息（设置 deleted_at）
func SoftDeleteMessage(ctx context.Context, id int64) error {
	_, err := query.Message.WithContext(ctx).Where(query.Message.ID.Eq(id)).Delete()
	return err
}
