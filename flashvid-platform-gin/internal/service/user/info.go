package user

import (
	"context"
	"errors"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"gorm.io/gorm"
)

// 获取用户信息服务
func GetUserInfo(ctx context.Context, userId int64) (*model.UserInfoOutput, api.ResCode, error) {
	// 1.根据userid来查询判断用户是否存在
	user, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 2.如果用户不存在，返回错误码
			return nil, api.CodeUserNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 3.如果用户存在，返回用户信息
	return &model.UserInfoOutput{
		UserId:         user.ID,
		Username:       user.Username,
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		Bio:            user.Bio,
		City:           user.City,
		FollowersCount: user.FollowerCount,
		FollowingCount: user.FollowingCount,
		VideosCount:    user.VideoCount,
		LikesCount:     user.LikeCount,
		Phone:          user.Phone,
		Gender:         user.Gender,
		Birthday:       user.Birthday,
		Email:          user.Email,
		CreatedAt:      user.CreatedAt,
	}, api.CodeSuccess, nil
}