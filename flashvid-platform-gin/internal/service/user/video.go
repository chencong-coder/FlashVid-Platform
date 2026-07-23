package user

import (
	"context"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"errors"
	"gorm.io/gorm"
)

// 获取用户视频列表服务
func GetUserVideos(ctx context.Context, userId int64, page, pageSize int) (*model.VideoListOutput, api.ResCode, error) {
	// 1. 查询用户是否存在
	user, err := query.User.WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeUserNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 查视频总数 video表 
	// user_id = userId AND status = 2(已发布) AND deleted_at IS NULL
	videoCount, err := query.Video.WithContext(ctx).
		Where(query.Video.UserID.Eq(userId), query.Video.Status.Eq(2), query.Video.DeletedAt.IsNull()).
		Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 3. 分页查询视频列表
	//  created_at DESC（最新在前） Limit(pageSize).Offset((page-1)*pageSize)
	var videos []*model.Video
	videos, err = query.Video.WithContext(ctx).
		Where(query.Video.UserID.Eq(userId), query.Video.Status.Eq(2), query.Video.DeletedAt.IsNull()).
		Order(query.Video.CreatedAt.Desc()).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 4. 如果视频列表为空 直接返回
	if len(videos) == 0 {
		return &model.VideoListOutput{
			Videos:     []model.VideoInfo{},
			Pagination: model.Pagination{Page: page, PageSize: pageSize, Total: 0, TotalPages: 0},
		}, api.CodeSuccess, nil
	}
	// 5. 封装返回结果
	var videoInfos []model.VideoInfo
	//  每个视频查video_topic表得到视频对应的话题id 再查topic表的到话题名称
	for _, video := range videos {
		// 查video_topic表得到话题id
		var topicIds []int64
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
				for _, topic := range videoTopics {
				topicIds = append(topicIds, topic.TopicID)
			}
			// 查topic表得到话题名称
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
			Author:      model.VideoAuthor{
				ID:       user.ID,
				Username: user.Username,
				Avatar:   user.Avatar,
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
		videoInfos = append(videoInfos, videoInfo)
	}
	// 6. 计算总页数
		// 向上取整 当videoCount为0时 totalPages为0 
	totalPages := (videoCount + int64(pageSize) - 1) / int64(pageSize)
		// 封装分页信息
	pagination := model.Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      videoCount,
		TotalPages: int(totalPages),
	}
	// 7.封装返回结果
	return &model.VideoListOutput{
		Videos:     videoInfos,
		Pagination: pagination,
	}, api.CodeSuccess, nil
}