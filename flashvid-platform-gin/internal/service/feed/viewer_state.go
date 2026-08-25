package feed

import (
	"context"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
	"fmt"
	"strconv"
)

// AttachViewerState 批量填充当前登录用户对视频的互动状态：
// 是否点赞（isLiked）、是否收藏（isFavorited）、是否已关注作者（isFollowing）。
// viewerID <= 0 表示未登录，直接返回（所有状态保持 false）。
// 导出供 feed 之外的服务（如个人主页的作品/喜欢/收藏列表）复用，作为回填互动状态的唯一实现。
func AttachViewerState(ctx context.Context, viewerID int64, videos []model.VideoInfo) error {
	if viewerID <= 0 || len(videos) == 0 {
		return nil
	}

	// 收集视频ID和作者ID（作者去重）
	videoIDs := make([]int64, 0, len(videos))
	authorIDSet := make(map[int64]struct{}, len(videos))
	for _, v := range videos {
		videoIDs = append(videoIDs, v.ID)
		if v.Author.ID > 0 {
			authorIDSet[v.Author.ID] = struct{}{}
		}
	}
	authorIDs := make([]int64, 0, len(authorIDSet))
	for id := range authorIDSet {
		authorIDs = append(authorIDs, id)
	}

	// Redis 客户端
	rdb := dao.RedisClient

	// 1. 查点赞（优先 Redis，降级 MySQL）
	likedSet := make(map[int64]struct{})
	if rdb != nil {
		// Redis 批量查询：SMISMEMBER user:{viewerID}:liked_videos {videoID1} {videoID2} ...
		userLikedKey := fmt.Sprintf("user:%d:liked_videos", viewerID)
		videoIDStrs := make([]string, len(videoIDs))
		for i, vid := range videoIDs {
			videoIDStrs[i] = strconv.FormatInt(vid, 10)
		}

		// 使用 SMIsMember 批量检查（Redis 6.2+）
		// 如果 Redis 版本 < 6.2，回退到循环 SIsMember
		results, err := rdb.SMIsMember(ctx, userLikedKey, videoIDStrs).Result()
		if err == nil {
			for i, isMember := range results {
				if isMember {
					likedSet[videoIDs[i]] = struct{}{}
				}
			}
		} else {
			// Redis 失败降级到 MySQL
			likedSet = getLikedSetFromDB(ctx, viewerID, videoIDs)
		}
	} else {
		// Redis 不可用，直接查 DB
		likedSet = getLikedSetFromDB(ctx, viewerID, videoIDs)
	}

	// 2. 查收藏（优先 Redis，降级 MySQL）
	favoritedSet := make(map[int64]struct{})
	if rdb != nil {
		userFavoritedKey := fmt.Sprintf("user:%d:favorited_videos", viewerID)
		videoIDStrs := make([]string, len(videoIDs))
		for i, vid := range videoIDs {
			videoIDStrs[i] = strconv.FormatInt(vid, 10)
		}

		results, err := rdb.SMIsMember(ctx, userFavoritedKey, videoIDStrs).Result()
		if err == nil {
			for i, isMember := range results {
				if isMember {
					favoritedSet[videoIDs[i]] = struct{}{}
				}
			}
		} else {
			// Redis 失败降级到 MySQL
			favoritedSet = getFavoritedSetFromDB(ctx, viewerID, videoIDs)
		}
	} else {
		// Redis 不可用，直接查 DB
		favoritedSet = getFavoritedSetFromDB(ctx, viewerID, videoIDs)
	}

	// 3. 查关注（viewer 关注了哪些作者）
	followingSet := make(map[int64]struct{})
	if len(authorIDs) > 0 {
		follows, err := query.Follow.WithContext(ctx).
			Where(
				query.Follow.FollowerID.Eq(viewerID),
				query.Follow.FollowingID.In(authorIDs...),
			).
			Find()
		if err != nil {
			return err
		}
		for _, fl := range follows {
			followingSet[fl.FollowingID] = struct{}{}
		}
	}

	// 4. 回填状态
	for i := range videos {
		if _, ok := likedSet[videos[i].ID]; ok {
			videos[i].IsLiked = true
		}
		if _, ok := favoritedSet[videos[i].ID]; ok {
			videos[i].IsFavorited = true
		}
		if _, ok := followingSet[videos[i].Author.ID]; ok {
			videos[i].IsFollowing = true
		}
	}
	return nil
}

// getLikedSetFromDB 从 MySQL 查询点赞状态（Redis 降级逻辑）
func getLikedSetFromDB(ctx context.Context, viewerID int64, videoIDs []int64) map[int64]struct{} {
	likedSet := make(map[int64]struct{})
	likes, err := query.Like.WithContext(ctx).
		Where(
			query.Like.UserID.Eq(viewerID),
			query.Like.TargetType.Eq(1),
			query.Like.TargetID.In(videoIDs...),
		).
		Find()
	if err != nil {
		return likedSet
	}
	for _, l := range likes {
		likedSet[l.TargetID] = struct{}{}
	}
	return likedSet
}

// getFavoritedSetFromDB 从 MySQL 查询收藏状态（Redis 降级逻辑）
func getFavoritedSetFromDB(ctx context.Context, viewerID int64, videoIDs []int64) map[int64]struct{} {
	favoritedSet := make(map[int64]struct{})
	favorites, err := query.Favorite.WithContext(ctx).
		Where(
			query.Favorite.UserID.Eq(viewerID),
			query.Favorite.VideoID.In(videoIDs...),
		).
		Find()
	if err != nil {
		return favoritedSet
	}
	for _, f := range favorites {
		favoritedSet[f.VideoID] = struct{}{}
	}
	return favoritedSet
}
