package v1

import (
	"flashvid-platform-gin/internal/model"
)

// 获取评论请求
type GetCommentsReq struct {
	Cursor   string `form:"cursor" binding:"omitempty"` // 游标，首次不传
    Count    int    `form:"count"    binding:"omitempty,max=50"` // 请求数量，最大50，小于10时handler层默认为10
}

// 获取评论响应
type GetCommentsResp struct {
	Comments        []model.CommentInfo `json:"comments"`
	NextCursorToken string              `json:"nextCursorToken"`
	HasMore         bool                `json:"hasMore"`
}

// 获取回复列表响应
type GetRepliesResp struct {
	Replies []model.ReplyInfo `json:"replies"`
}

// 创建评论请求
type CreateCommentReq struct {
	Content       string `json:"content" binding:"required,min=1,max=500"`        // 评论内容
	ParentID      int64  `json:"parentId" binding:"omitempty,min=0"`              // 父评论ID，0表示一级评论
	ReplyToUserID int64  `json:"replyToUserId" binding:"omitempty,min=0"`         // 被回复的用户ID
}

// 创建评论响应（根据parentId返回不同类型）
type CreateCommentResp struct {
	Comment *model.CommentInfo `json:"comment,omitempty"` // 一级评论
	Reply   *model.ReplyInfo   `json:"reply,omitempty"`   // 回复
}

// 点赞评论响应
type LikeCommentResp struct {
	IsLiked   bool  `json:"isLiked"`
	LikeCount int32 `json:"likeCount"`
}
