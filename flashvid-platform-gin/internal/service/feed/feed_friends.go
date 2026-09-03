package feed

import (
	"context"
	"errors"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"fmt"
	"strconv"
	"time"
)

// GetFeedFriends1 获取好友视频流（混合模式：Redis List 推模式 + MySQL 补全大V视频）
func GetFeedFriends1(ctx context.Context, userId int64, cursor string, count int) (*model.FeedOutput, api.ResCode, error) {
	// Redis 不可用时降级到 MySQL 拉模式
	if rdb == nil {
		return GetFeedFriends(ctx, userId, cursor, count)
	}

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

	// 4. 分离普通用户和大V（查询每个好友的粉丝数）
	var normalFriendIDs []int64  // 普通用户（粉丝 < 10000）
	var bigVFriendIDs []int64    // 大V（粉丝 >= 10000）

	for _, friendID := range friendIds {
		followerCount, err := query.Follow.WithContext(ctx).
			Where(query.Follow.FollowingID.Eq(friendID)).
			Count()
		if err != nil {
			continue
		}
		if followerCount >= 10000 {
			bigVFriendIDs = append(bigVFriendIDs, friendID)
		} else {
			normalFriendIDs = append(normalFriendIDs, friendID)
		}
	}

	// 5. 从 Redis List 读取视频 ID（普通好友的视频，推模式预写的）
	friendFeedKey := fmt.Sprintf("feed:friends:%d", userId)
	start := int64(0)
	if cursor != "" {
		parsedStart, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			return nil, api.CodeInvalidParam, err
		}
		start = parsedStart
	}

	videoIDStrs, err := rdb.LRange(ctx, friendFeedKey, start, start+int64(count*2)).Result() // 多读一些，后面合并
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	// 6. 如果 Redis List 为空且没有大V好友，降级到完全拉模式
	if len(videoIDStrs) == 0 && len(bigVFriendIDs) == 0 {
		return GetFeedFriends(ctx, userId, cursor, count)
	}

	// 7. 转换 Redis 中的视频 ID
	redisVideoIDs := make([]int64, 0, len(videoIDStrs))
	for _, idStr := range videoIDStrs {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		redisVideoIDs = append(redisVideoIDs, id)
	}

	// 8. 从 MySQL 查询大V好友的最新视频（补全）
	var bigVVideos []*model.Video
	if len(bigVFriendIDs) > 0 {
		bigVVideos, err = query.Video.WithContext(ctx).
			Where(query.Video.Status.Eq(2), query.Video.UserID.In(bigVFriendIDs...)).
			Order(query.Video.PublishedAt.Desc(), query.Video.ID.Desc()).
			Limit(count).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
	}

	// 9. 合并 Redis 视频 ID 和大V视频 ID
	allVideoIDs := make([]int64, 0, len(redisVideoIDs)+len(bigVVideos))
	allVideoIDs = append(allVideoIDs, redisVideoIDs...)
	for _, v := range bigVVideos {
		allVideoIDs = append(allVideoIDs, v.ID)
	}

	if len(allVideoIDs) == 0 {
		return &model.FeedOutput{
			Videos:          []model.VideoInfo{},
			NextCursorToken: "",
			HasMore:         false,
		}, api.CodeSuccess, nil
	}

	// 10. 批量查 MySQL（统一查询所有视频详情）
	videos, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.In(allVideoIDs...), query.Video.Status.Eq(2)).
		Order(query.Video.PublishedAt.Desc(), query.Video.ID.Desc()). // 按时间排序
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

	// 11. 构建 videoID -> Video 映射
	videoMap := make(map[int64]*model.Video)
	authorIds := make([]int64, 0, len(videos))
	videoIds := make([]int64, 0, len(videos))
	for _, video := range videos {
		videoMap[video.ID] = video
		authorIds = append(authorIds, video.UserID)
		videoIds = append(videoIds, video.ID)
	}

	// 12. 批量查询作者信息
	authors, err := query.User.WithContext(ctx).
		Where(query.User.ID.In(authorIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	authorMap := make(map[int64]*model.User)
	for _, author := range authors {
		authorMap[author.ID] = author
	}

	// 13. 批量查询视频话题
	videoTopics, err := query.VideoTopic.WithContext(ctx).
		Where(query.VideoTopic.VideoID.In(videoIds...)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	videoTopicMap := make(map[int64][]int64)
	topicIds := make([]int64, 0)
	for _, vt := range videoTopics {
		videoTopicMap[vt.VideoID] = append(videoTopicMap[vt.VideoID], vt.TopicID)
		topicIds = append(topicIds, vt.TopicID)
	}

	// 14. 批量查询话题信息
	topicMap := make(map[int64]string)
	if len(topicIds) > 0 {
		topics, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.In(topicIds...)).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		for _, topic := range topics {
			topicMap[topic.ID] = topic.Name
		}
	}

	// 15. 按发布时间顺序构建输出（MySQL 已排序）
	outputVideos := make([]model.VideoInfo, 0, len(videos))
	for _, video := range videos {
		author, ok := authorMap[video.UserID]
		if !ok {
			continue // 作者可能已删除
		}

		topicNames := make([]string, 0)
		for _, topicID := range videoTopicMap[video.ID] {
			if name, ok := topicMap[topicID]; ok {
				topicNames = append(topicNames, name)
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
				Nickname: author.Nickname,
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
		})
	}

	// 16. 附加当前用户的互动状态
	if err = AttachViewerState(ctx, userId, outputVideos); err != nil {
		return nil, api.CodeInternalError, err
	}

	// 17. 生成游标和判断是否有更多（简化：基于返回数量）
	hasMore := len(outputVideos) >= count
	nextCursor := ""
	if hasMore {
		outputVideos = outputVideos[:count]
		nextCursor = strconv.FormatInt(start+int64(count), 10)
	}

	return &model.FeedOutput{
		Videos:          outputVideos,
		NextCursorToken: nextCursor,
		HasMore:         hasMore,
	}, api.CodeSuccess, nil
}

// GetFeedFriends 获取好友视频流（拉模式：MySQL 实时查询，降级方案）
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
	// 6. 附加当前用户的互动状态（点赞/收藏/关注）
	if err = AttachViewerState(ctx, userId, outputVideos); err != nil {
		return nil, api.CodeInternalError, err
	}
	// 7. 返回视频流和下一个游标
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
