package notification

import (
	"context"
	"fmt"
	"strconv"

	v1 "flashvid-platform-gin/api/notification/v1"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
)

var rdb = dao.RedisClient

// CreateNotification 写入一条通知（fire-and-forget，用于在事务外异步调用或事务内直接调用）
func CreateNotification(ctx context.Context, tx *query.Query, n *model.Notification) {
	if tx != nil {
		_ = tx.Notification.WithContext(ctx).Create(n)
	} else {
		_ = query.Notification.WithContext(ctx).Create(n)
	}

	// 增加 Redis 未读数缓存
	if rdb != nil {
		cacheKey := fmt.Sprintf("user:%d:unread", n.UserID)
		fieldName := fmt.Sprintf("action_%d", n.ActionType)
		rdb.HIncrBy(ctx, cacheKey, fieldName, 1)
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

// GetUnreadCountsWithCache 按通知类型统计未读数（带缓存）
func GetUnreadCountsWithCache(ctx context.Context, userId int64) (*v1.UnreadCountsResp, error) {
	// Redis 不可用时降级
	if rdb == nil {
		return GetUnreadCounts(ctx, userId)
	}

	cacheKey := fmt.Sprintf("user:%d:unread", userId)

	// 1. 从 Redis Hash 读取未读数
	counts, err := rdb.HGetAll(ctx, cacheKey).Result()
	if err == nil && len(counts) > 0 {
		// 缓存命中，解析数据
		resp := &v1.UnreadCountsResp{}
		for field, val := range counts {
			count, _ := strconv.ParseInt(val, 10, 64)
			switch field {
			case "action_1": // 关注
				resp.Followers += count
			case "action_2", "action_3": // 点赞、收藏
				resp.LikesAndFavs += count
			case "action_4", "action_5": // 评论、回复
				resp.Comments += count
			case "action_6": // @提及
				resp.Mentions += count
			}
		}
		return resp, nil
	}

	// 2. 缓存未命中或 Redis 异常，查 DB
	resp, err := GetUnreadCounts(ctx, userId)
	if err != nil {
		return nil, err
	}

	// 3. 回写缓存（初始化 Hash）
	if resp.Followers > 0 {
		rdb.HSet(ctx, cacheKey, "action_1", resp.Followers)
	}
	if resp.LikesAndFavs > 0 {
		// 注意：这里无法区分 action_2 和 action_3，统一写入 action_2
		rdb.HSet(ctx, cacheKey, "action_2", resp.LikesAndFavs)
	}
	if resp.Comments > 0 {
		rdb.HSet(ctx, cacheKey, "action_4", resp.Comments)
	}
	if resp.Mentions > 0 {
		rdb.HSet(ctx, cacheKey, "action_6", resp.Mentions)
	}

	return resp, nil
}

// MarkAsRead 标记已读（指定类型，空=全部）
func MarkAsRead(ctx context.Context, userId int64, actionTypes []int32) error {
	// 先查询要标记的未读数量（用于更新 Redis）
	q := query.Notification.WithContext(ctx).
		Where(query.Notification.UserID.Eq(userId), query.Notification.IsRead.Eq(0))
	if len(actionTypes) > 0 {
		q = q.Where(query.Notification.ActionType.In(actionTypes...))
	}

	// 统计每个类型的未读数
	type CountRow struct {
		ActionType int32 `gorm:"column:action_type"`
		Count      int64 `gorm:"column:cnt"`
	}
	var rows []CountRow
	_ = q.Select(query.Notification.ActionType, query.Notification.ID.Count().As("cnt")).
		Group(query.Notification.ActionType).
		Scan(&rows)

	// 执行标记已读
	_, err := q.UpdateSimple(query.Notification.IsRead.Value(1))
	if err != nil {
		return err
	}

	// 更新 Redis 缓存（减少未读数）
	if rdb != nil && len(rows) > 0 {
		cacheKey := fmt.Sprintf("user:%d:unread", userId)
		for _, row := range rows {
			fieldName := fmt.Sprintf("action_%d", row.ActionType)
			newVal := rdb.HIncrBy(ctx, cacheKey, fieldName, -row.Count).Val()
			// 如果减到 0 或负数，删除该字段
			if newVal <= 0 {
				rdb.HDel(ctx, cacheKey, fieldName)
			}
		}
	}

	return nil
}
