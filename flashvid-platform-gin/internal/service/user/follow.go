package user

import (
	"context"
	"errors"
	"time"

	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	notifSvc "flashvid-platform-gin/internal/service/notification"
	"gorm.io/gorm"
)

// FollowUser 关注用户（幂等：已关注直接返回成功；事务内同步更新计数）
func FollowUser(ctx context.Context, loginUserId int64, followUserId int64) (bool, api.ResCode, error) {
	// 1. 检测目标用户是否存在
	_, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(followUserId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, api.CodeUserNotExist, err
		}
		return false, api.CodeInternalError, err
	}
	// 2. 检查是否已经关注（幂等：已关注直接返回成功，不报错）
	existing, err := query.Follow.WithContext(ctx).
		Where(query.Follow.FollowerID.Eq(loginUserId), query.Follow.FollowingID.Eq(followUserId)).
		First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, api.CodeInternalError, err
	}
	if existing != nil {
		return true, api.CodeSuccess, nil
	}
	// 3. 事务：创建关注记录 + 更新双方计数 + 写通知
	// 提前捕获时间：GORM gen-dao Create() 不会回写 default:CURRENT_TIMESTAMP 字段
	now := time.Now()
	txErr := query.Q.Transaction(func(tx *query.Query) error {
		if err := tx.Follow.WithContext(ctx).Create(&model.Follow{
			FollowerID:  loginUserId,
			FollowingID: followUserId,
			CreatedAt:   now,
		}); err != nil {
			return err
		}
		// 增加登录用户的关注数
		if _, err := tx.User.WithContext(ctx).
			Where(tx.User.ID.Eq(loginUserId)).
			UpdateSimple(tx.User.FollowingCount.Add(1)); err != nil {
			return err
		}
		// 增加被关注用户的粉丝数
		if _, err := tx.User.WithContext(ctx).
			Where(tx.User.ID.Eq(followUserId)).
			UpdateSimple(tx.User.FollowerCount.Add(1)); err != nil {
			return err
		}
		// 通知被关注者
		notifSvc.CreateNotification(ctx, tx, &model.Notification{
			UserID:     followUserId,
			ActorID:    loginUserId,
			ActionType: 1, // 关注
			TargetType: 1, // 目标是用户
			TargetID:   followUserId,
		})
		return nil
	})
	if txErr != nil {
		return false, api.CodeInternalError, txErr
	}
	return true, api.CodeSuccess, nil
}

// UnfollowUser 取消关注用户（幂等：未关注直接返回成功；硬删除避免软删后再关注时unique冲突；事务内同步更新计数）
func UnfollowUser(ctx context.Context, loginUserId int64, followUserId int64) (bool, api.ResCode, error) {
	// 1. 检测目标用户是否存在
	_, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(followUserId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, api.CodeUserNotExist, err
		}
		return false, api.CodeInternalError, err
	}
	// 2. 检查是否确实关注了（幂等：未关注直接返回成功，不报错）
	existing, err := query.Follow.WithContext(ctx).
		Where(query.Follow.FollowerID.Eq(loginUserId), query.Follow.FollowingID.Eq(followUserId)).
		First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, api.CodeInternalError, err
	}
	if existing == nil {
		return false, api.CodeSuccess, nil
	}
	// 3. 事务：删除关注记录 + 更新双方计数
	// Follow 已无 soft-delete，直接物理删除，不再需要 Unscoped
	txErr := query.Q.Transaction(func(tx *query.Query) error {
		if _, err := tx.Follow.WithContext(ctx).
			Where(tx.Follow.FollowerID.Eq(loginUserId), tx.Follow.FollowingID.Eq(followUserId)).
			Delete(); err != nil {
			return err
		}
		// 减少登录用户的关注数（WHERE following_count > 0 防止变负）
		if _, err := tx.User.WithContext(ctx).
			Where(tx.User.ID.Eq(loginUserId), tx.User.FollowingCount.Gt(0)).
			UpdateSimple(tx.User.FollowingCount.Add(-1)); err != nil {
			return err
		}
		// 减少被关注用户的粉丝数（WHERE follower_count > 0 防止变负）
		if _, err := tx.User.WithContext(ctx).
			Where(tx.User.ID.Eq(followUserId), tx.User.FollowerCount.Gt(0)).
			UpdateSimple(tx.User.FollowerCount.Add(-1)); err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return false, api.CodeInternalError, txErr
	}
	return false, api.CodeSuccess, nil
}

