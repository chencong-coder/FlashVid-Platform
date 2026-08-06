package comment

import (
	"context"
	"errors"
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/comment/v1"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"time"

	"gorm.io/gorm"
)

// 获取评论列表（只返回一级评论，回复需调 GetReplies 接口）
func GetComments(ctx context.Context, userId int64, videoId int64, count int, cursor string) (*model.CommentListOutput, api.ResCode, error) {
	// 1. 检查视频是否存在
	video, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeVideoNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	_ = video // video.UserID 用于 IsAuthored，后面用到
	// 2. 查询一级评论（游标分页）
	q := query.Comment.WithContext(ctx).
		Where(
			query.Comment.VideoID.Eq(videoId),
			query.Comment.ParentID.Eq(0),
			query.Comment.Status.Eq(1),
		).
		Order(query.Comment.CreatedAt.Desc(), query.Comment.ID.Desc()).
		Limit(count)
	if cursor != "" {
		cursorTime, err := time.Parse("2006-01-02 15:04:05", cursor)
		if err != nil {
			return nil, api.CodeInvalidParam, errors.New("invalid cursor format")
		}
		q = q.Where(query.Comment.CreatedAt.Lt(cursorTime))
	}
	commentList, err := q.Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if len(commentList) == 0 {
		return &model.CommentListOutput{
			Comments:        []model.CommentInfo{},
			NextCursorToken: "",
			HasMore:         false,
		}, api.CodeSuccess, nil
	}
	// 3. 收集评论ID和用户ID
	var commentIDs []int64
	var userIDs []int64
	for _, c := range commentList {
		commentIDs = append(commentIDs, c.ID)
		userIDs = append(userIDs, c.UserID)
	}
	// 4. 批量查用户
	userMap := make(map[int64]*model.User)
	users, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(userIDs...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	for _, u := range users {
		userMap[u.ID] = u
	}
	// 5. 批量查点赞状态（需要登录）
	likedMap := make(map[int64]bool)
	if userId > 0 {
		likes, err := query.Like.WithContext(ctx).
			Where(
				query.Like.TargetType.Eq(2),
				query.Like.TargetID.In(commentIDs...),
				query.Like.UserID.Eq(userId),
			).Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		for _, like := range likes {
			likedMap[like.TargetID] = true
		}
	}
	// 6. 构建输出列表
	var commentInfos []model.CommentInfo
	for _, c := range commentList {
		u, ok := userMap[c.UserID]
		if !ok {
			continue // 用户已被删除
		}
		commentInfos = append(commentInfos, model.CommentInfo{
			ID: c.ID,
			User: model.CommentUser{
				ID:       u.ID,
				Username: u.Username,
				Nickname: u.Nickname,
				Avatar:   u.Avatar,
			},
			Content:    c.Content,
			LikeCount:  c.LikeCount,
			ReplyCount: c.ReplyCount,
			IsLiked:    likedMap[c.ID],
			IsAuthored: c.UserID == video.UserID,
			Replies:    []model.ReplyInfo{}, // 回复通过 GetReplies 接口单独加载
			CreatedAt:  c.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 7. 游标分页
	hasMore := len(commentList) == count
	nextCursorToken := ""
	if hasMore {
		nextCursorToken = commentList[len(commentList)-1].CreatedAt.Format("2006-01-02 15:04:05")
	}
	return &model.CommentListOutput{
		Comments:        commentInfos,
		NextCursorToken: nextCursorToken,
		HasMore:         hasMore,
	}, api.CodeSuccess, nil
}

// 获取评论的回复列表
func GetReplies(ctx context.Context, userId int64, commentId int64) ([]model.ReplyInfo, api.ResCode, error) {
	// 1. 检查父评论是否存在
	_, err := query.Comment.WithContext(ctx).
		Where(query.Comment.ID.Eq(commentId), query.Comment.Status.Eq(1)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeCommentNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 查询所有子评论（按时间升序，和发布顺序一致）
	replies, err := query.Comment.WithContext(ctx).
		Where(query.Comment.ParentID.Eq(commentId), query.Comment.Status.Eq(1)).
		Order(query.Comment.CreatedAt.Asc()).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if len(replies) == 0 {
		return []model.ReplyInfo{}, api.CodeSuccess, nil
	}
	// 3. 收集用户ID和被回复用户ID
	var userIDs []int64
	var replyToUserIDs []int64
	var replyIDs []int64
	for _, r := range replies {
		userIDs = append(userIDs, r.UserID)
		replyIDs = append(replyIDs, r.ID)
		if r.ReplyToUserID > 0 {
			replyToUserIDs = append(replyToUserIDs, r.ReplyToUserID)
		}
	}
	// 4. 批量查回复作者
	userMap := make(map[int64]*model.User)
	users, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(userIDs...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	for _, u := range users {
		userMap[u.ID] = u
	}
	// 5. 批量查被回复用户
	replyToUserMap := make(map[int64]*model.User)
	if len(replyToUserIDs) > 0 {
		replyToUsers, err := query.User.WithContext(ctx).
			Where(query.User.ID.In(replyToUserIDs...)).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		for _, u := range replyToUsers {
			replyToUserMap[u.ID] = u
		}
	}
	// 6. 批量查点赞状态
	likedMap := make(map[int64]bool)
	if userId > 0 {
		likes, err := query.Like.WithContext(ctx).
			Where(
				query.Like.TargetType.Eq(2),
				query.Like.TargetID.In(replyIDs...),
				query.Like.UserID.Eq(userId),
			).Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		for _, like := range likes {
			likedMap[like.TargetID] = true
		}
	}
	// 7. 构建输出列表
	var replyInfos []model.ReplyInfo
	for _, r := range replies {
		u, ok := userMap[r.UserID]
		if !ok {
			continue
		}
		var replyTo model.ReplyToUser
		if r.ReplyToUserID > 0 {
			if ru, ok := replyToUserMap[r.ReplyToUserID]; ok {
				replyTo = model.ReplyToUser{ID: ru.ID, Nickname: ru.Nickname}
			}
		}
		replyInfos = append(replyInfos, model.ReplyInfo{
			ID: r.ID,
			User: model.CommentUser{
				ID:       u.ID,
				Username: u.Username,
				Nickname: u.Nickname,
				Avatar:   u.Avatar,
			},
			ReplyTo:   replyTo,
			Content:   r.Content,
			LikeCount: r.LikeCount,
			IsLiked:   likedMap[r.ID],
			CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return replyInfos, api.CodeSuccess, nil
}

// CreateComment 发表评论（一级评论或回复）
func CreateComment(ctx context.Context, userId int64, videoId int64, content string, parentId int64, replyToUserId int64) (*model.CommentInfo, *model.ReplyInfo, api.ResCode, error) {
	// 1. 检查视频是否存在
	video, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, api.CodeVideoNotExist, err
		}
		return nil, nil, api.CodeInternalError, err
	}
	// 2. 如果是回复，检查父评论是否存在
	if parentId > 0 {
		_, err := query.Comment.WithContext(ctx).
			Where(query.Comment.ID.Eq(parentId), query.Comment.Status.Eq(1)).
			First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, api.CodeCommentNotExist, err
			}
			return nil, nil, api.CodeInternalError, err
		}
	}
	// 3. 查评论者信息（用于构建返回值）
	author, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		return nil, nil, api.CodeInternalError, err
	}
	commentUser := model.CommentUser{
		ID:       author.ID,
		Username: author.Username,
		Nickname: author.Nickname,
		Avatar:   author.Avatar,
	}
	// 4. 事务：创建评论 + 更新计数
	// 提前捕获当前时间：GORM gen-dao Create() 不会回写 default:CURRENT_TIMESTAMP 字段
	now := time.Now()
	newComment := &model.Comment{
		VideoID:       videoId,
		UserID:        userId,
		ParentID:      parentId,
		ReplyToUserID: replyToUserId,
		Content:       content,
		Status:        1,
		CreatedAt:     now,
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 插入评论
		if err := tx.Comment.WithContext(ctx).Create(newComment); err != nil {
			return err
		}
		if parentId > 0 {
			// 回复：父评论 reply_count + 1
			_, err = tx.Comment.WithContext(ctx).
				Where(tx.Comment.ID.Eq(parentId)).
				UpdateSimple(tx.Comment.ReplyCount.Add(1))
		} else {
			// 一级评论：视频 comment_count + 1
			_, err = tx.Video.WithContext(ctx).
				Where(tx.Video.ID.Eq(videoId)).
				UpdateSimple(tx.Video.CommentCount.Add(1))
		}
		return err
	})
	if err != nil {
		return nil, nil, api.CodeInternalError, err
	}
	createdAt := newComment.CreatedAt.Format("2006-01-02 15:04:05")
	// 5. 构建并返回评论数据
	if parentId == 0 {
		commentInfo := &model.CommentInfo{
			ID:         newComment.ID,
			Content:    newComment.Content,
			User:       commentUser,
			LikeCount:  0,
			ReplyCount: 0,
			IsLiked:    false,
			IsAuthored: userId == video.UserID,
			Replies:    []model.ReplyInfo{},
			CreatedAt:  createdAt,
		}
		return commentInfo, nil, api.CodeSuccess, nil
	}
	// 回复：查被回复用户信息
	var replyTo model.ReplyToUser
	if replyToUserId > 0 {
		if ru, err := query.User.WithContext(ctx).
			Where(query.User.ID.Eq(replyToUserId)).
			First(); err == nil {
			replyTo = model.ReplyToUser{ID: ru.ID, Nickname: ru.Nickname}
		}
	}
	replyInfo := &model.ReplyInfo{
		ID:        newComment.ID,
		Content:   newComment.Content,
		User:      commentUser,
		ReplyTo:   replyTo,
		LikeCount: 0,
		IsLiked:   false,
		CreatedAt: createdAt,
	}
	return nil, replyInfo, api.CodeSuccess, nil
}

// DeleteComment 删除评论（软删除，status=2）
func DeleteComment(ctx context.Context, userId int64, commentId int64) (api.ResCode, error) {
	// 1. 查评论
	c, err := query.Comment.WithContext(ctx).
		Where(query.Comment.ID.Eq(commentId), query.Comment.Status.Eq(1)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.CodeCommentNotExist, err
		}
		return api.CodeInternalError, err
	}
	// 2. 只有作者才能删除自己的评论
	if c.UserID != userId {
		return api.CodePermissionDenied, errors.New("无权删除他人评论")
	}
	// 3. 事务：软删除 + 更新计数
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 软删除：status 置为 2
		if _, err := tx.Comment.WithContext(ctx).
			Where(tx.Comment.ID.Eq(commentId)).
			UpdateSimple(tx.Comment.Status.Value(2)); err != nil {
			return err
		}
		// 同时更新 deleted_at（GORM 软删除，方便审计和数据恢复）
		if _, err := tx.Comment.WithContext(ctx).
			Where(tx.Comment.ID.Eq(commentId)).
			Delete(); err != nil {
			return err
		}
		if c.ParentID > 0 {
			// 回复：父评论 reply_count - 1
			_, err = tx.Comment.WithContext(ctx).
				Where(tx.Comment.ID.Eq(c.ParentID)).
				UpdateSimple(tx.Comment.ReplyCount.Sub(1))
		} else {
			// 一级评论：视频 comment_count - 1
			_, err = tx.Video.WithContext(ctx).
				Where(tx.Video.ID.Eq(c.VideoID)).
				UpdateSimple(tx.Video.CommentCount.Sub(1))
		}
		return err
	})
	if err != nil {
		return api.CodeInternalError, err
	}
	return api.CodeSuccess, nil
}

