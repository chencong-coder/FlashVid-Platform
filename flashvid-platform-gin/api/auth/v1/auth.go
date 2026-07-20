package v1

// RegisterReq 用户注册请求
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=4,max=32"`  // 用户名（4-32字符）
	Password string `json:"password" binding:"required,min=6,max=20"`  // 密码（6-20字符）
	Phone    string `json:"phone" binding:"required,len=11"`           // 手机号（11位）
	Code     string `json:"code" binding:"required,len=6"`             // 验证码（6位）
}

// RegisterResp 用户注册响应
type RegisterResp struct {
	UserID   int64  `json:"user_id"`  // 用户ID
	Username string `json:"username"` // 用户名
	Token    string `json:"token"`    // JWT Token
}

