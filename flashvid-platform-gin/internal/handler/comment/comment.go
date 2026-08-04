package comment

import (
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/comment/v1"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/internal/service/comment"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 获取评论列表接口（公开，userId 可选，登录后可返回 isLiked 状态）
func GetCommentsHandler(c *gin.Context) {
	// 获取登录用户ID（可选，未登录时为0）
	var userId int64
	if loginUserId, exists := c.Get(middleware.CtxKeyUserID); exists {
		if id, ok := loginUserId.(int64); ok {
			userId = id
		}
	}
	// 获取视频ID
	videoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || videoId <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 绑定分页参数
	var req v1.GetCommentsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Count < 10 {
		req.Count = 10
	}
	// 调用 service
	output, resCode, err := comment.GetComments(c, userId, videoId, req.Count, req.Cursor)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, v1.GetCommentsResp{
		Comments:        output.Comments,
		NextCursorToken: output.NextCursorToken,
		HasMore:         output.HasMore,
	})
}

// 发表评论接口（需要登录）
func CreateCommentHandler(c *gin.Context) {
	loginUserId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}
	userId, ok := loginUserId.(int64)
	if !ok || userId <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	videoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || videoId <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	var req v1.CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	parentID, err := parseOptionalID(req.ParentID)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	replyToUserID, err := parseOptionalID(req.ReplyToUserID)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	commentInfo, replyInfo, resCode, err := comment.CreateComment(c, userId, videoId, req.Content, parentID, replyToUserID)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, v1.CreateCommentResp{
		Comment: commentInfo,
		Reply:   replyInfo,
	})
}

func parseOptionalID(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 0 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}

// 删除评论接口（需要登录）
func DeleteCommentHandler(c *gin.Context) {
	loginUserId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}
	userId, ok := loginUserId.(int64)
	if !ok || userId <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	commentId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || commentId <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	resCode, err := comment.DeleteComment(c, userId, commentId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, gin.H{})
}

// 获取评论回复列表接口（公开，userId 可选）
func GetRepliesHandler(c *gin.Context) {
	// 获取登录用户ID（可选）
	var userId int64
	if loginUserId, exists := c.Get(middleware.CtxKeyUserID); exists {
		if id, ok := loginUserId.(int64); ok {
			userId = id
		}
	}
	// 获取评论ID
	commentId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || commentId <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 调用 service
	replies, resCode, err := comment.GetReplies(c, userId, commentId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, v1.GetRepliesResp{Replies: replies})
}

// 点赞评论接口（需要登录）
func LikeCommentHandler(c *gin.Context) {
	loginUserId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}
	userId, ok := loginUserId.(int64)
	if !ok || userId <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	commentId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || commentId <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	resp, resCode, err := comment.LikeComment(c, userId, commentId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, resp)
}

// 取消点赞评论接口（需要登录）
func UnlikeCommentHandler(c *gin.Context) {
	loginUserId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}
	userId, ok := loginUserId.(int64)
	if !ok || userId <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	commentId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || commentId <= 0 {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	resp, resCode, err := comment.UnlikeComment(c, userId, commentId)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, resp)
}
