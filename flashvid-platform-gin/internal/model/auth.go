package model

// 注册输出
type RegisterOutput struct {
	UserID   int64  `json:"userId"`  // 用户ID
	Username string `json:"username"` // 用户名
}

// 登录输出
type LoginOutput struct {
	UserID       int64  `json:"userId"`       // 用户ID
	Username     string `json:"username"`      // 用户名
	Nickname     string `json:"nickname"`      // 昵称
	Avatar       string `json:"avatar"`        // 头像
	AccessToken  string `json:"accessToken"`   // JWT Token
	RefreshToken string `json:"refreshToken"`  // 刷新Token
}

// 刷新Token输出
type RefreshOutput struct {
	AccessToken  string `json:"accessToken"`  // 新的JWT Token
	RefreshToken string `json:"refreshToken"` // 新的刷新Token
}