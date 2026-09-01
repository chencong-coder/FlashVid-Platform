package notification

import (
	"github.com/gin-gonic/gin"

	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/notification/v1"
	notifSvc "flashvid-platform-gin/internal/service/notification"
)

// GetNotificationsHandler 获取通知列表
func GetNotificationsHandler(c *gin.Context) {
	var req v1.GetNotificationsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	userId := c.GetInt64("user_id")
	list, total, err := notifSvc.GetNotifications(c, userId, req.ActionTypes, req.Page, req.PageSize)
	if err != nil {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	api.ResponseSuccess(c, v1.GetNotificationsResp{List: list, Total: total})
}

// GetUnreadCountsHandler 获取各类型未读数
func GetUnreadCountsHandler(c *gin.Context) {
	userId := c.GetInt64("user_id")
	counts, err := notifSvc.GetUnreadCountsWithCache(c, userId)
	if err != nil {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	api.ResponseSuccess(c, counts)
}

// MarkAsReadHandler 标记已读
func MarkAsReadHandler(c *gin.Context) {
	var req v1.MarkAsReadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	userId := c.GetInt64("user_id")
	if err := notifSvc.MarkAsRead(c, userId, req.ActionTypes); err != nil {
		api.ResponseError(c, api.CodeInternalError)
		return
	}
	api.ResponseSuccess[any](c, nil)
}
