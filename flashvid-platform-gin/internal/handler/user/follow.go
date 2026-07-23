package user

import (
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/middleware"
	"github.com/gin-gonic/gin"
	"flashvid-platform-gin/internal/service/user"
	"strconv"
	v1 "flashvid-platform-gin/api/user/v1"
)

// 关注用户接口
func FollowUserHandler(c *gin.Context) {
	// 1. 获取登录用户ID
	loginUserId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	loginUserIdInt64, ok := loginUserId.(int64)
	if !ok || loginUserIdInt64 <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	// 2. 获取要关注的用户ID
	followUserId , err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	if followUserId <= 0 {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 3. 关注用户不能是自己
	if loginUserIdInt64 == followUserId {
		api.ResponseError(c, api.CodeCannotFollowSelf)
		return
	}
	// 4. 调用service进行关注操作
	isFollow, resCode, err := user.FollowUser(c, loginUserIdInt64, followUserId)
	if err != nil {
		if resCode == api.CodeAlreadyFollowed {
			api.ResponseErrorWithMsg(c, resCode, "已经关注该用户")
			return
		}
		api.ResponseError(c, resCode)
		return
	}
	// 5. 返回响应
	api.ResponseSuccess(c, v1.FollowResp{
		IsFollowing: isFollow,
	})
}

// 取消关注用户接口
func UnfollowUserHandler(c *gin.Context) {
	// 1. 获取登录用户ID
	loginUserId, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	loginUserIdInt64, ok := loginUserId.(int64)
	if !ok || loginUserIdInt64 <= 0 {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	// 2. 获取要取消关注的用户ID
	followUserId , err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	if followUserId <= 0 {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 3. 取消关注用户不能是自己
	if loginUserIdInt64 == followUserId {
		api.ResponseError(c, api.CodeCannotFollowSelf)
		return
	}
	// 4. 调用service进行取消关注操作
	isFollow, resCode, err := user.UnfollowUser(c, loginUserIdInt64, followUserId)
	if err != nil {
		if resCode == api.CodeNotFollowed {
			api.ResponseErrorWithMsg(c, resCode, "未关注该用户")
			return
		}
		api.ResponseError(c, resCode)
		return
	}
	// 5. 返回响应
	api.ResponseSuccess(c, v1.FollowResp{
		IsFollowing: isFollow,
	})
}

// 查看用户的粉丝列表接口
func GetUserFollowersHandler(c *gin.Context) {
	// 1. 获取要查看的用户ID
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	if userId <= 0 {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2. 获取分页参数
	var req v1.GetUserFollowReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	// 3. 调用service获取用户粉丝列表
	output, resCode, err := user.GetUserFollowers(c, userId, req.Page, req.PageSize)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, v1.GetUserFollowersResp{
		Followers:  output.Followers,
		Pagination: output.Pagination,
	})
}

// 查看用户的关注列表接口
func GetUserFollowingHandler(c *gin.Context) {
		// 1. 获取要查看的用户ID
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	if userId <= 0 {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2. 获取分页参数
	var req v1.GetUserFollowReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	// 3. 调用service获取用户关注列表
	output, resCode, err := user.GetUserFollowing(c, userId, req.Page, req.PageSize)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4. 返回响应
	api.ResponseSuccess(c, v1.GetUserFollowingResp{
		Followings: output.Followings,
		Pagination: output.Pagination,
	})
}
