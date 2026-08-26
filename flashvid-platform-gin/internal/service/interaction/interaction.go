package interaction

import (
	"context"
	"errors"
	"fmt"
	"flashvid-platform-gin/api"
	v1 "flashvid-platform-gin/api/interaction/v1"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/pkg/hotrank"
	notifSvc "flashvid-platform-gin/internal/service/notification"
	"gorm.io/gorm"
	"strconv"
)

var rdb = dao.RedisClient

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
	// 4. 异步更新 Redis 热度（重新计算带时间衰减的分数）
	go func(vid int64) {
		hotrank.UpdateVideoHotScore(context.Background(), vid)
	}(videoId)
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
	// 4. 异步更新 Redis 热度（重新计算带时间衰减的分数）
	go func(vid int64) {
		hotrank.UpdateVideoHotScore(context.Background(), vid)
	}(videoId)
	// 5. 返回响应
	return &v1.LikeVideoResp{
		IsLiked:   false,
		LikeCount: video.LikeCount - 1,
	}, api.CodeSuccess, nil
}

// LikeVideo1 点赞视频（Redis 优化版）
// 优化点：
// 1. 点赞状态用 Redis Set 存储，避免 COUNT 查询和软删唯一键冲突
// 2. 点赞计数用 Redis Hash 累积，定时任务批量刷回 MySQL（解决热点行写）
// 3. 通知暂保留同步（阶段 3 用 MQ 异步化）
func LikeVideo1(ctx context.Context, userId int64, videoId int64) (*v1.LikeVideoResp, api.ResCode, error) {
	if rdb == nil {
		// 降级：Redis 不可用时走原逻辑
		return LikeVideo(ctx, userId, videoId)
	}

	// 1. 检查视频是否存在（仍需查 DB 获取作者信息用于通知）
	video, err := query.Video.WithContext(ctx).
		Where(query.Video.ID.Eq(videoId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeVideoNotExist, err
		}
		return nil, api.CodeInternalError, err
	}

	// 2. Redis Set 检查是否已点赞（替代 COUNT 查询，天然幂等）
	userLikedKey := fmt.Sprintf("user:%d:liked_videos", userId)
	videoIdStr := fmt.Sprintf("%d", videoId)
	alreadyLiked, err := rdb.SIsMember(ctx, userLikedKey, videoIdStr).Result()
	if err != nil {
		// Redis 查询失败降级到 DB
		return LikeVideo(ctx, userId, videoId)
	}
	if alreadyLiked {
		// 已点赞，幂等返回（计数从 Redis 读）
		count, _ := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "like_count").Int64()
		if count == 0 {
			count = int64(video.LikeCount) // Redis 无值时用 DB 兜底
		}
		return &v1.LikeVideoResp{IsLiked: true, LikeCount: int32(count)}, api.CodeSuccess, nil
	}

	// 3. 先用 Redis SADD 做分布式锁（返回 1=成功添加，0=已存在）
	// 只有成功添加的请求才能继续，避免并发重复创建 likes 记录
	added, err := rdb.SAdd(ctx, userLikedKey, videoIdStr).Result()
	if err != nil {
		// Redis 写入失败降级到 DB
		return LikeVideo(ctx, userId, videoId)
	}
	if added == 0 {
		// 并发场景下另一个请求已添加，幂等返回
		count, err := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "like_count").Int64()
		if err != nil {
			// Redis 没有这个 key/field，用 DB 兜底
			count = int64(video.LikeCount)
		}
		return &v1.LikeVideoResp{IsLiked: true, LikeCount: int32(count)}, api.CodeSuccess, nil
	}

	// 4. SADD 成功，持有分布式锁，执行事务
	like := &model.Like{
		UserID:     userId,
		TargetType: 1,
		TargetID:   videoId,
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 4.1 创建点赞记录（DB，保留持久化）
		if err := tx.Like.WithContext(ctx).Create(like); err != nil {
			return err
		}

		// 4.2 Redis Hash 累积点赞计数（不直接写 MySQL）
		videoStatsKey := fmt.Sprintf("video:%d:stats", videoId)

		// 先检查 Hash 是否存在，不存在则用 DB 值初始化
		exists, _ := rdb.Exists(ctx, videoStatsKey).Result()
		if exists == 0 {
			// 首次写入：用 DB 的当前值 + 1
			rdb.HSet(ctx, videoStatsKey, "like_count", video.LikeCount+1)
		} else {
			// Hash 已存在：直接递增
			_, err := rdb.HIncrBy(ctx, videoStatsKey, "like_count", 1).Result()
			if err != nil {
				return fmt.Errorf("redis hincrby failed: %w", err)
			}
		}

		// 4.4 通知作者（暂保留同步，阶段 3 改 MQ）
		if video.UserID != userId {
			notifSvc.CreateNotification(ctx, tx, &model.Notification{
				UserID:     video.UserID,
				ActorID:    userId,
				ActionType: 2,
				TargetType: 2,
				TargetID:   videoId,
			})
		}

		return nil
	})
	if err != nil {
		// 事务失败回滚：删除 Redis Set 中的点赞状态（释放分布式锁）
		rdb.SRem(ctx, userLikedKey, videoIdStr)
		return nil, api.CodeInternalError, err
	}

	// 5. 返回响应（计数从 Redis 读）
	count, err := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "like_count").Int64()
	if err != nil {
		// Redis 没有这个 key/field，用 DB 兜底
		count = int64(video.LikeCount + 1)
	}
	return &v1.LikeVideoResp{
		IsLiked:   true,
		LikeCount: int32(count),
	}, api.CodeSuccess, nil
}

