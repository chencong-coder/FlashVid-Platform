package user
import (
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/model"
	"context"
	"errors"
	"flashvid-platform-gin/internal/dao/query"
	"gorm.io/gorm"
)

func GetUserLikes(ctx context.Context, userId int64, page, pageSize int) (*model.VideoListOutput, api.ResCode, error) {
	// 1. 查询用户是否存在
	_, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeUserNotExist, nil
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 查询视频总数
	likeCount, err := query.Like.WithContext(ctx).
		Where(query.Like.UserID.Eq(userId), query.Like.TargetType.Eq(1)).
		Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 3. 查询用户点赞的视频列表 分页 按创建顺序降序 类型为视频
	likes, err := query.Like.WithContext(ctx).
		Where(query.Like.UserID.Eq(userId), query.Like.TargetType.Eq(1)).
		Order(query.Like.CreatedAt.Desc()).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 4. 封装点赞的视频列表
	var likesList []model.VideoInfo
	// 如果点赞列表为空 直接返回
	if len(likes) == 0 {
		return &model.VideoListOutput{
			Videos:     []model.VideoInfo{},
			Pagination: model.Pagination{Page: page, PageSize: pageSize, Total: 0, TotalPages: 0},
		}, api.CodeSuccess, nil
	} else {
		for _, like := range likes {
			// 用target_id查询视频信息 
			video, err := query.Video.WithContext(ctx).
				Where(query.Video.ID.Eq(like.TargetID)).
				First()
			if err != nil {
				return nil, api.CodeInternalError, err
			}
			// 再根据video_id查询视频作者信息
			author, err := query.User.WithContext(ctx).
				Where(query.User.ID.Eq(video.UserID)).
				First()
			if err != nil {
				return nil, api.CodeInternalError, err
			}
			// 再获取话题信息
			var topidIDs []int64
			var topicNames []string
			videoTopics, err := query.VideoTopic.WithContext(ctx).
				Where(query.VideoTopic.VideoID.Eq(video.ID)).
				Find()
			if err != nil {
				return nil, api.CodeInternalError, err
			}
			if len(videoTopics) == 0 {
				topicNames = []string{}
			} else {
				for _, videoTopic := range videoTopics {
					topidIDs = append(topidIDs, videoTopic.TopicID)
				}
				topics, err := query.Topic.WithContext(ctx).
					Where(query.Topic.ID.In(topidIDs...)).
					Find()
				if err != nil {
					return nil, api.CodeInternalError, err
				}
				for _, topic := range topics {
					topicNames = append(topicNames, topic.Name)
				}
			}
			// 封装视频信息
			videoInfo := model.VideoInfo{
				ID:          video.ID,
				Title:       video.Title,
				Description: video.Description,
				CoverUrl:    video.CoverURL,
				VideoUrl:    video.VideoURL,
				Duration:    video.Duration,
				Width: 	 	video.Width,
				Height:      video.Height,
				MusicId:     video.MusicID,
				City:        video.City,
				Topics:      topicNames,
				Author:      model.VideoAuthor{
					ID:       author.ID,
					Username: author.Username,
					Avatar:   author.Avatar,
				},
				Stats: model.VideoStats{
					ViewCount:     video.ViewCount,
					LikeCount:     video.LikeCount,
					CommentCount:  video.CommentCount,
					ShareCount:    video.ShareCount,
					FavoriteCount: video.FavoriteCount,
				},
				PublishedAt: video.PublishedAt.Format("2006-01-02 15:04:05"),
			}
			likesList = append(likesList, videoInfo)
		}
	}
	// 5. 封装分页信息
	totalPages := (likeCount + int64(pageSize) - 1) / int64(pageSize)
	pagination := model.Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      likeCount,
		TotalPages: int(totalPages),
	}
	// 6. 封装返回结果
	return &model.VideoListOutput{
		Videos:     likesList,
		Pagination: pagination,
	}, api.CodeSuccess, nil
}

