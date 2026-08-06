package interaction

import (
	"context"
	"errors"
	"fmt"
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/interaction/v1"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	notifSvc "flashvid-platform-gin/internal/service/notification"
	"gorm.io/gorm"
)

// LikeVideo 点赞视频
func LikeVideo(ctx context.Context, userId int64, videoId int64) (*v1.LikeVideoResp, api.ResCode, error) {
	// 1. 检查视频是否存在
	video, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeVideoNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 检查用户是否已经点赞
	countLike, err := query.Like.WithContext(ctx).
		Where(
			query.Like.UserID.Eq(userId), 
			query.Like.TargetType.Eq(1),
			query.Like.TargetID.Eq(videoId),
			).
		Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if countLike > 0 {
		// 已经点赞，幂等：直接返回成功
		return &v1.LikeVideoResp{IsLiked: true, LikeCount: video.LikeCount}, api.CodeSuccess, nil
	}
	// 3. 如果未点赞，则创建点赞记录并更新点赞数 事务包裹
	like := &model.Like{
		UserID:     userId,
		TargetType: 1, // 1表示视频
		TargetID:   videoId,
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 创建点赞记录
		if err = tx.Like.WithContext(ctx).Create(like); err != nil {
			return err
		}
		// 更新视频的点赞数
		_, err = tx.Video.WithContext(ctx).
			Where(query.Video.ID.Eq(videoId)).
			UpdateSimple(query.Video.LikeCount.Add(1))
		if err != nil {
			return err
		}
		// 通知视频作者（不通知自己）
		if video.UserID != userId {
			notifSvc.CreateNotification(ctx, tx, &model.Notification{
				UserID:     video.UserID,
				ActorID:    userId,
				ActionType: 2, // 点赞视频
				TargetType: 2, // 视频
				TargetID:   videoId,
			})
		}
		return nil
	})
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 5. 返回响应
	return &v1.LikeVideoResp{
		IsLiked:   true,
		LikeCount: video.LikeCount + 1,
	}, api.CodeSuccess, nil
}

// UnlikeVideo 取消点赞视频
func UnlikeVideo(ctx context.Context, userId int64, videoId int64) (*v1.LikeVideoResp, api.ResCode, error) {
	// 1. 检查视频是否存在
	video, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeVideoNotExist, err
		}
		return nil, api.CodeInternalError, err
	}
	// 2. 检查用户是否已经点赞
	countLike, err := query.Like.WithContext(ctx).
		Where(
			query.Like.UserID.Eq(userId), 
			query.Like.TargetType.Eq(1),
			query.Like.TargetID.Eq(videoId),
			).
		Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if countLike == 0 {
		// 已经取消点赞，幂等：直接返回成功
		return &v1.LikeVideoResp{IsLiked: false, LikeCount: video.LikeCount}, api.CodeSuccess, nil
	}
	// 3. 如果已经点赞，则删除点赞记录并更新点赞数 事务包裹
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 删除点赞记录（Like 已无 soft-delete，直接物理删除）
		_, err = tx.Like.WithContext(ctx).
			Where(
				query.Like.UserID.Eq(userId),
				query.Like.TargetType.Eq(1),
				query.Like.TargetID.Eq(videoId),
			).
			Delete()
		if err != nil {
			return err
		}
		// 更新视频的点赞数
		_, err = tx.Video.WithContext(ctx).
			Where(query.Video.ID.Eq(videoId)).
			UpdateSimple(query.Video.LikeCount.Sub(1))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 5. 返回响应
	return &v1.LikeVideoResp{
		IsLiked:   false,
		LikeCount: video.LikeCount - 1,
	}, api.CodeSuccess, nil
}

// FavoriteVideo 收藏视频
func FavoriteVideo(ctx context.Context, userId int64, videoId int64) (*v1.FavoriteVideoResp, api.ResCode, error) {
	// 1. 检查视频是否存在
	video, err := query.Video.WithContext(ctx).Where(query.Video.ID.Eq(videoId)).Take()
	if err != nil {
		return nil, api.CodeVideoNotExist, err
	}
	// 2. 检查是否已收藏
	count, err := query.Favorite.WithContext(ctx).
		Where(query.Favorite.UserID.Eq(userId), query.Favorite.VideoID.Eq(videoId)).
		Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if count > 0 {
		// 已收藏，幂等：直接返回成功
		return &v1.FavoriteVideoResp{IsFavorited: true, FavoriteCount: video.FavoriteCount}, api.CodeSuccess, nil
	}
	// 3. 事务：创建收藏记录 + 收藏数 +1 + 通知
	err = query.Q.Transaction(func(tx *query.Query) error {
		if err := tx.Favorite.WithContext(ctx).Create(&model.Favorite{
			UserID:  userId,
			VideoID: videoId,
		}); err != nil {
			return err
		}
		if _, err := tx.Video.WithContext(ctx).Where(tx.Video.ID.Eq(videoId)).
			UpdateSimple(tx.Video.FavoriteCount.Add(1)); err != nil {
			return err
		}
		// 通知视频作者（不通知自己）
		if video.UserID != userId {
			notifSvc.CreateNotification(ctx, tx, &model.Notification{
				UserID:     video.UserID,
				ActorID:    userId,
				ActionType: 3, // 收藏视频
				TargetType: 2, // 视频
				TargetID:   videoId,
			})
		}
		return nil
	})
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 4. 返回结果
	return &v1.FavoriteVideoResp{
		IsFavorited:   true,
		FavoriteCount: video.FavoriteCount + 1,
	}, api.CodeSuccess, nil
}

