package video

import (
	"context"
	"errors"
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/video/v1"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"

	"gorm.io/gen/field"
	"gorm.io/gorm"
	"time"
)

// CreateVideo 创建视频
func CreateVideo(ctx context.Context, userId int64, req v1.CreateVideoReq) (*model.CreateVideoOutput, api.ResCode, error) {
	// 1. 查看用户是否存在
	_, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeUserNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 创建视频记录
	newVideo := &model.Video{
		UserID:   userId,
		Title:    *req.Title,
		CoverURL: *req.CoverUrl,
		VideoURL: *req.VideoUrl,
		Duration: *req.Duration,
		Status:   2,
		PublishedAt: time.Now(),
	}
	// 判断可选字段是否为空，如果不为空则赋值
	selectFields := []field.Expr{
		query.Video.Title,
		query.Video.CoverURL,
		query.Video.VideoURL,
		query.Video.Duration,
		query.Video.Status,
		query.Video.UserID,
		query.Video.PublishedAt,
	}
	if req.Description != nil {
		newVideo.Description = *req.Description
		selectFields = append(selectFields, query.Video.Description)
	}
	if req.Width != nil {
		newVideo.Width = *req.Width
		selectFields = append(selectFields, query.Video.Width)
	}
	if req.Height != nil {
		newVideo.Height = *req.Height
		selectFields = append(selectFields, query.Video.Height)
	}
	if req.MusicId != nil {
		newVideo.MusicID = *req.MusicId
		selectFields = append(selectFields, query.Video.MusicID)
	}
	if req.Location != nil {
		newVideo.Location = *req.Location
		selectFields = append(selectFields, query.Video.Location)
	}
	if req.City != nil {
		newVideo.City = *req.City
		selectFields = append(selectFields, query.Video.City)
	}
	if req.Latitude != nil {
		newVideo.Latitude = *req.Latitude
		selectFields = append(selectFields, query.Video.Latitude)
	}
	if req.Longitude != nil {
		newVideo.Longitude = *req.Longitude
		selectFields = append(selectFields, query.Video.Longitude)
	}
	err = query.Video.WithContext(ctx).
		Select(selectFields...).
		Create(newVideo)
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 3. 如果有话题标签，则创建视频话题关联记录
	if req.Topics != nil && len(*req.Topics) > 0 {
		videoTopics := make([]*model.VideoTopic, 0, len(*req.Topics))
		for _, topic := range *req.Topics {
			videoTopics = append(videoTopics, &model.VideoTopic{
				VideoID: newVideo.ID,
				TopicID: topic,
			})
		}
		if err := query.VideoTopic.WithContext(ctx).Create(videoTopics...); err != nil {
			return nil, api.CodeInternalError, err
		}
	}
	// 4. 返回结果
	return &model.CreateVideoOutput{
		VideoID: newVideo.ID,
		Status:  2, // 假设创建成功状态为2
	}, api.CodeSuccess, nil
}

// GetVideo 获取视频详情
func GetVideo(ctx context.Context, videoId int64) (*model.GetVideoOutput, api.ResCode, error) {
	// 1. 查看视频是否存在
	video, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeVideoNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 获取视频详情 如果存在
	// 2.1 获取视频作者信息
	author, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(video.UserID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeUserNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2.2 获取视频话题信息
	videoTopics, err := query.VideoTopic.WithContext(ctx).
		Where(query.VideoTopic.VideoID.Eq(videoId)).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 2.3 获取话题ID列表
	topicIds := make([]int64, 0, len(videoTopics))
	for _, videoTopic := range videoTopics {
		topicIds = append(topicIds, videoTopic.TopicID)
	}
	// 2.4 获取话题信息
	var topicNames []string
	if len(topicIds) > 0 {
		topics, err := query.Topic.WithContext(ctx).
			Where(query.Topic.ID.In(topicIds...)).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		for _, topic := range topics {
			topicNames = append(topicNames, topic.Name)
		}
	}
	// 3. 返回结果
	return &model.GetVideoOutput{
		Video: model.VideoInfo{
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
		},
	}, api.CodeSuccess, nil
}

// DeleteVideo 删除视频
func DeleteVideo(ctx context.Context, videoId int64, userId int64) (api.ResCode, error) {
	// 1. 查看视频是否存在
	video, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.CodeVideoNotExist, err
		}
		return api.CodeInternalError, err
	}
	// 2. 检查视频是否属于当前用户
	if video.UserID != userId {
		return api.CodeNotDeleteOwnVideo, errors.New("不能删除非自己发布的视频")
	}
	// 3. 删除视频记录 软删除
	_, err = query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoId)).
		Delete()
	if err != nil {
		return api.CodeInternalError, err
	}
	return api.CodeSuccess, nil
}	