// GetUserLikes1 批量查询优化版本
func GetUserLikes1(ctx context.Context, userId int64, page, pageSize int) (*model.VideoListOutput, api.ResCode, error) {
	// 1. 查询用户是否存在
	_, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeUserNotExist, nil
		}
		return nil, api.CodeInternalError, err
	}

	// 2. 查询点赞总数
	likeCount, err := query.Like.WithContext(ctx).
		Where(query.Like.UserID.Eq(userId), query.Like.TargetType.Eq(1)).
		Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	// 3. 查询用户点赞的视频列表（分页）
	likes, err := query.Like.WithContext(ctx).
		Where(query.Like.UserID.Eq(userId), query.Like.TargetType.Eq(1)).
		Order(query.Like.CreatedAt.Desc()).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	// 如果点赞列表为空，直接返回
	if len(likes) == 0 {
		return &model.VideoListOutput{
			Videos:     []model.VideoInfo{},
			Pagination: model.Pagination{Page: page, PageSize: pageSize, Total: 0, TotalPages: 0},
		}, api.CodeSuccess, nil
	}

	// 4. 提取所有视频ID
	videoIds := make([]int64, 0, len(likes))
	for _, like := range likes {
		videoIds = append(videoIds, like.TargetID)
	}

	// 5. 批量查询所有视频信息
	videos, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.In(videoIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	// 构建 videoId -> video 的映射
	videoMap := make(map[int64]*model.Video)
	authorIds := make([]int64, 0, len(videos))
	for i := range videos {
		videoMap[videos[i].ID] = videos[i]
		authorIds = append(authorIds, videos[i].UserID)
	}

	// 6. 批量查询所有作者信息
	authors, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(authorIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	// 构建 authorId -> author 的映射
	authorMap := make(map[int64]*model.User)
	for i := range authors {
		authorMap[authors[i].ID] = authors[i]
	}

	// 7. 批量查询所有视频的话题关联
	videoTopics, err := query.VideoTopic.WithContext(ctx).
		Where(query.VideoTopic.VideoID.In(videoIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	// 提取所有话题ID
	topicIds := make([]int64, 0)
	videoTopicMap := make(map[int64][]int64) // videoId -> []topicId
	for _, vt := range videoTopics {
		topicIds = append(topicIds, vt.TopicID)
		videoTopicMap[vt.VideoID] = append(videoTopicMap[vt.VideoID], vt.TopicID)
	}

	// 8. 批量查询所有话题信息
	var topicMap map[int64]string
	if len(topicIds) > 0 {
		topics, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.In(topicIds...)).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}

		// 构建 topicId -> topicName 的映射
		topicMap = make(map[int64]string)
		for _, topic := range topics {
			topicMap[topic.ID] = topic.Name
		}
	}

	// 9. 按点赞顺序组装视频信息
	likesList := make([]model.VideoInfo, 0, len(likes))
	for _, like := range likes {
		video, ok := videoMap[like.TargetID]
		if !ok {
			continue // 视频可能被删除
		}

		author, ok := authorMap[video.UserID]
		if !ok {
			continue // 作者可能被删除
		}

		// 获取话题名称列表
		topicNames := make([]string, 0)
		if topicIdList, exists := videoTopicMap[video.ID]; exists {
			for _, topicId := range topicIdList {
				if topicName, ok := topicMap[topicId]; ok {
					topicNames = append(topicNames, topicName)
				}
			}
		}

		// 封装视频信息
		videoInfo := model.VideoInfo{
			ID:          video.ID,
			Title:       video.Title,
			Description: video.Description,
			CoverUrl:    video.CoverURL,
			VideoUrl:    video.VideoURL,
			Duration:    video.Duration,
			Width:       video.Width,
			Height:      video.Height,
			MusicId:     video.MusicID,
			City:        video.City,
			Topics:      topicNames,
			Author: model.VideoAuthor{
				ID:       author.ID,
				Username: author.Username,
				Avatar:   author.Avatar,
			},
			Stats: model.VideoStats{
				ViewCount:     video.ViewCount,
				LikeCount:     video.LikeCount,
				CommentCount:  video.CommentCount,
				ShareCount:    video.ShareCount,
				FavoriteCount: video.FavoriteCount,
			},
			PublishedAt: video.PublishedAt.Format("2006-01-02 15:04:05"),
		}
		likesList = append(likesList, videoInfo)
	}

	// 10. 封装分页信息
	totalPages := (likeCount + int64(pageSize) - 1) / int64(pageSize)
	pagination := model.Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      likeCount,
		TotalPages: int(totalPages),
	}

	// 11. 返回结果
	return &model.VideoListOutput{
		Videos:     likesList,
		Pagination: pagination,
	}, api.CodeSuccess, nil
}