// UnlikeVideo1 取消点赞（Redis 优化版）
func UnlikeVideo1(ctx context.Context, userId, videoId int64) (*v1.LikeVideoResp, api.ResCode, error) { 
	if rdb == nil {
		// 降级：Redis 不可用时走原逻辑
		return UnlikeVideo(ctx, userId, videoId)
	}

	// 1. 视频存在性检查
	video, err := query.Video.WithContext(ctx).Where(query.Video.ID.Eq(videoId)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeVideoNotExist, err
		}
		return nil, api.CodeInternalError, err
	}

	userLikedKey := fmt.Sprintf("user:%d:liked_videos", userId)
	videoIdStr := strconv.FormatUint(uint64(videoId), 10)

	// 2. 先用 Redis SREM 做分布式锁（返回 1=成功删除，0=不存在）
	removed, err := rdb.SRem(ctx, userLikedKey, videoIdStr).Result()
	if err != nil {
		// Redis 故障降级到原逻辑
		return UnlikeVideo(ctx, userId, videoId)
	}
	if removed == 0 {
		// 并发场景下另一个请求已删除，或本来就没点赞，幂等返回
		count, err := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "like_count").Int64()
		if err != nil {
			// Redis 没有这个 key/field，用 DB 兜底
			count = int64(video.LikeCount)
		}
		return &v1.LikeVideoResp{IsLiked: false, LikeCount: int32(count)}, api.CodeSuccess, nil
	}

	// 3. SREM 成功，持有分布式锁，执行事务
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 3.1 删除点赞记录（物理删除，Like 表无 soft-delete）
		_, err := tx.Like.WithContext(ctx).Where(
			query.Like.UserID.Eq(userId),
			query.Like.TargetType.Eq(1),
			query.Like.TargetID.Eq(videoId),
		).Delete()
		if err != nil {
			return err
		}

		// 3.2 Redis Hash 计数器递减
		videoStatsKey := fmt.Sprintf("video:%d:stats", videoId)

		// 先检查 Hash 是否存在
		exists, _ := rdb.Exists(ctx, videoStatsKey).Result()
		if exists == 0 {
			// 首次写入（取消点赞时 Hash 不存在）：用 DB 的当前值 - 1
			if video.LikeCount > 0 {
				rdb.HSet(ctx, videoStatsKey, "like_count", video.LikeCount-1)
			} else {
				rdb.HSet(ctx, videoStatsKey, "like_count", 0)
			}
		} else {
			// Hash 已存在：直接递减
			newCount, err := rdb.HIncrBy(ctx, videoStatsKey, "like_count", -1).Result()
			if err != nil {
				return fmt.Errorf("redis hincrby failed: %w", err)
			}
			// 防止递减到负数
			if newCount < 0 {
				rdb.HSet(ctx, videoStatsKey, "like_count", 0)
			}
		}

		return nil
	})

	if err != nil {
		// 事务失败回滚：重新添加回 Redis Set（释放分布式锁）
		rdb.SAdd(ctx, userLikedKey, videoIdStr)
		return nil, api.CodeInternalError, err
	}

	// 4. 读取最终计数返回
	finalCount := video.LikeCount
	if count, err := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "like_count").Int64(); err == nil {
		finalCount = int32(count)
	}

	return &v1.LikeVideoResp{IsLiked: false, LikeCount: int32(finalCount)}, api.CodeSuccess, nil
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
	// 4. 异步更新 Redis 热度（重新计算带时间衰减的分数）
	go func(vid int64) {
		hotrank.UpdateVideoHotScore(context.Background(), vid)
	}(videoId)
	// 5. 返回结果
	return &v1.FavoriteVideoResp{
		IsFavorited:   true,
		FavoriteCount: video.FavoriteCount + 1,
	}, api.CodeSuccess, nil
}

