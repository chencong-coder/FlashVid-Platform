package v1

// RegisterReq 用户注册请求
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=4,max=32"`  // 用户名（4-32字符）
	Password string `json:"password" binding:"required,min=6,max=20"`  // 密码（6-20字符）
	Phone    string `json:"phone" binding:"required,len=11,numeric"`   // 手机号（11位纯数字）
	Code     string `json:"code" binding:"required,len=6"`             // 验证码（6位）
	Birthday string `json:"birthday" binding:"omitempty,datetime=2006-01-02"` // 生日，可选（格式：YYYY-MM-DD）
	Email    string `json:"email" binding:"omitempty,email"`           // 邮箱，可选
}

// RegisterResp 用户注册响应
type RegisterResp struct {
	UserID   int64  `json:"user_id"`  // 用户ID
	Username string `json:"username"` // 用户名
}

// LoginReq 用户登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// LoginResp 用户登录响应
type LoginResp struct {
	UserID   int64  `json:"user_id"`  // 用户ID
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像
	AccessToken    string `json:"accessToken"`    // JWT Token
	RefreshToken   string `json:"refreshToken"`   // 刷新Token
}

