package feed

import (
	"context"
	"errors"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"time"
	"math"
)

// GetFeedRecommend 获取推荐视频流
func GetFeedRecommend(ctx context.Context, cursor string, count int) (*model.FeedOutput, api.ResCode, error) {
	// 1. 获取推荐视频流 查已发布 并且按发布时间降序
	var err error
	var videos []*model.Video
	if cursor == "" {
		videos, err = query.Video.WithContext(ctx).
			Where(query.Video.Status.Eq(2)).
			Order(query.Video.PublishedAt.Desc(), query.Video.ID.Desc()).
			Limit(count).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		// 如果没有搜索到视频，返回空列表和空游标
		if len(videos) == 0 {
			return &model.FeedOutput{
				Videos:          []model.VideoInfo{},
				NextCursorToken: "",
				HasMore:         false,
			}, api.CodeSuccess, nil
		}
	} else {
		// 2. 如果有游标，则按游标获取下一页数据
		// 将游标转换为时间戳
		cursorTime, err := time.Parse("2006-01-02 15:04:05", cursor)
		if err != nil {
			return nil, api.CodeInvalidParam, err
		}
		// 查询发布时间小于游标的已发布视频，并按发布时间降序排序 如果发布时间相同，则按ID降序排序，确保分页的稳定性
		videos, err = query.Video.WithContext(ctx).
			Where(query.Video.Status.Eq(2), query.Video.PublishedAt.Lt(cursorTime)).
			Order(query.Video.PublishedAt.Desc(), query.Video.ID.Desc()).
			Limit(count).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		// 如果没有搜索到视频，返回空列表和空游标
		if len(videos) == 0 {
			return &model.FeedOutput{
				Videos:          []model.VideoInfo{},
				NextCursorToken: "",
				HasMore:         false,
			}, api.CodeSuccess, nil
		}
	}
	// 3. 构建输出结果 批量查询
	// 构建videoid -> videoInfo的映射
	var videosMap = make(map[int64]*model.Video)
	var videoIds []int64
	var authorIds []int64
	for _, video := range videos {
		videoIds = append(videoIds, video.ID)
		videosMap[video.ID] = video
		authorIds = append(authorIds, video.UserID)
	}
	// authorIds -> userInfo的映射
	authors, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(authorIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	var authorsMap = make(map[int64]*model.User)
	for _, author := range authors {
		authorsMap[author.ID] = author
	}
	// video-> topicids的映射
	var topidIds []int64
	var videoIdToTopicIds = make(map[int64][]int64)
	videoTopics, err := query.VideoTopic.WithContext(ctx).
		Where(query.VideoTopic.VideoID.In(videoIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	for _, videoTopic := range videoTopics {
		videoIdToTopicIds[videoTopic.VideoID] = append(videoIdToTopicIds[videoTopic.VideoID], videoTopic.TopicID)
		topidIds = append(topidIds, videoTopic.TopicID)
	}
	// topicid -> topicName的映射
	var topicMap = make(map[int64]string)
	if len(topidIds) > 0 {
		topics, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.In(topidIds...)).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		for _, topic := range topics {
			topicMap[topic.ID] = topic.Name
		}
	}
	// 构建输出
	var outputVideos []model.VideoInfo
	for _, video := range videos {
		author, ok := authorsMap[video.UserID]
		if !ok {
			return nil, api.CodeInternalError, errors.New("作者信息不存在")
		}
		topicNames := make([]string, 0)
		for _, topicId := range videoIdToTopicIds[video.ID] {
			if topicName, ok := topicMap[topicId]; ok {
				topicNames = append(topicNames, topicName)
			}
		}
		outputVideos = append(outputVideos, model.VideoInfo{
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
				Nickname: author.Nickname,
			},
			Stats: model.VideoStats{
				ViewCount:     video.ViewCount,
				LikeCount:     video.LikeCount,
				CommentCount:  video.CommentCount,
				ShareCount:    video.ShareCount,
				FavoriteCount: video.FavoriteCount,
			},
			PublishedAt: video.PublishedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 4. 返回视频流和下一个游标
	hasMore := false
	nextCursor := ""
	if len(videos) == count {
		hasMore = true
		lastVideo := videos[len(videos)-1]
		nextCursor = lastVideo.PublishedAt.Format("2006-01-02 15:04:05")
	}

	return &model.FeedOutput{
		Videos:          outputVideos,
		NextCursorToken: nextCursor,
		HasMore:         hasMore,
	}, api.CodeSuccess, nil
}

// GetFeedFollow 获取关注视频流
func GetFeedFollow(ctx context.Context, userId int64, cursor string, count int) (*model.FeedOutput, api.ResCode, error) {
	var err error
	var follows []*model.Follow
	var followerIds []int64
	// 1. 获取用户关注id列表
	follows, err = query.Follow.WithContext(ctx).
		Where(query.Follow.FollowerID.Eq(userId)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if len(follows) == 0 {
		return &model.FeedOutput{
			Videos:          []model.VideoInfo{},
			NextCursorToken: "",
			HasMore:         false,
	}, api.CodeSuccess, nil
	}
	for _, follow := range follows {
		followerIds = append(followerIds, follow.FollowingID)
	}
	// 2. 获取关注视频流 查已发布 并且按发布时间降序
	var videos []*model.Video
	if cursor == "" {
		videos, err = query.Video.WithContext(ctx).
			Where(query.Video.Status.Eq(2), query.Video.UserID.In(followerIds...)).
			Order(query.Video.PublishedAt.Desc(), query.Video.ID.Desc()).
			Limit(count).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		// 如果没有搜索到视频，返回空列表和空游标
		if len(videos) == 0 {
			return &model.FeedOutput{
				Videos:          []model.VideoInfo{},
				NextCursorToken: "",
				HasMore:         false,
			}, api.CodeSuccess, nil
		}
	} else {
		// 3. 如果有游标，则按游标获取下一页数据
		// 将游标转换为时间戳
		cursorTime, err := time.Parse("2006-01-02 15:04:05", cursor)
		if err != nil {
			return nil, api.CodeInvalidParam, err
		}
		// 查询发布时间小于游标的已发布视频，并按发布时间降序排序 如果发布时间相同，则按ID降序排序，确保分页的稳定性
		videos, err = query.Video.WithContext(ctx).
			Where(query.Video.Status.Eq(2), query.Video.PublishedAt.Lt(cursorTime), query.Video.UserID.In(followerIds...)).
			Order(query.Video.PublishedAt.Desc(), query.Video.ID.Desc()).
			Limit(count).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		// 如果没有搜索到视频，返回空列表和空游标
		if len(videos) == 0 {
			return &model.FeedOutput{
				Videos:          []model.VideoInfo{},
				NextCursorToken: "",
				HasMore:         false,
			}, api.CodeSuccess, nil
		}
	}
	// 4. 构建输出结果 批量查询
	// 构建videoid -> videoInfo的映射
	var videosMap = make(map[int64]*model.Video)
	var videoIds []int64
	var authorIds []int64
	for _, video := range videos {
		videoIds = append(videoIds, video.ID)
		videosMap[video.ID] = video
		authorIds = append(authorIds, video.UserID)
	}
	// authorIds -> userInfo的映射
	authors, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(authorIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	var authorsMap = make(map[int64]*model.User)
	for _, author := range authors {
		authorsMap[author.ID] = author
	}
	// video-> topicids的映射
	var topidIds []int64
	var videoIdToTopicIds = make(map[int64][]int64)
	videoTopics, err := query.VideoTopic.WithContext(ctx).
		Where(query.VideoTopic.VideoID.In(videoIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	for _, videoTopic := range videoTopics {
		videoIdToTopicIds[videoTopic.VideoID] = append(videoIdToTopicIds[videoTopic.VideoID], videoTopic.TopicID)
		topidIds = append(topidIds, videoTopic.TopicID)
	}
	// topicid -> topicName的映射
	var topicMap = make(map[int64]string)
	if len(topidIds) > 0 {
		topics, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.In(topidIds...)).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		for _, topic := range topics {
			topicMap[topic.ID] = topic.Name
		}
	}
	// 构建输出
	var outputVideos []model.VideoInfo
	for _, video := range videos {
		author, ok := authorsMap[video.UserID]
		if !ok {
			return nil, api.CodeInternalError, errors.New("作者信息不存在")
		}
		topicNames := make([]string, 0)
		for _, topicId := range videoIdToTopicIds[video.ID] {
			if topicName, ok := topicMap[topicId]; ok {
				topicNames = append(topicNames, topicName)
			}
		}
		outputVideos = append(outputVideos, model.VideoInfo{
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
				Nickname: author.Nickname,
			},
			Stats: model.VideoStats{
				ViewCount:     video.ViewCount,
				LikeCount:     video.LikeCount,
				CommentCount:  video.CommentCount,
				ShareCount:    video.ShareCount,
				FavoriteCount: video.FavoriteCount,
			},
			PublishedAt: video.PublishedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 5. 返回视频流和下一个游标
	hasMore := false
	nextCursor := ""
	if len(videos) == count {
		hasMore = true
		lastVideo := videos[len(videos)-1]
		nextCursor = lastVideo.PublishedAt.Format("2006-01-02 15:04:05")
	}

	return &model.FeedOutput{
		Videos:          outputVideos,
		NextCursorToken: nextCursor,
		HasMore:         hasMore,
	}, api.CodeSuccess, nil
}

// GetFeedFriends 获取好友视频流（只含互相关注的用户）
func GetFeedFriends(ctx context.Context, userId int64, cursor string, count int) (*model.FeedOutput, api.ResCode, error) {
	var err error
	// 1. 获取我关注的人
	following, err := query.Follow.WithContext(ctx).
		Where(query.Follow.FollowerID.Eq(userId)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if len(following) == 0 {
		return &model.FeedOutput{
			Videos:          []model.VideoInfo{},
			NextCursorToken: "",
			HasMore:         false,
		}, api.CodeSuccess, nil
	}
	// 2. 获取关注我的人
	followers, err := query.Follow.WithContext(ctx).
		Where(query.Follow.FollowingID.Eq(userId)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if len(followers) == 0 {
		return &model.FeedOutput{
			Videos:          []model.VideoInfo{},
			NextCursorToken: "",
			HasMore:         false,
		}, api.CodeSuccess, nil
	}
	// 3. 求交集：既被我关注、又关注我的，即互相关注的好友
	followingSet := make(map[int64]struct{}, len(following))
	for _, f := range following {
		followingSet[f.FollowingID] = struct{}{}
	}
	var friendIds []int64
	for _, f := range followers {
		if _, ok := followingSet[f.FollowerID]; ok {
			friendIds = append(friendIds, f.FollowerID)
		}
	}
	if len(friendIds) == 0 {
		return &model.FeedOutput{
			Videos:          []model.VideoInfo{},
			NextCursorToken: "",
			HasMore:         false,
		}, api.CodeSuccess, nil
	}
	// 4. 获取好友视频流 查已发布 并且按发布时间降序
	var videos []*model.Video
	if cursor == "" {
		videos, err = query.Video.WithContext(ctx).
			Where(query.Video.Status.Eq(2), query.Video.UserID.In(friendIds...)).
			Order(query.Video.PublishedAt.Desc(), query.Video.ID.Desc()).
			Limit(count).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		if len(videos) == 0 {
			return &model.FeedOutput{
				Videos:          []model.VideoInfo{},
				NextCursorToken: "",
				HasMore:         false,
			}, api.CodeSuccess, nil
		}
	} else {
		// 如果有游标，则按游标获取下一页数据
		cursorTime, err := time.Parse("2006-01-02 15:04:05", cursor)
		if err != nil {
			return nil, api.CodeInvalidParam, err
		}
		videos, err = query.Video.WithContext(ctx).
			Where(query.Video.Status.Eq(2), query.Video.PublishedAt.Lt(cursorTime), query.Video.UserID.In(friendIds...)).
			Order(query.Video.PublishedAt.Desc(), query.Video.ID.Desc()).
			Limit(count).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		if len(videos) == 0 {
			return &model.FeedOutput{
				Videos:          []model.VideoInfo{},
				NextCursorToken: "",
				HasMore:         false,
			}, api.CodeSuccess, nil
		}
	}
	// 5. 构建输出结果 批量查询
	var videoIds []int64
	var authorIds []int64
	for _, video := range videos {
		videoIds = append(videoIds, video.ID)
		authorIds = append(authorIds, video.UserID)
	}
	// authorIds -> userInfo的映射
	authors, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(authorIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	var authorsMap = make(map[int64]*model.User)
	for _, author := range authors {
		authorsMap[author.ID] = author
	}
	// video-> topicids的映射
	var topidIds []int64
	var videoIdToTopicIds = make(map[int64][]int64)
	videoTopics, err := query.VideoTopic.WithContext(ctx).
		Where(query.VideoTopic.VideoID.In(videoIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	for _, videoTopic := range videoTopics {
		videoIdToTopicIds[videoTopic.VideoID] = append(videoIdToTopicIds[videoTopic.VideoID], videoTopic.TopicID)
		topidIds = append(topidIds, videoTopic.TopicID)
	}
	// topicid -> topicName的映射
	var topicMap = make(map[int64]string)
	if len(topidIds) > 0 {
		topics, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.In(topidIds...)).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		for _, topic := range topics {
			topicMap[topic.ID] = topic.Name
		}
	}
	// 构建输出
	var outputVideos []model.VideoInfo
	for _, video := range videos {
		author, ok := authorsMap[video.UserID]
		if !ok {
			return nil, api.CodeInternalError, errors.New("作者信息不存在")
		}
		topicNames := make([]string, 0)
		for _, topicId := range videoIdToTopicIds[video.ID] {
			if topicName, ok := topicMap[topicId]; ok {
				topicNames = append(topicNames, topicName)
			}
		}
		outputVideos = append(outputVideos, model.VideoInfo{
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
				Nickname: author.Nickname,
			},
			Stats: model.VideoStats{
				ViewCount:     video.ViewCount,
				LikeCount:     video.LikeCount,
				CommentCount:  video.CommentCount,
				ShareCount:    video.ShareCount,
				FavoriteCount: video.FavoriteCount,
			},
			PublishedAt: video.PublishedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 6. 返回视频流和下一个游标
	hasMore := false
	nextCursor := ""
	if len(videos) == count {
		hasMore = true
		lastVideo := videos[len(videos)-1]
		nextCursor = lastVideo.PublishedAt.Format("2006-01-02 15:04:05")
	}

	return &model.FeedOutput{
		Videos:          outputVideos,
		NextCursorToken: nextCursor,
		HasMore:         hasMore,
	}, api.CodeSuccess, nil
}

// GetFeedNearby 获取附近视频流
func GetFeedNearby(ctx context.Context, latitude, longitude float64, distance int, cursor string, count int) (*model.FeedOutput, api.ResCode, error) {
	// 1. 计算经纬度范围
	latRange := float64(distance) / 111.0 // 1度纬度约等于111公里
	lngRange := float64(distance) / (111.0 * math.Cos(latitude*math.Pi/180.0)) // 经度范围根据纬度计算
	minLat := latitude - latRange
	maxLat := latitude + latRange
	minLng := longitude - lngRange
	maxLng := longitude + lngRange
	// 2. 获取附近视频流 查已发布 并且按发布时间降序 如果发布时间相同，则按ID降序排序，确保分页的稳定性
	var err error
	var videos []*model.Video
	if cursor == "" {
		videos, err = query.Video.WithContext(ctx).
			Where(
				query.Video.Status.Eq(2),
				query.Video.Latitude.Between(minLat, maxLat),
				query.Video.Longitude.Between(minLng, maxLng),
			).
			Order(
				query.Video.PublishedAt.Desc(),
				query.Video.ID.Desc(),
				).
			Limit(count).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		// 如果没有搜索到视频，返回空列表和空游标
		if len(videos) == 0 {
			return &model.FeedOutput{
				Videos:          []model.VideoInfo{},
				NextCursorToken: "",
				HasMore:         false,
			}, api.CodeSuccess, nil
		}
	} else {
		// 2. 如果有游标，则按游标获取下一页数据
		// 将游标转换为时间戳
		cursorTime, err := time.Parse("2006-01-02 15:04:05", cursor)
		if err != nil {
			return nil, api.CodeInvalidParam, err
		}
		// 查询发布时间小于游标的已发布视频，并按发布时间降序排序 如果发布时间相同，则按ID降序排序，确保分页的稳定性
		videos, err = query.Video.WithContext(ctx).
			Where(
				query.Video.Status.Eq(2), 
				query.Video.PublishedAt.Lt(cursorTime),
				query.Video.Latitude.Between(minLat, maxLat),
				query.Video.Longitude.Between(minLng, maxLng),
				).
			Order(
				query.Video.PublishedAt.Desc(),
				query.Video.ID.Desc(),
				).
			Limit(count).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		// 如果没有搜索到视频，返回空列表和空游标
		if len(videos) == 0 {
			return &model.FeedOutput{
				Videos:          []model.VideoInfo{},
				NextCursorToken: "",
				HasMore:         false,
			}, api.CodeSuccess, nil
		}
	}
	// 3. 构建输出结果 批量查询
	// 构建videoid -> videoInfo的映射
	var videosMap = make(map[int64]*model.Video)
	var videoIds []int64
	var authorIds []int64
	for _, video := range videos {
		videoIds = append(videoIds, video.ID)
		videosMap[video.ID] = video
		authorIds = append(authorIds, video.UserID)
	}
	// authorIds -> userInfo的映射
	authors, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(authorIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	var authorsMap = make(map[int64]*model.User)
	for _, author := range authors {
		authorsMap[author.ID] = author
	}
	// video-> topicids的映射
	var topidIds []int64
	var videoIdToTopicIds = make(map[int64][]int64)
	videoTopics, err := query.VideoTopic.WithContext(ctx).
		Where(query.VideoTopic.VideoID.In(videoIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	for _, videoTopic := range videoTopics {
		videoIdToTopicIds[videoTopic.VideoID] = append(videoIdToTopicIds[videoTopic.VideoID], videoTopic.TopicID)
		topidIds = append(topidIds, videoTopic.TopicID)
	}
	// topicid -> topicName的映射
	var topicMap = make(map[int64]string)
	if len(topidIds) > 0 {
		topics, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.In(topidIds...)).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		for _, topic := range topics {
			topicMap[topic.ID] = topic.Name
		}
	}
	// 构建输出
	var outputVideos []model.VideoInfo
	for _, video := range videos {
		author, ok := authorsMap[video.UserID]
		if !ok {
			return nil, api.CodeInternalError, errors.New("作者信息不存在")
		}
		topicNames := make([]string, 0)
		for _, topicId := range videoIdToTopicIds[video.ID] {
			if topicName, ok := topicMap[topicId]; ok {
				topicNames = append(topicNames, topicName)
			}
		}
		outputVideos = append(outputVideos, model.VideoInfo{
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
				Nickname: author.Nickname,
			},
			Stats: model.VideoStats{
				ViewCount:     video.ViewCount,
				LikeCount:     video.LikeCount,
				CommentCount:  video.CommentCount,
				ShareCount:    video.ShareCount,
				FavoriteCount: video.FavoriteCount,
			},
			PublishedAt: video.PublishedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 4. 返回视频流和下一个游标
	hasMore := false
	nextCursor := ""
	if len(videos) == count {
		hasMore = true
		lastVideo := videos[len(videos)-1]
		nextCursor = lastVideo.PublishedAt.Format("2006-01-02 15:04:05")
	}

	return &model.FeedOutput{
		Videos:          outputVideos,
		NextCursorToken: nextCursor,
		HasMore:         hasMore,
	}, api.CodeSuccess, nil
}