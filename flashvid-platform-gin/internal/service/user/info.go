package user

import (
	"context"
	"encoding/json"
	"errors"
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/user/v1"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/internal/model"
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var rdb = dao.RedisClient
var userInfoGroup singleflight.Group

// 获取用户信息服务
func GetUserInfo(ctx context.Context, userId int64) (*model.UserInfoOutput, api.ResCode, error) {
	// 1. 根据userId查询用户
	user, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeUserNotExist, nil
		}
		return nil, api.CodeInternalError, err
	}

	// 2. 检查当前登录用户是否关注了目标用户（未登录则为 false）
	isFollowing := false
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if loginUserIdVal, exists := ginCtx.Get(middleware.CtxKeyUserID); exists {
			if loginUserId, ok := loginUserIdVal.(int64); ok && loginUserId > 0 && loginUserId != userId {
				cnt, err := query.Follow.WithContext(ctx).
					Where(query.Follow.FollowerID.Eq(loginUserId), query.Follow.FollowingID.Eq(userId)).
					Count()
				if err == nil && cnt > 0 {
					isFollowing = true
				}
			}
		}
	}

	// 3. 返回用户信息
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
		IsFollowing:    isFollowing,
	}, api.CodeSuccess, nil
}

// GetUserInfoWithCache 获取用户信息服务（带缓存 + 三大问题防护）
func GetUserInfoWithCache(ctx context.Context, userId int64) (*model.UserInfoOutput, api.ResCode, error) {
	// redis不可用的时候降级
	if rdb == nil {
		return GetUserInfo(ctx, userId)
	}
	cacheKey := fmt.Sprintf("user:%d", userId)

	// 1.查redis缓存
	cacheData, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// 缓存命中
		if cacheData == "null" {
			// 缓存中存储了用户不存在的标记
			return nil, api.CodeUserNotExist, errors.New("user not exist")
		}
		// 反序列化缓存数据
		var output model.UserInfoOutput
		if json.Unmarshal([]byte(cacheData), &output) == nil {
			return &output, api.CodeSuccess, nil
		}
		// 如果反序列化失败 回源
	} else if err != redis.Nil {
		// Redis网络/服务异常，不是缓存miss，禁止穿透DB，直接降级报错
		return nil, api.CodeInternalError, fmt.Errorf("redis get failed: %w", err)
	}

	// err == redis.Nil：真正缓存未命中，进入singleflight

	// 2.缓存未命中，使用singleflight防止缓存击穿
	result, err, shared := userInfoGroup.Do(cacheKey, func() (interface{}, error) {
		// 回调内部二次检查缓存
		cacheDateTwice, innerErr := rdb.Get(ctx, cacheKey).Result()
		if innerErr == nil {
			if cacheDateTwice == "null" {
				return nil, errors.New("user not exist")
			}
			var outputTwice model.UserInfoOutput
			if json.Unmarshal([]byte(cacheDateTwice), &outputTwice) == nil {
				return &outputTwice, nil
			}
		}

		// 缓存真没数据 查数据库
		outputSql, code, err := GetUserInfo(ctx, userId)
		if err != nil || code != api.CodeSuccess {
			if code == api.CodeUserNotExist {
				rdb.Set(ctx, cacheKey, "null", 60*time.Second) // 缓存用户不存在标记，防止缓存穿透
			}
			return nil, err
		}

		// 写缓存 设置ttl防雪崩
		data, err := json.Marshal(outputSql)
		if err != nil {
			return nil, fmt.Errorf("marshal output failed: %w", err)
		}
		ttl := 3600 + rand.Intn(600) // 1小时 ± 10分钟随机数，防止缓存雪崩
		_ = rdb.Set(ctx, cacheKey, data, time.Duration(ttl)*time.Second).Err()

		return outputSql, nil
	})

	// 可观测：打印是否复用结果，方便排查击穿
	_ = shared

	if err != nil {
		return nil, api.CodeInternalError, err
	}

	output, ok := result.(*model.UserInfoOutput)
	if !ok {
		return nil, api.CodeInternalError, errors.New("type assertion failed")
	}

	return output, api.CodeSuccess, nil
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

	// 5. 删除 Redis 缓存
	if rdb != nil {
		cacheKey := fmt.Sprintf("user:%d", userId)
		rdb.Del(ctx, cacheKey)
	}

	// 6. 返回更新后的用户信息
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
