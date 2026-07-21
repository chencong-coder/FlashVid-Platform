package service

import (
	"context"
	"time"

	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/auth/v1"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"flashvid-platform-gin/pkg/snowflake"

	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"
	"gorm.io/gen/field"
)

const userAvatarDefault = "https://img1.baidu.com/it/u=470345945,3074368414&fm=253&app=138&f=JPEG?w=800&h=1319"

func Register(ctx context.Context, req *v1.RegisterReq, ip string) (*model.RegisterOutput, api.ResCode, error) {
	// 1. 校验验证码
	if req.Code != "123456" {
		return nil, api.CodeInvalidCode, nil
	}
	// 2. 校验用户名是否已存在
	count, err := query.User.WithContext(ctx).
		Where(query.User.Username.Eq(req.Username)).
		Count()
	if err != nil {
		zap.L().Error("failed to check username existence", zap.Error(err))
		return nil, api.CodeInternalError, err
	}
	if count > 0 {
		return nil, api.CodeUserExist, nil
	}
	// 3. 校验手机号是否已存在
	count, err = query.User.WithContext(ctx).
		Where(query.User.Phone.Eq(req.Phone)).
		Count()
	if err != nil {
		zap.L().Error("failed to check phone existence", zap.Error(err))
		return nil, api.CodeInternalError, err
	}
	if count > 0 {
		return nil, api.CodePhoneExist, nil
	}
	// 4. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("failed to hash password", zap.Error(err))
		return nil, api.CodeInternalError, err
	}
	// 5. 雪花算法生成唯一id
	userID, err := snowflake.NextID()
	if err != nil {
		zap.L().Error("failed to generate user ID", zap.Error(err))
		return nil, api.CodeInternalError, err
	}

	// 6. 解析生日（可选）
	var birthday time.Time
	if req.Birthday != "" {
		birthday, err = time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			zap.L().Error("failed to parse birthday", zap.Error(err))
			return nil, api.CodeInvalidBirthday, nil
		}
	}

	// 7. 创建用户
	user := &model.User{
		ID:        userID,
		Username:  req.Username,
		Password:  string(hashedPassword),
		Nickname:  req.Username, // 默认昵称等于用户名
		Phone:     req.Phone,
		Status:    1,              // 1-正常
		IPAddress: ip,             // 注册IP
		Email:     req.Email,      // 邮箱（可选）
		Avatar:    userAvatarDefault,
	}

	// 构建要插入的字段列表
	selectFields := []field.Expr{
		query.User.ID,
		query.User.Username,
		query.User.Password,
		query.User.Nickname,
		query.User.Phone,
		query.User.Status,
		query.User.IPAddress,
		query.User.Avatar,
	}

	// 如果提供了邮箱，加入插入列表
	if req.Email != "" {
		selectFields = append(selectFields, query.User.Email)
	}

	// 如果提供了生日，设置并加入插入列表
	if req.Birthday != "" {
		user.Birthday = birthday
		selectFields = append(selectFields, query.User.Birthday)
	}

	// 只插入指定字段，避免 last_login_at 零值问题
	err = query.User.WithContext(ctx).Select(selectFields...).Create(user)
	if err != nil {
		zap.L().Error("failed to create user", zap.Error(err))
		return nil, api.CodeInternalError, err
	}
	// 7. 返回注册结果
	return &model.RegisterOutput{
		UserID:   userID,
		Username: user.Username,
	}, api.CodeSuccess, nil
}