package user

import (
	"context"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"errors"
	"gorm.io/gorm"
	"flashvid-platform-gin/internal/model"
)

func FollowUser(ctx context.Context, loginUserId int64, followUserId int64) (bool, api.ResCode, error) {
	// 1. 检测用户是否存在
	_, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(followUserId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, api.CodeUserNotExist, err
		}
		return false, api.CodeInternalError, err
	}
	// 2. 检查是否已经关注
	followed, err := query.Follow.WithContext(ctx).
		Where(query.Follow.FollowerID.Eq(loginUserId), query.Follow.FollowingID.Eq(followUserId)).
		First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, api.CodeInternalError, err
	}
	if followed != nil {
		// 已经关注，返回特定错误码告知前端
		return true, api.CodeAlreadyFollowed, errors.New("already following")
	}
	// 3. 如果未关注，则创建关注记录
	newFollow := &model.Follow{
		FollowerID:  loginUserId,
		FollowingID: followUserId,
	}
	if err := query.Follow.WithContext(ctx).Create(newFollow); err != nil {
		return false, api.CodeInternalError, err
	}
	return true, api.CodeSuccess, nil
}