// FavoriteVideo1 收藏视频 Redis 优化版
func FavoriteVideo1(ctx context.Context, userId int64, videoId int64) (*v1.FavoriteVideoResp, api.ResCode, error) {
	// 降级检查
	if rdb == nil {
		return FavoriteVideo(ctx, userId, videoId)
	}

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

	userFavoritedKey := fmt.Sprintf("user:%d:favorited_videos", userId)
	videoIdStr := fmt.Sprintf("%d", videoId)

	// 2. 先用 Redis SADD 做分布式锁（返回 1=成功添加，0=已存在）
	added, err := rdb.SAdd(ctx, userFavoritedKey, videoIdStr).Result()
	if err != nil {
		// Redis 故障降级到原逻辑
		return FavoriteVideo(ctx, userId, videoId)
	}
	if added == 0 {
		// 并发场景下另一个请求已添加，幂等返回
		favoriteCount, err := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "favorite_count").Int64()
		if err != nil {
			// Redis 没有这个 key/field，用 DB 兜底
			favoriteCount = int64(video.FavoriteCount)
		}
		return &v1.FavoriteVideoResp{IsFavorited: true, FavoriteCount: int32(favoriteCount)}, api.CodeSuccess, nil
	}

	// 3. SADD 成功，持有分布式锁，执行事务
	favorite := &model.Favorite{
		UserID:  userId,
		VideoID: videoId,
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 3.1 创建收藏记录
		if err := tx.Favorite.WithContext(ctx).Create(favorite); err != nil {
			return err
		}

		// 3.2 Redis Hash 累积收藏计数
		videoStatsKey := fmt.Sprintf("video:%d:stats", videoId)
		exists, _ := rdb.Exists(ctx, videoStatsKey).Result()
		if exists == 0 {
			// 首次写入：用 DB 的当前值 + 1
			rdb.HSet(ctx, videoStatsKey, "favorite_count", video.FavoriteCount+1)
		} else {
			// Hash 已存在：直接递增
			_, err := rdb.HIncrBy(ctx, videoStatsKey, "favorite_count", 1).Result()
			if err != nil {
				return fmt.Errorf("redis hincrby failed: %w", err)
			}
		}

		// 3.3 通知视频作者（不通知自己）
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
		// 事务失败回滚：删除 Redis Set 中的收藏状态（释放分布式锁）
		rdb.SRem(ctx, userFavoritedKey, videoIdStr)
		return nil, api.CodeInternalError, err
	}

	// 4. 读取最终计数返回
	favoriteCount, err := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "favorite_count").Int64()
	if err != nil {
		// Redis 没有这个 key/field，用 DB 兜底
		favoriteCount = int64(video.FavoriteCount + 1)
	}
	return &v1.FavoriteVideoResp{
		IsFavorited:   true,
		FavoriteCount: int32(favoriteCount),
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
	// 4. 异步更新 Redis 热度（重新计算带时间衰减的分数）
	go func(vid int64) {
		hotrank.UpdateVideoHotScore(context.Background(), vid)
	}(videoId)
	// 5. 返回结果
	return &v1.FavoriteVideoResp{
		IsFavorited:   false,
		FavoriteCount: video.FavoriteCount - 1,
	}, api.CodeSuccess, nil
}

