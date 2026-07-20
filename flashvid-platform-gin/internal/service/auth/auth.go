package service

import (
	"flashvid-platform-gin/api/auth/v1"
	"flashvid-platform-gin/internal/model"
)

func Register(req *v1.RegisterReq) (*model.RegisterOutput, error) {
	// 1. 校验验证码
	
	// 2. 校验用户名是否已存在
	// 3. 创建用户
	// 4. 生成JWT Token
	// 5. 返回注册结果
}