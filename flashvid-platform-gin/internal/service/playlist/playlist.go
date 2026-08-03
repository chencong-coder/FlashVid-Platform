package playlist

import (
	"context"
	"errors"
	v1 "flashvid-platform-gin/api/playlist/v1"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"gorm.io/gorm"
	"time"
)

// playlistToInfo 将 Playlist 转换为 PlayListInfo
func playlistToInfo(p *model.Playlist) model.PlayListInfo {
	return model.PlayListInfo{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		CoverURL:    p.CoverURL,
		VideoCount:  p.VideoCount,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// 获取用户播放列表
func GetUserPlaylists(ctx context.Context, userID int64) ([]model.PlayListInfo, api.ResCode, error) {
	// 1.根据用户id查playlists表，获取用户的播放列表信息
	Playlists, err := query.Playlist.WithContext(ctx).
		Where(query.Playlist.UserID.Eq(userID)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 2.将查询结果转换为PlayListInfo结构体切片
	var output []model.PlayListInfo
	for _, playlist := range Playlists {
		output = append(output, model.PlayListInfo{
			ID:          playlist.ID,
			Title:       playlist.Title,
			Description: playlist.Description,
			CoverURL:    playlist.CoverURL,
			VideoCount:  playlist.VideoCount,
			CreatedAt:   playlist.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 3.返回播放列表信息
	return output, api.CodeSuccess, nil
}

// 创建播放列表
func CreatePlaylist(ctx context.Context, userID int64, req v1.CreatePlaylistReq) (model.PlayListInfo, api.ResCode, error) {
	// 1.创建播放列表对象
	playlist := &model.Playlist{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		VideoCount:  0,
	}
	// 2.将播放列表对象插入数据库
	if err := query.Playlist.WithContext(ctx).Create(playlist); err != nil {
		return model.PlayListInfo{}, api.CodeCreatePlaylistFailed, err
	}
	// 3.返回创建的播放列表信息
	return model.PlayListInfo{
		ID:          playlist.ID,
		Title:       playlist.Title,
		Description: playlist.Description,
		CoverURL:    playlist.CoverURL,
		VideoCount:  playlist.VideoCount,
		CreatedAt:   playlist.CreatedAt.Format("2006-01-02 15:04:05"),
	}, api.CodeSuccess, nil
}

// 更新播放列表信息
func UpdatePlaylist(ctx context.Context, userID int64, req v1.UpdatePlaylistReq) (api.ResCode, error) {
	playlistID := req.ID
	// 1.根据播放列表ID查找播放列表是否存在
	pl, err := query.Playlist.WithContext(ctx).
		Where(query.Playlist.ID.Eq(playlistID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.CodePlaylistNotExist, nil
		}
		return  api.CodeInternalError, err
	}
	// 校验归属权
	if pl.UserID != userID {
		return api.CodeNotPlaylistOwner, nil
	}
	// 2.更新播放列表信息 如果存在
	var updateData = make(map[string]any)
	if req.Title != nil {
		updateData["title"] = *req.Title
	}
	if req.Description != nil {
		updateData["description"] = *req.Description
	}
	if req.CoverURL != nil {
		updateData["cover_url"] = *req.CoverURL
	}
	// 3.执行更新操作
	if len(updateData) > 0 {
		_, err = query.Playlist.WithContext(ctx).
			Where(query.Playlist.ID.Eq(playlistID)).
			Updates(updateData)
		if err != nil {
			return api.CodeInternalError, err
		}
	}
	return api.CodeSuccess, nil
}

// 删除播放列表
func DeletePlaylist(ctx context.Context, userID int64, playlistID int64) (api.ResCode, error) {
	// 1.根据播放列表ID查找播放列表是否存在
	playlist, err := query.Playlist.WithContext(ctx).
		Where(query.Playlist.ID.Eq(playlistID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.CodePlaylistNotExist, err
		}
		return api.CodeInternalError, err
	}
	// 校验归属权
	if playlist.UserID != userID {
		return api.CodeNotPlaylistOwner, nil
	}
	// 2.开启事务：软删除播放列表 + 硬删除关联视频
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 软删除播放列表（设置 deleted_at）
		if _, err = tx.Playlist.WithContext(ctx).Delete(playlist); err != nil {
			return err
		}
		// 硬删除 playlist_videos 关联记录（playlist_videos 无 deleted_at）
		if _, err = tx.PlaylistVideo.WithContext(ctx).
			Where(tx.PlaylistVideo.PlaylistID.Eq(playlistID)).
			Delete(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return api.CodeInternalError, err
	}
	return api.CodeSuccess, nil
}

// GetPlaylistVideos 游标分页获取播放列表内的视频（按加入时间倒序）
func GetPlaylistVideos(ctx context.Context, playlistID int64, cursor string, limit int) (v1.GetPlaylistVideosResp, api.ResCode, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	// 1. 查播放列表信息
	pl, err := query.Playlist.WithContext(ctx).
		Where(query.Playlist.ID.Eq(playlistID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return v1.GetPlaylistVideosResp{}, api.CodePlaylistNotExist, nil
		}
		return v1.GetPlaylistVideosResp{}, api.CodeInternalError, err
	}
	// 2. 查 playlist_videos（游标分页，按加入时间倒序）
	pv := query.PlaylistVideo
	pvQ := pv.WithContext(ctx).Where(pv.PlaylistID.Eq(playlistID))
	if cursor != "" {
		cursorTime, err := time.Parse("2006-01-02 15:04:05", cursor)
		if err != nil {
			return v1.GetPlaylistVideosResp{}, api.CodeInvalidParam, err
		}
		pvQ = pvQ.Where(pv.CreatedAt.Lt(cursorTime))
	}
	pvRows, err := pvQ.Order(pv.CreatedAt.Desc()).Limit(limit + 1).Find()
	if err != nil {
		return v1.GetPlaylistVideosResp{}, api.CodeInternalError, err
	}
	// 3. 判断是否有更多
	hasMore := len(pvRows) > limit
	if hasMore {
		pvRows = pvRows[:limit]
	}
	nextCursor := ""
	if hasMore && len(pvRows) > 0 {
		nextCursor = pvRows[len(pvRows)-1].CreatedAt.Format("2006-01-02 15:04:05")
	}
	if len(pvRows) == 0 {
		return v1.GetPlaylistVideosResp{
			Playlist: playlistToInfo(pl),
			Videos:   []model.VideoInfo{},
			HasMore:  false,
		}, api.CodeSuccess, nil
	}
	// 4. 批量查视频 + 作者 + 话题（同 feed service 模式）
	videoIDs := make([]int64, 0, len(pvRows))
	for _, row := range pvRows {
		videoIDs = append(videoIDs, row.VideoID)
	}
	videos, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.In(videoIDs...)).
		Find()
	if err != nil {
		return v1.GetPlaylistVideosResp{}, api.CodeInternalError, err
	}
	videoMap := make(map[int64]*model.Video, len(videos))
	authorIDs := make([]int64, 0, len(videos))
	for _, v := range videos {
		videoMap[v.ID] = v
		authorIDs = append(authorIDs, v.UserID)
	}
	authors, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(authorIDs...)).
		Find()
	if err != nil {
		return v1.GetPlaylistVideosResp{}, api.CodeInternalError, err
	}
	authorMap := make(map[int64]*model.User, len(authors))
	for _, a := range authors {
		authorMap[a.ID] = a
	}
	videoTopics, err := query.VideoTopic.WithContext(ctx).
		Where(query.VideoTopic.VideoID.In(videoIDs...)).
		Find()
	if err != nil {
		return v1.GetPlaylistVideosResp{}, api.CodeInternalError, err
	}
	topicIDs := make([]int64, 0)
	videoTopicMap := make(map[int64][]int64)
	for _, vt := range videoTopics {
		videoTopicMap[vt.VideoID] = append(videoTopicMap[vt.VideoID], vt.TopicID)
		topicIDs = append(topicIDs, vt.TopicID)
	}
	topicMap := make(map[int64]string)
	if len(topicIDs) > 0 {
		topics, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.In(topicIDs...)).
			Find()
		if err != nil {
			return v1.GetPlaylistVideosResp{}, api.CodeInternalError, err
		}
		for _, t := range topics {
			topicMap[t.ID] = t.Name
		}
	}
	// 5. 按 playlist_videos 顺序构建输出（保持倒序）
	outputVideos := make([]model.VideoInfo, 0, len(pvRows))
	for _, row := range pvRows {
		v, ok := videoMap[row.VideoID]
		if !ok {
			continue // 视频已软删除，跳过
		}
		author := authorMap[v.UserID]
		topicNames := make([]string, 0)
		for _, tid := range videoTopicMap[v.ID] {
			if name, ok := topicMap[tid]; ok {
				topicNames = append(topicNames, name)
			}
		}
		videoAuthor := model.VideoAuthor{}
		if author != nil {
			videoAuthor = model.VideoAuthor{
				ID:       author.ID,
				Username: author.Username,
				Nickname: author.Nickname,
				Avatar:   author.Avatar,
			}
		}
		outputVideos = append(outputVideos, model.VideoInfo{
			ID:          v.ID,
			Title:       v.Title,
			Description: v.Description,
			CoverUrl:    v.CoverURL,
			VideoUrl:    v.VideoURL,
			Duration:    v.Duration,
			Width:       v.Width,
			Height:      v.Height,
			MusicId:     v.MusicID,
			City:        v.City,
			Topics:      topicNames,
			Author:      videoAuthor,
			Stats: model.VideoStats{
				ViewCount:     v.ViewCount,
				LikeCount:     v.LikeCount,
				CommentCount:  v.CommentCount,
				ShareCount:    v.ShareCount,
				FavoriteCount: v.FavoriteCount,
			},
			PublishedAt: v.PublishedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return v1.GetPlaylistVideosResp{
		Playlist:   playlistToInfo(pl),
		Videos:     outputVideos,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, api.CodeSuccess, nil
}

// AddVideoToPlaylist 添加视频到播放列表
func AddVideoToPlaylist(ctx context.Context, userID, playlistID, videoID int64) (api.ResCode, error) {
	// 1. 校验播放列表归属
	pl, err := query.Playlist.WithContext(ctx).
		Where(query.Playlist.ID.Eq(playlistID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.CodePlaylistNotExist, nil
		}
		return api.CodeInternalError, err
	}
	if pl.UserID != userID {
		return api.CodeNotPlaylistOwner, nil
	}
	// 2. 校验视频存在
	_, err = query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.CodeVideoNotExist, nil
		}
		return api.CodeInternalError, err
	}
	// 3. 幂等校验：视频是否已在列表中
	_, err = query.PlaylistVideo.WithContext(ctx).
		Where(query.PlaylistVideo.PlaylistID.Eq(playlistID), query.PlaylistVideo.VideoID.Eq(videoID)).
		First()
	if err == nil {
		return api.CodeVideoAlreadyInList, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return api.CodeInternalError, err
	}
	// 4. 事务：插入关联记录 + 累加 video_count
	err = query.Q.Transaction(func(tx *query.Query) error {
		if err = tx.PlaylistVideo.WithContext(ctx).Create(&model.PlaylistVideo{
			PlaylistID: playlistID,
			VideoID:    videoID,
		}); err != nil {
			return err
		}
		_, err = tx.Playlist.WithContext(ctx).
			Where(tx.Playlist.ID.Eq(playlistID)).
			UpdateSimple(tx.Playlist.VideoCount.Add(1))
		return err
	})
	if err != nil {
		return api.CodeInternalError, err
	}
	return api.CodeSuccess, nil
}

// RemoveVideoFromPlaylist 从播放列表移除视频
func RemoveVideoFromPlaylist(ctx context.Context, userID, playlistID, videoID int64) (api.ResCode, error) {
	// 1. 校验播放列表归属
	pl, err := query.Playlist.WithContext(ctx).
		Where(query.Playlist.ID.Eq(playlistID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.CodePlaylistNotExist, nil
		}
		return api.CodeInternalError, err
	}
	if pl.UserID != userID {
		return api.CodeNotPlaylistOwner, nil
	}
	// 2. 校验视频在播放列表中
	_, err = query.PlaylistVideo.WithContext(ctx).
		Where(query.PlaylistVideo.PlaylistID.Eq(playlistID), query.PlaylistVideo.VideoID.Eq(videoID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.CodeVideoNotInList, nil
		}
		return api.CodeInternalError, err
	}
	// 3. 事务：硬删除关联记录 + 递减 video_count
	err = query.Q.Transaction(func(tx *query.Query) error {
		if _, err = tx.PlaylistVideo.WithContext(ctx).
			Where(tx.PlaylistVideo.PlaylistID.Eq(playlistID), tx.PlaylistVideo.VideoID.Eq(videoID)).
			Delete(); err != nil {
			return err
		}
		_, err = tx.Playlist.WithContext(ctx).
			Where(tx.Playlist.ID.Eq(playlistID)).
			UpdateSimple(tx.Playlist.VideoCount.Sub(1))
		return err
	})
	if err != nil {
		return api.CodeInternalError, err
	}
	return api.CodeSuccess, nil
}