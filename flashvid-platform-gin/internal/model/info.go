package model

import "time"

// 用户信息输出
type UserInfoOutput struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
	Gender   int32  `json:"gender"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
	City     string `json:"city"`
	FollowersCount int32    `json:"followersCount"`
	FollowingCount int32   `json:"followingCount"`
	VideosCount    int32    `json:"videosCount"`
	LikesCount     int32    `json:"likesCount"`
	Phone    string `json:"phone"`
	Birthday time.Time `json:"birthday"`
	Email    string `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}