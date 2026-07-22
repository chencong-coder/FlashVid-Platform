package user

import (
	"context"
	"errors"
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/user/v1"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"time"

	"gorm.io/gorm"
)

// 获取用户信息服务
func GetUserInfo(ctx context.Context, userId int64) (*model.UserInfoOutput, api.ResCode, error) {
	// 1. 根据userId查询用户
	user, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在是预期的业务错误，不返回底层错误
			return nil, api.CodeUserNotExist, nil
		}
		return nil, api.CodeInternalError, err
	}

	// 2. 返回用户信息
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

// 更新用户信息服务
func UpdateUserInfo(ctx context.Context, userId int64, req *v1.UpdateUserInfoReq) (*model.UpdateUserInfoOutput, api.ResCode, error) {
	// 1. 检查用户是否存在
	user, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在是预期的业务错误，不返回底层错误
			return nil, api.CodeUserNotExist, nil
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 构建更新字段
	updateData := make(map[string]interface{})
	if req.Nickname != nil {
		updateData["nickname"] = *req.Nickname
	}
	if req.Avatar != nil {
		updateData["avatar"] = *req.Avatar
	}
	if req.Bio != nil {
		updateData["bio"] = *req.Bio
	}
	if req.City != nil {
		updateData["city"] = *req.City
	}
	if req.Gender != nil {
		updateData["gender"] = *req.Gender
	}
	if req.Birthday != nil {
		updateData["birthday"] = *req.Birthday
	}
	if req.Email != nil {
		updateData["email"] = *req.Email
	}
	if req.Phone != nil {
		updateData["phone"] = *req.Phone
	}

	// 3. 如果没有要更新的字段，直接返回当前用户信息
	if len(updateData) == 0 {
		return &model.UpdateUserInfoOutput{
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
			UpdatedAt:      user.UpdatedAt,
		}, api.CodeSuccess, nil
	}

	// 4. 执行更新
	updateData["updated_at"] = time.Now() // 更新更新时间
	_, err = query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		Updates(updateData)
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 5. 返回更新后的用户信息
	updatedUser, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	return &model.UpdateUserInfoOutput{
		UserId:         updatedUser.ID,
		Username:       updatedUser.Username,
		Nickname:       updatedUser.Nickname,
		Avatar:         updatedUser.Avatar,
		Bio:            updatedUser.Bio,
		City:           updatedUser.City,
		FollowersCount: updatedUser.FollowerCount,
		FollowingCount: updatedUser.FollowingCount,
		VideosCount:    updatedUser.VideoCount,
		LikesCount:     updatedUser.LikeCount,
		Phone:          updatedUser.Phone,
		Gender:         updatedUser.Gender,
		Birthday:       updatedUser.Birthday,
		Email:          updatedUser.Email,
		UpdatedAt:      updatedUser.UpdatedAt,
	}, api.CodeSuccess, nil
}