// LikeComment 点赞评论（target_type=2）
func LikeComment(ctx context.Context, userId int64, commentId int64) (*v1.LikeCommentResp, api.ResCode, error) {
	// 1. 检查评论是否存在
	c, err := query.Comment.WithContext(ctx).
		Where(query.Comment.ID.Eq(commentId), query.Comment.Status.Eq(1)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeCommentNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 检查是否已点赞
	count, err := query.Like.WithContext(ctx).
		Where(
			query.Like.UserID.Eq(userId),
			query.Like.TargetType.Eq(2),
			query.Like.TargetID.Eq(commentId),
		).Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if count > 0 {
		return nil, api.CodeAlreadyLiked, errors.New("已点赞该评论")
	}
	// 3. 事务：创建点赞记录 + comment.like_count+1
	err = query.Q.Transaction(func(tx *query.Query) error {
		if err := tx.Like.WithContext(ctx).Create(&model.Like{
			UserID:     userId,
			TargetType: 2,
			TargetID:   commentId,
		}); err != nil {
			return err
		}
		_, err = tx.Comment.WithContext(ctx).
			Where(tx.Comment.ID.Eq(commentId)).
			UpdateSimple(tx.Comment.LikeCount.Add(1))
		return err
	})
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	return &v1.LikeCommentResp{
		IsLiked:   true,
		LikeCount: c.LikeCount + 1,
	}, api.CodeSuccess, nil
}

// UnlikeComment 取消点赞评论
func UnlikeComment(ctx context.Context, userId int64, commentId int64) (*v1.LikeCommentResp, api.ResCode, error) {
	// 1. 检查评论是否存在
	c, err := query.Comment.WithContext(ctx).
		Where(query.Comment.ID.Eq(commentId), query.Comment.Status.Eq(1)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeCommentNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 检查是否已点赞
	count, err := query.Like.WithContext(ctx).
		Where(
			query.Like.UserID.Eq(userId),
			query.Like.TargetType.Eq(2),
			query.Like.TargetID.Eq(commentId),
		).Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if count == 0 {
		return nil, api.CodeAlreadyUnliked, errors.New("未点赞该评论")
	}
	// 3. 事务：删除点赞记录 + comment.like_count-1
	err = query.Q.Transaction(func(tx *query.Query) error {
		if _, err := tx.Like.WithContext(ctx).
			Where(
				tx.Like.UserID.Eq(userId),
				tx.Like.TargetType.Eq(2),
				tx.Like.TargetID.Eq(commentId),
			).Delete(); err != nil {
			return err
		}
		_, err = tx.Comment.WithContext(ctx).
			Where(tx.Comment.ID.Eq(commentId)).
			UpdateSimple(tx.Comment.LikeCount.Sub(1))
		return err
	})
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	return &v1.LikeCommentResp{
		IsLiked:   false,
		LikeCount: c.LikeCount - 1,
	}, api.CodeSuccess, nil
}
