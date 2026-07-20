package service

import (
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/api/auth/v1"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"context"
	"golang.org/x/crypto/bcrypt"
	"flashvid-platform-gin/pkg/snowflake"
	"go.uber.org/zap"
)

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
	// 6. 创建用户
	user := &model.User{
		ID:       userID,
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Username, // 默认昵称等于用户名
		Phone:    req.Phone,
		Status:    1,  // 1-正常
		IPAddress: ip, // 注册IP
	}
	// 只插入指定字段，避免 birthday/last_login_at 零值问题
	err = query.User.WithContext(ctx).Select(
		query.User.ID,
		query.User.Username,
		query.User.Password,
		query.User.Nickname,
		query.User.Phone,
		query.User.Status,
		query.User.IPAddress,
	).Create(user)
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