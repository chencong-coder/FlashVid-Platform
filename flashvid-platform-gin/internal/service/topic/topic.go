package topic

import (
	"context"
	"errors"
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/topic/v1"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	feedsvc "flashvid-platform-gin/internal/service/feed"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// GetTopics 获取话题列表
func GetTopics(ctx context.Context, sort string, cursor string, count int) (*model.TopicListOutput, api.ResCode, error) {
	// 1. 按照sort进行排序 区分游标是否为空分页 返回count条
	q := query.Topic.WithContext(ctx).
			Where(query.Topic.Status.Eq(1)) // 只获取状态为1的有效话题
	if cursor != "" {
		if sort == "latest" {
			cursorTime, err := time.Parse("2006-01-02 15:04:05", cursor)
			if err != nil {
				return nil, api.CodeInvalidParam, err
			}
			q = q.Where(query.Topic.CreatedAt.Lt(cursorTime))
		} else {
			cursorViewCount, err := strconv.ParseInt(cursor, 10, 64)
			if err != nil {
				return nil, api.CodeInvalidParam, err
			}
			q = q.Where(query.Topic.ViewCount.Lt(cursorViewCount))
		}
	}
	if sort == "latest" {
		q = q.Order(query.Topic.CreatedAt.Desc())
	} else {
		// 默认按热门排序
		q = q.Order(query.Topic.ViewCount.Desc())
	}
	topics, err := q.Limit(count + 1).Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	hasMore := len(topics) > count
	if hasMore {
		topics = topics[:count]
	}
	// 2. 封装成model.TopicListOutput
	var topicList []model.TopicInfo
	for _, t := range topics {
		topicList = append(topicList, model.TopicInfo{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			CoverURL:    t.CoverURL,
			ViewCount:   t.ViewCount,
			VideoCount:  t.VideoCount,
			CreatedAt:   t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 3. 生成游标
	nextCursor := ""
	if hasMore {
		last := topics[len(topics)-1]
		if sort == "latest" {
			nextCursor = last.CreatedAt.Format("2006-01-02 15:04:05")
		} else {
			nextCursor = strconv.FormatInt(last.ViewCount, 10)
		}
	}
	// 4. 返回结果
	return &model.TopicListOutput{
		Topics:          topicList,
		NextCursorToken: nextCursor,
		HasMore:         hasMore,
	}, api.CodeSuccess, nil
}

// GetTopicByID 根据话题ID获取话题详情
func GetTopicByID(ctx context.Context, topicId int64) (*v1.GetTopicByIDResp, api.ResCode, error) {
	// 1. 查询话题详情
	topic, err := query.Topic.WithContext(ctx).
					Where(query.Topic.ID.Eq(topicId)).
					First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeTopicNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 异步更新浏览量
	go func() {
		_, _ = query.Topic.WithContext(context.Background()).
			Where(query.Topic.ID.Eq(topicId)).
			UpdateSimple(query.Topic.ViewCount.Add(1))
	}()
	// 3. 封装成model.TopicInfo
	topicInfo := model.TopicInfo{
		ID:          topic.ID,
		Name:        topic.Name,
		Description: topic.Description,
		CoverURL:    topic.CoverURL,
		ViewCount:   topic.ViewCount,
		VideoCount:  topic.VideoCount,
		CreatedAt:   topic.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	// 3. 返回结果
	return &v1.GetTopicByIDResp{
		Topic: topicInfo,
	}, api.CodeSuccess, nil
}

// GetTopicVideos 根据话题ID获取话题下的视频列表
func GetTopicVideos(ctx context.Context, topicId int64, viewerID int64, sort string, cursor string, count int) (*model.GetTopicVideosOutput, api.ResCode, error) {
	// 1. 查询话题是否存在
	_, err := query.Topic.WithContext(ctx).
		Where(
			query.Topic.ID.Eq(topicId),
			query.Topic.Status.Eq(1),
			).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeTopicNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 查询话题下的视频列表
	// 2.1 先根据topicid去查询topic_video表，获取视频id列表，再根据视频id列表去查询视频表，获取视频详情
	var videoIds []int64
	videoIdToTopicIds := make(map[int64][]int64) // videoId -> topicIds
	videoTopics, err := query.VideoTopic.WithContext(ctx).
		Where(query.VideoTopic.TopicID.Eq(topicId)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	for _, vt := range videoTopics {
		videoIds = append(videoIds, vt.VideoID)
		videoIdToTopicIds[vt.VideoID] = append(videoIdToTopicIds[vt.VideoID], vt.TopicID)
	}
	// 2.2 加上cursor和sort条件 根据视频id列表去查询视频表，获取视频详情
	q := query.Video.WithContext(ctx).
		Where(
			query.Video.ID.In(videoIds...),
			query.Video.Status.Eq(2), // 2 = 已发布，与其他 feed 保持一致
		)
	if cursor != "" {
		if sort == "popular" {
			cursorLikeCount, err := strconv.ParseInt(cursor, 10, 64)
			if err != nil {
				return nil, api.CodeInvalidParam, err
			}
			q = q.Where(query.Video.LikeCount.Lt(int32(cursorLikeCount)))
		} else {
			cursorTime, err := time.Parse("2006-01-02 15:04:05", cursor)
			if err != nil {
				return nil, api.CodeInvalidParam, err
			}
			q = q.Where(query.Video.PublishedAt.Lt(cursorTime))
		}
	}
	if sort == "popular" {
		q = q.Order(
			query.Video.LikeCount.Desc(),
			query.Video.ViewCount.Desc(),
		)
	} else {
		// 默认按最新排序
		q = q.Order(query.Video.PublishedAt.Desc())
	}
	videos, err := q.Limit(count + 1).Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	hasMore := len(videos) > count
	if hasMore {
		videos = videos[:count]
	}
	// 2.3 构建videoId -> VideoInfo的映射
	videosIdMap := make(map[int64]*model.Video)
	var authorIds []int64
	for _, v := range videos {
		videosIdMap[v.ID] = v
		authorIds = append(authorIds, v.UserID)
	}
	// 2.4 构建authorId -> UserInfo的映射
	authorsIdMap := make(map[int64]*model.User)
	authors, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(authorIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	for _, a := range authors {
		authorsIdMap[a.ID] = a
	}
	// 2.5 批量查询所有涉及的话题名称（一次查询，不用N次）
	allTopicIdSet := make(map[int64]struct{})
	for _, topicIds := range videoIdToTopicIds {
		for _, id := range topicIds {
			allTopicIdSet[id] = struct{}{} // 相当于一个Set，去重
		}
	}
	topicIdToName := make(map[int64]string)
	if len(allTopicIdSet) > 0 {
		allTopicIds := make([]int64, 0, len(allTopicIdSet))
		for id := range allTopicIdSet {
			allTopicIds = append(allTopicIds, id)
		}
		allTopics, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.In(allTopicIds...)).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		for _, t := range allTopics {
			topicIdToName[t.ID] = t.Name
		}
	}
	// 3. 封装成model.GetTopicVideosOutput
	var videoList []model.VideoInfo
	for _, v := range videos {
		author, exists := authorsIdMap[v.UserID]
		if !exists {
			continue
		}
		var topicNames []string
		if topicIds, ok := videoIdToTopicIds[v.ID]; ok {
			for _, topicId := range topicIds {
				if name, ok := topicIdToName[topicId]; ok {
					topicNames = append(topicNames, name)
				}
			}
		}
		videoList = append(videoList, model.VideoInfo{
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
			Author: model.VideoAuthor{
				ID:       author.ID,
				Username: author.Username,
				Nickname: author.Nickname,
				Avatar:   author.Avatar,
			},
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
	// 4. 回填观看者互动状态（点赞/收藏/关注）
	if err := feedsvc.AttachViewerState(ctx, viewerID, videoList); err != nil {
		return nil, api.CodeInternalError, err
	}
	// 5. 生成游标
	nextCursor := ""
	if hasMore {
		last := videos[len(videos)-1]
		if sort == "popular" {
			nextCursor = strconv.FormatInt(int64(last.LikeCount), 10)
		} else {
			nextCursor = last.PublishedAt.Format("2006-01-02 15:04:05")
		}
	}
	// 5. 返回结果
	return &model.GetTopicVideosOutput{
		Videos:          videoList,
		NextCursorToken: nextCursor,
		HasMore:         hasMore,
	}, api.CodeSuccess, nil
}


// SearchTopics 根据关键词搜索话题
func SearchTopics(ctx context.Context, keyword string, cursor string, count int) (*model.TopicListOutput, api.ResCode, error) {
	q := query.Topic.WithContext(ctx).
		Where(query.Topic.Status.Eq(1)).
		Where(query.Topic.Name.Like("%" + keyword + "%"))
	if cursor != "" {
		cursorVideoCount, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			return nil, api.CodeInvalidParam, err
		}
		q = q.Where(query.Topic.VideoCount.Lt(int32(cursorVideoCount)))
	}
	q = q.Order(query.Topic.VideoCount.Desc())
	topics, err := q.Limit(count + 1).Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	hasMore := len(topics) > count
	if hasMore {
		topics = topics[:count]
	}
	// 2. 封装成model.TopicListOutput
	var topicList []model.TopicInfo
	for _, t := range topics {
		topicList = append(topicList, model.TopicInfo{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			CoverURL:    t.CoverURL,
			ViewCount:   t.ViewCount,
			VideoCount:  t.VideoCount,
			CreatedAt:   t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 3. 生成游标
	nextCursor := ""
	if hasMore {
		nextCursor = strconv.FormatInt(int64(topics[len(topics)-1].VideoCount), 10)
	}
	// 4. 返回结果
	return &model.TopicListOutput{
		Topics:          topicList,
		NextCursorToken: nextCursor,
		HasMore:         hasMore,
	}, api.CodeSuccess, nil
}