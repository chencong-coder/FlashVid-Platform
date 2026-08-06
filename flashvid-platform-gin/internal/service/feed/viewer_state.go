package feed

import (
	"context"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
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

	// 1. 查点赞（target_type=1 表示视频）
	likedSet := make(map[int64]struct{})
	likes, err := query.Like.WithContext(ctx).
		Where(
			query.Like.UserID.Eq(viewerID),
			query.Like.TargetType.Eq(1),
			query.Like.TargetID.In(videoIDs...),
		).
		Find()
	if err != nil {
		return err
	}
	for _, l := range likes {
		likedSet[l.TargetID] = struct{}{}
	}

	// 2. 查收藏
	favoritedSet := make(map[int64]struct{})
	favorites, err := query.Favorite.WithContext(ctx).
		Where(
			query.Favorite.UserID.Eq(viewerID),
			query.Favorite.VideoID.In(videoIDs...),
		).
		Find()
	if err != nil {
		return err
	}
	for _, f := range favorites {
		favoritedSet[f.VideoID] = struct{}{}
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
