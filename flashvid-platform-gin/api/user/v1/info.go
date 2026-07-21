package v1

import "time"

// 用户信息响应
type UserInfoResp struct {
	UserId 	 int64  `json:"userId"`       // 用户ID
	Username string `json:"username"`     // 用户名
	Nickname string `json:"nickname"`     // 昵称
	Avatar   string `json:"avatar"`       // 头像
	Bio     string `json:"bio"`          // 个人简介
	City     string `json:"city"`         // 城市
	FollowersConut int32    `json:"followersCount"`    // 粉丝数
	FollowingConut int32   `json:"followingCount"`    // 关注数
	VideosCount    int32    `json:"videosCount"`       // 发布视频数
	LikesCount     int32    `json:"likesCount"`        // 获赞数
	Phone    string `json:"phone"`        // 手机号
	Gender   int32  `json:"gender"`       // 性别：0-未知 1-男 2-女
	Birthday time.Time `json:"birthday"`     // 生日
	Email    string `json:"email"`        // 邮箱
	CreatedAt time.Time `json:"createdAt"`    // 创建时间
}
