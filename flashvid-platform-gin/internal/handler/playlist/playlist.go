package playlist

import (
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/playlist/v1"
	"flashvid-platform-gin/internal/middleware"
	"flashvid-platform-gin/internal/service/playlist"
	"github.com/gin-gonic/gin"
)

// 获取用户播放列表接口
func GetUserPlaylistsHandler(c *gin.Context) {
	// 1.获取登录用户ID
	loginUserID, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userID, exists := loginUserID.(int64)
	if !exists {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2.调用service层获取用户播放列表
	output, resCode, err := playlist.GetUserPlaylists(c, userID)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 3.返回播放列表信息
	api.ResponseSuccess(c, &v1.GetPlayListsResp{
		PlayLists: output,
	})
}

// 创建播放列表接口
func CreatePlaylistHandler(c *gin.Context) {
	// 1.获取登录用户ID
	loginUserID, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userID, exists := loginUserID.(int64)
	if !exists {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2.获取创建播放列表的请求参数
	var req v1.CreatePlaylistReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3.调用service层创建播放列表
	output, resCode, err := playlist.CreatePlaylist(c, userID, req)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	// 4.返回创建的播放列表信息
	api.ResponseSuccess(c, &v1.CreatePlaylistResp{
		Playlist: output,
	})
}

// 更新播放列表接口
func UpdatePlaylistHandler(c *gin.Context) {
	// 1.获取登录用户ID
	loginUserID, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userID, exists := loginUserID.(int64)
	if !exists {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2.获取更新播放列表的请求参数
	var req v1.UpdatePlaylistReq
	// 先获取请求参数中的ID
	if err := c.ShouldBindUri(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 再获取请求体中的其他字段
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3.调用service层更新播放列表
	resCode, err := playlist.UpdatePlaylist(c, userID, req)
	if err != nil || resCode != api.CodeSuccess {
		api.ResponseError(c, resCode)
		return
	}
	// 4.返回更新后的播放列表信息
	api.ResponseSuccess(c, "更新播放列表成功")
}

// 获取播放列表内的视频接口（游标分页）
func GetPlaylistVideosHandler(c *gin.Context) {
	// 1. 获取路径参数 + 查询参数
	var req v1.GetPlaylistVideosReq
	if err := c.ShouldBindUri(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 2. 调用 service
	resp, resCode, err := playlist.GetPlaylistVideos(c, req.ID, req.Cursor, req.Count)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	if resCode != api.CodeSuccess {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, resp)
}

// 添加视频到播放列表接口
func AddVideoToPlaylistHandler(c *gin.Context) {
	// 1. 获取登录用户ID
	loginUserID, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userID, exists := loginUserID.(int64)
	if !exists {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2. 获取请求参数
	var req v1.AddVideoToPlaylistReq
	if err := c.ShouldBindUri(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3. 调用 service
	resCode, err := playlist.AddVideoToPlaylist(c, userID, req.PlaylistID, req.VideoID)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	if resCode != api.CodeSuccess {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, "添加成功")
}

// 从播放列表移除视频接口
func RemoveVideoFromPlaylistHandler(c *gin.Context) {
	// 1. 获取登录用户ID
	loginUserID, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userID, exists := loginUserID.(int64)
	if !exists {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2. 获取请求参数
	var req v1.RemoveVideoFromPlaylistReq
	if err := c.ShouldBindUri(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3. 调用 service
	resCode, err := playlist.RemoveVideoFromPlaylist(c, userID, req.PlaylistID, req.VideoID)
	if err != nil {
		api.ResponseError(c, resCode)
		return
	}
	if resCode != api.CodeSuccess {
		api.ResponseError(c, resCode)
		return
	}
	api.ResponseSuccess(c, "移除成功")
}

// 删除播放列表接口
func DeletePlaylistHandler(c *gin.Context) {
	// 1.获取登录用户ID
	loginUserID, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		api.ResponseError(c, api.CodeValueNotExist)
		return
	}
	userID, exists := loginUserID.(int64)
	if !exists {
		api.ResponseError(c, api.CodeInvalidUserID)
		return
	}
	// 2.获取删除播放列表的请求参数
	var req v1.DeletePlaylistReq
	if err := c.ShouldBindUri(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	// 3.调用service层删除播放列表
	resCode, err := playlist.DeletePlaylist(c, userID, req.ID)
	if err != nil || resCode != api.CodeSuccess {
		api.ResponseError(c, resCode)
		return
	}
	// 4.返回删除结果
	api.ResponseSuccess(c, "删除播放列表成功")
}
