package model

// 评论信息
type CommentInfo struct {
	ID		int64  `json:"id,string"`         // 评论ID
	Content string `json:"content"`    // 评论内容
	User CommentUser `json:"user"`       // 评论作者信息
	LikeCount int32  `json:"likeCount"`  // 点赞数
	ReplyCount int32  `json:"replyCount"` // 回复数
	IsLiked   bool   `json:"isLiked"`    // 当前登录用户是否已点赞
	IsAuthored bool   `json:"isAuthored"` // 当前登录用户是否为该视频作者
	Replies []ReplyInfo `json:"replies"`    // 回复列表
	CreatedAt string `json:"createdAt"` // "2006-01-02 15:04:05"
}

// 评论用户信息
type CommentUser struct {
	ID       int64  `json:"id,string"`       // 用户ID
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像
}

// 回复信息
type ReplyInfo struct {
	ID		int64  `json:"id,string"`        // 回复ID(评论)
	Content string `json:"content"`   // 回复内容
	User CommentUser `json:"user"`      // 回复作者信息
	ReplyTo ReplyToUser `json:"replyTo"`  // 被回复的用户信息 
	LikeCount int32  `json:"likeCount"` // 点赞数
	IsLiked   bool   `json:"isLiked"`   // 当前登录用户是否已点赞
	CreatedAt string `json:"createdAt"` // "2006-01-02 15:04:05"
}

// 评论回复用户信息
type ReplyToUser struct {
	ID       int64  `json:"id,string"`       // 被回复的用户ID
	Nickname string `json:"nickname"` // 昵称
}

// 评论列表输出
type CommentListOutput struct {
	Comments 	[]CommentInfo `json:"comments"` // 评论列表
	NextCursorToken string      `json:"nextCursorToken"` // 下次请求传这个，空字符串表示没有更多
	HasMore         bool        `json:"hasMore"` // 是否还有更多数据
}