// GetUserFollowers 获取用户粉丝列表
func GetUserFollowers(ctx context.Context, userId int64, page int, pageSize int) (*model.UserFollowersOutput, api.ResCode, error) {
	// 1. 检测用户是否存在
	_, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeUserNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 获取粉丝列表（分页）
	followers, err := query.Follow.WithContext(ctx).
		Where(query.Follow.FollowingID.Eq(userId)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 3. 获取粉丝总数
	totalCount, err := query.Follow.WithContext(ctx).
		Where(query.Follow.FollowingID.Eq(userId)).
		Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if len(followers) == 0 {
		return &model.UserFollowersOutput{
			Followers: []model.UserInfo{},
			Pagination: model.Pagination{
				Page:       page,
				PageSize:   pageSize,
				Total:      totalCount,
				TotalPages: int((totalCount + int64(pageSize) - 1) / int64(pageSize)),
			},
		}, api.CodeSuccess, nil
	}
	// 4. 批量查询粉丝用户信息
	followerIds := make([]int64, 0, len(followers))
	for _, f := range followers {
		followerIds = append(followerIds, f.FollowerID)
	}
	users, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(followerIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 构建 userId -> user 映射
	userMap := make(map[int64]*model.User)
	for i := range users {
		userMap[users[i].ID] = users[i]
	}
	// 5. 按粉丝关注顺序组装结果
	var userInfos []model.UserInfo
	for _, follower := range followers {
		u, ok := userMap[follower.FollowerID]
		if !ok {
			continue // 用户可能已被删除
		}
		userInfos = append(userInfos, model.UserInfo{
			UserId:         u.ID,
			Username:       u.Username,
			Gender:         u.Gender,
			Nickname:       u.Nickname,
			Avatar:         u.Avatar,
			Bio:            u.Bio,
			City:           u.City,
			FollowersCount: u.FollowerCount,
			FollowingCount: u.FollowingCount,
			VideosCount:    u.VideoCount,
			LikesCount:     u.LikeCount,
			Phone:          u.Phone,
			Birthday:       u.Birthday.Format("2006-01-02"),
			Email:          u.Email,
			CreatedAt:      u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 6. 返回结果
	return &model.UserFollowersOutput{
		Followers: userInfos,
		Pagination: model.Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      totalCount,
			TotalPages: int((totalCount + int64(pageSize) - 1) / int64(pageSize)),
		},
	}, api.CodeSuccess, nil
}

// GetUserFollowing 获取用户关注列表
func GetUserFollowing(ctx context.Context, userId int64, page int, pageSize int) (*model.UserFollowingOutput, api.ResCode, error) {
	// 1. 检测用户是否存在
	_, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeUserNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 获取关注列表
	followings, err := query.Follow.WithContext(ctx).
		Where(query.Follow.FollowerID.Eq(userId)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 3. 查询关注总数
	totalCount, err := query.Follow.WithContext(ctx).
		Where(query.Follow.FollowerID.Eq(userId)).
		Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if len(followings) == 0 {
		return &model.UserFollowingOutput{
			Followings: []model.UserInfo{},
			Pagination: model.Pagination{
				Page:       page,
				PageSize:   pageSize,
				Total:      totalCount,
				TotalPages: int((totalCount + int64(pageSize) - 1) / int64(pageSize)),
			},
		}, api.CodeSuccess, nil
	}
	// 3. 构建followingids -> 用户列表的映射
	followingIds := make([]int64, 0, len(followings))
	for _, following := range followings {
		followingIds = append(followingIds, following.FollowingID)
	}
	// 一次性查询所有被关注用户信息
	users, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(followingIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	followingIdsMap := make(map[int64]*model.User)
	for i := range users {
		followingIdsMap[users[i].ID] = users[i]
	}
	// 4. 构建关注列表输出
	var userInfos []model.UserInfo
	for _, following := range followings {
		user, exists := followingIdsMap[following.FollowingID]
		if !exists {
			continue // 如果用户不存在，跳过
		}
		userInfos = append(userInfos, model.UserInfo{
			UserId:         user.ID,
			Username:       user.Username,
			Gender:		 user.Gender,
			Nickname: 	 user.Nickname,
			Avatar:         user.Avatar,
			Bio:            user.Bio,
			City:           user.City,
			FollowersCount: user.FollowerCount,
			FollowingCount: user.FollowingCount,
			VideosCount:    user.VideoCount,
			LikesCount:     user.LikeCount,
			Phone:          user.Phone,
			Birthday:       user.Birthday.Format("2006-01-02"),
			Email:          user.Email,
			CreatedAt:      user.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 5. 返回结果
	return &model.UserFollowingOutput{
		Followings: userInfos,
		Pagination: model.Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      totalCount,
			TotalPages: int((totalCount + int64(pageSize) - 1) / int64(pageSize)),
		},
	}, api.CodeSuccess, nil
}
