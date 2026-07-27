package interaction

import (
	"context"
	"errors"
	"fmt"
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/interaction/v1"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
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
		// 已经点赞，返回特定错误码告知前端
		return nil, api.CodeAlreadyLiked, errors.New("已经点赞过该视频")
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
		// 如果没点赞，返回特定错误码告知前端
		return nil, api.CodeAlreadyUnliked, errors.New("已经取消点赞过该视频")
	}
	// 3. 如果已经点赞，则删除点赞记录并更新点赞数 事务包裹
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 删除点赞记录
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
		return nil, api.CodeAlreadyFavorited, errors.New("已收藏过该视频")
	}
	// 3. 事务：创建收藏记录 + 收藏数 +1
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
		return nil, api.CodeNotFavorited, errors.New("未收藏过该视频")
	}
	// 3. 事务：删除收藏记录 + 收藏数 -1
	err = query.Q.Transaction(func(tx *query.Query) error {
		if _, err := tx.Favorite.WithContext(ctx).
			Where(tx.Favorite.UserID.Eq(userId), tx.Favorite.VideoID.Eq(videoId)).
			Delete(); err != nil {
			return err
		}
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