// UnfavoriteVideo1 取消收藏视频 Redis 优化版
func UnfavoriteVideo1(ctx context.Context, userId int64, videoId int64) (*v1.FavoriteVideoResp, api.ResCode, error) {
	// 1. 视频存在性检查
	video, err := query.Video.WithContext(ctx).Where(query.Video.ID.Eq(videoId)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.CodeVideoNotExist, err
		}
		return nil, api.CodeInternalError, err
	}

	userFavoritedKey := fmt.Sprintf("user:%d:favorited_videos", userId)
	videoIdStr := fmt.Sprintf("%d", videoId)

	// 2. 先用 Redis SREM 做分布式锁（返回 1=成功删除，0=不存在）
	removed, err := rdb.SRem(ctx, userFavoritedKey, videoIdStr).Result()
	if err != nil {
		// Redis 故障降级到原逻辑
		return UnfavoriteVideo(ctx, userId, videoId)
	}
	if removed == 0 {
		// 并发场景下另一个请求已删除，或本来就没收藏，幂等返回
		favoriteCount, err := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "favorite_count").Int64()
		if err != nil {
			// Redis 没有这个 key/field，用 DB 兜底
			favoriteCount = int64(video.FavoriteCount)
		}
		return &v1.FavoriteVideoResp{IsFavorited: false, FavoriteCount: int32(favoriteCount)}, api.CodeSuccess, nil
	}

	// 3. SREM 成功，持有分布式锁，执行事务
	err = query.Q.Transaction(func(tx *query.Query) error {
		// 3.1 删除收藏记录（物理删除）
		if _, err := tx.Favorite.WithContext(ctx).
			Where(tx.Favorite.UserID.Eq(userId), tx.Favorite.VideoID.Eq(videoId)).
			Delete(); err != nil {
			return err
		}

		// 3.2 Redis Hash 计数器递减
		videoStatsKey := fmt.Sprintf("video:%d:stats", videoId)
		exists, _ := rdb.Exists(ctx, videoStatsKey).Result()
		if exists == 0 {
			// 首次写入：用 DB 的当前值 - 1
			if video.FavoriteCount > 0 {
				rdb.HSet(ctx, videoStatsKey, "favorite_count", video.FavoriteCount-1)
			} else {
				rdb.HSet(ctx, videoStatsKey, "favorite_count", 0)
			}
		} else {
			// Hash 已存在：直接递减
			newCount, err := rdb.HIncrBy(ctx, videoStatsKey, "favorite_count", -1).Result()
			if err != nil {
				return fmt.Errorf("redis hincrby failed: %w", err)
			}
			// 防止递减到负数
			if newCount < 0 {
				rdb.HSet(ctx, videoStatsKey, "favorite_count", 0)
			}
		}
		return nil
	})

	if err != nil {
		// 事务失败回滚：重新添加回 Redis Set（释放分布式锁）
		rdb.SAdd(ctx, userFavoritedKey, videoIdStr)
		return nil, api.CodeInternalError, err
	}
	// 4. 返回结果
	// 如果 Redis Hash 中没有值，则用 DB 的值兜底
	favoriteCount := video.FavoriteCount
	if count, err := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "favorite_count").Int64(); err == nil {
		favoriteCount = int32(count)
	}
	return &v1.FavoriteVideoResp{
		IsFavorited:   false,
		FavoriteCount: int32(favoriteCount),
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