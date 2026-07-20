package model

// 注册输出
type RegisterOutput struct {
	UserID   int64  `json:"user_id"`  // 用户ID
	Username string `json:"username"` // 用户名
	Token    string `json:"token"`    // JWT Token
}