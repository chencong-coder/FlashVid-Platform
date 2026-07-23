package model

// 用户粉丝列表输出
type UserFollowersOutput struct {
	Followers  []UserInfo `json:"followers"`
	Pagination Pagination `json:"pagination"`
}

// 用户关注列表输出
type UserFollowingOutput struct {
	Followings []UserInfo `json:"following"`
	Pagination Pagination `json:"pagination"`
}

// 用户信息
type UserInfo struct {
	UserId         int64     `json:"userId"`
	Username       string    `json:"username"`
	Gender         int32     `json:"gender"`
	Nickname       string    `json:"nickname"`
	Avatar         string    `json:"avatar"`
	Bio            string    `json:"bio"`
	City           string    `json:"city"`
	FollowersCount int32     `json:"followersCount"`
	FollowingCount int32     `json:"followingCount"`
	VideosCount    int32     `json:"videosCount"`
	LikesCount     int32     `json:"likesCount"`
	Phone          string    `json:"phone"`
	Birthday       string `json:"birthday"`
	Email          string    `json:"email"`
	CreatedAt      string `json:"createdAt"`
}