// UnfavoriteVideo 取消收藏视频
func UnfavoriteVideo(ctx context.Context, userId int64, videoId int64) (*v1.FavoriteVideoResp, api.ResCode, error) {
	// 1. 检查视频是否存在
	video, err := query.Video.WithContext(ctx).Where(query.Video.ID.Eq(videoId)).Take()
	if err != nil {
		return nil, api.CodeVideoNotExist, err
	}
	// 2. 检查是否已收藏
	count, err := query.Favorite.WithContext(ctx).
		Where(query.Favorite.UserID.Eq(userId), query.Favorite.VideoID.Eq(videoId)).
		Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	if count == 0 {
		// 未收藏，幂等：直接返回成功
		return &v1.FavoriteVideoResp{IsFavorited: false, FavoriteCount: video.FavoriteCount}, api.CodeSuccess, nil
	}
	// 3. 事务：级联删除 playlist_videos + 删除收藏记录 + 收藏数 -1
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 3-a. 查出该用户所有包含该视频的播放列表（playlist_videos JOIN playlists WHERE playlists.user_id = userId）
		userPlaylists, err := tx.Playlist.WithContext(ctx).
			Where(tx.Playlist.UserID.Eq(userId)).
			Find()
		if err != nil {
			return err
		}
		if len(userPlaylists) > 0 {
			playlistIDs := make([]int64, 0, len(userPlaylists))
			for _, pl := range userPlaylists {
				playlistIDs = append(playlistIDs, pl.ID)
			}
			// 3-b. 找到需要删除的关联行（每个播放列表最多一条，因为有唯一约束）
			pvRows, err := tx.PlaylistVideo.WithContext(ctx).
				Where(
					tx.PlaylistVideo.PlaylistID.In(playlistIDs...),
					tx.PlaylistVideo.VideoID.Eq(videoId),
				).Find()
			if err != nil {
				return err
			}
			if len(pvRows) > 0 {
				// 3-c. 硬删除 playlist_videos
				if _, err = tx.PlaylistVideo.WithContext(ctx).
					Where(
						tx.PlaylistVideo.PlaylistID.In(playlistIDs...),
						tx.PlaylistVideo.VideoID.Eq(videoId),
					).Delete(); err != nil {
					return err
				}
				// 3-d. 对每个受影响的播放列表 video_count - 1
				for _, pv := range pvRows {
					if _, err = tx.Playlist.WithContext(ctx).
						Where(tx.Playlist.ID.Eq(pv.PlaylistID)).
						UpdateSimple(tx.Playlist.VideoCount.Sub(1)); err != nil {
						return err
					}
				}
			}
		}
		// 3-e. 删除收藏记录（Favorite 已无 soft-delete，直接物理删除）
		if _, err := tx.Favorite.WithContext(ctx).
			Where(tx.Favorite.UserID.Eq(userId), tx.Favorite.VideoID.Eq(videoId)).
			Delete(); err != nil {
			return err
		}
		// 3-f. 视频收藏数 -1
		if _, err := tx.Video.WithContext(ctx).Where(tx.Video.ID.Eq(videoId)).
			UpdateSimple(tx.Video.FavoriteCount.Sub(1)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, api.CodeInternalError, err
	}
	// 4. 返回结果
	return &v1.FavoriteVideoResp{
		IsFavorited:   false,
		FavoriteCount: video.FavoriteCount - 1,
	}, api.CodeSuccess, nil
}

// ShareVideo 分享视频
func ShareVideo(ctx context.Context, videoId int64) (*v1.ShareVideoResp, api.ResCode, error) {
	// 1. 检查视频是否存在
	video, err := query.Video.WithContext(ctx).Where(query.Video.ID.Eq(videoId)).Take()
	if err != nil {
		return nil, api.CodeVideoNotExist, err
	}
	// 2. share_count + 1
	if _, err = query.Video.WithContext(ctx).Where(query.Video.ID.Eq(videoId)).
		UpdateSimple(query.Video.ShareCount.Add(1)); err != nil {
		return nil, api.CodeInternalError, err
	}
	// 3. 返回结果
	return &v1.ShareVideoResp{
		ShareUrl:   fmt.Sprintf("https://flashvid.com/video/%d", videoId),
		ShareCount: video.ShareCount + 1,
	}, api.CodeSuccess, nil
}