package notification

import (
	"context"

	v1 "flashvid-platform-gin/api/notification/v1"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
)

// CreateNotification 写入一条通知（fire-and-forget，用于在事务外异步调用或事务内直接调用）
func CreateNotification(ctx context.Context, tx *query.Query, n *model.Notification) {
	if tx != nil {
		_ = tx.Notification.WithContext(ctx).Create(n)
	} else {
		_ = query.Notification.WithContext(ctx).Create(n)
	}
}

// GetNotifications 获取通知列表（分页 + 按类型筛选）
func GetNotifications(ctx context.Context, userId int64, actionTypes []int32, page, pageSize int) ([]v1.NotificationInfo, int64, error) {
	q := query.Notification.WithContext(ctx).
		Where(query.Notification.UserID.Eq(userId))

	if len(actionTypes) > 0 {
		q = q.Where(query.Notification.ActionType.In(actionTypes...))
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	notifications, err := q.Order(query.Notification.CreatedAt.Desc()).
		Offset(offset).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}
	if len(notifications) == 0 {
		return []v1.NotificationInfo{}, total, nil
	}

	// 批量加载 actor 用户信息
	actorIDs := make([]int64, 0, len(notifications))
	for _, n := range notifications {
		actorIDs = append(actorIDs, n.ActorID)
	}
	actors, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(actorIDs...)).
		Find()
	if err != nil {
		return nil, 0, err
	}
	actorMap := make(map[int64]*model.User, len(actors))
	for _, u := range actors {
		actorMap[u.ID] = u
	}

	// 批量加载视频信息（target_type=2）
	videoIDs := make([]int64, 0)
	for _, n := range notifications {
		if n.TargetType == 2 {
			videoIDs = append(videoIDs, n.TargetID)
		}
	}
	videoMap := make(map[int64]*model.Video)
	if len(videoIDs) > 0 {
		videos, err := query.Video.WithContext(ctx).
			Where(query.Video.ID.In(videoIDs...)).
			Find()
		if err == nil {
			for _, v := range videos {
				videoMap[v.ID] = v
			}
		}
	}

	// 组装 DTO
	result := make([]v1.NotificationInfo, 0, len(notifications))
	for _, n := range notifications {
		info := v1.NotificationInfo{
			ID:         n.ID,
			ActorID:    n.ActorID,
			ActionType: n.ActionType,
			TargetID:   n.TargetID,
			Content:    n.Content,
			IsRead:     n.IsRead,
			CreatedAt:  n.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if actor, ok := actorMap[n.ActorID]; ok {
			info.ActorName = actor.Nickname
			if info.ActorName == "" {
				info.ActorName = actor.Username
			}
			info.ActorAvatar = actor.Avatar
		}
		if n.TargetType == 2 {
			if video, ok := videoMap[n.TargetID]; ok {
				info.TargetTitle = video.Title
				info.TargetCover = video.CoverURL
			}
		}
		result = append(result, info)
	}
	return result, total, nil
}

// GetUnreadCounts 按通知类型统计未读数
func GetUnreadCounts(ctx context.Context, userId int64) (*v1.UnreadCountsResp, error) {
	type CountRow struct {
		ActionType int32 `gorm:"column:action_type"`
		Count      int64 `gorm:"column:cnt"`
	}
	var rows []CountRow
	err := query.Notification.WithContext(ctx).
		Where(query.Notification.UserID.Eq(userId), query.Notification.IsRead.Eq(0)).
		Select(query.Notification.ActionType, query.Notification.ID.Count().As("cnt")).
		Group(query.Notification.ActionType).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	resp := &v1.UnreadCountsResp{}
	for _, row := range rows {
		switch row.ActionType {
		case 1:
			resp.Followers += row.Count
		case 2, 3:
			resp.LikesAndFavs += row.Count
		case 4, 5:
			resp.Comments += row.Count
		case 6:
			resp.Mentions += row.Count
		}
	}
	return resp, nil
}

// MarkAsRead 标记已读（指定类型，空=全部）
func MarkAsRead(ctx context.Context, userId int64, actionTypes []int32) error {
	q := query.Notification.WithContext(ctx).
		Where(query.Notification.UserID.Eq(userId), query.Notification.IsRead.Eq(0))
	if len(actionTypes) > 0 {
		q = q.Where(query.Notification.ActionType.In(actionTypes...))
	}
	_, err := q.UpdateSimple(query.Notification.IsRead.Value(1))
	return err
}
