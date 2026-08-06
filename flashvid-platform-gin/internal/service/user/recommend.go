package user

import (
	"context"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
)

// GetRecommendUsers 获取推荐用户列表（按粉丝数降序，排除自己和已关注的用户）
func GetRecommendUsers(ctx context.Context, loginUserID int64, count int) ([]model.UserInfo, api.ResCode, error) {
	q := query.User.WithContext(ctx).
		Order(query.User.FollowerCount.Desc()).
		Limit(count)

	if loginUserID > 0 {
		followings, err := query.Follow.WithContext(ctx).
			Where(query.Follow.FollowerID.Eq(loginUserID)).
			Find()
		if err != nil {
			return nil, api.CodeInternalError, err
		}
		excludeIDs := make([]int64, 0, len(followings)+1)
		excludeIDs = append(excludeIDs, loginUserID) // 排除自己
		for _, f := range followings {
			excludeIDs = append(excludeIDs, f.FollowingID) // 排除已关注
		}
		q = q.Where(query.User.ID.NotIn(excludeIDs...))
	}

	users, err := q.Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	userInfos := make([]model.UserInfo, 0, len(users))
	for _, u := range users {
		userInfos = append(userInfos, model.UserInfo{
			UserId:         u.ID,
			Username:       u.Username,
			Gender:         u.Gender,
			Nickname:       u.Nickname,
			Avatar:         u.Avatar,
			Bio:            u.Bio,
			City:           u.City,
			FollowersCount: u.FollowerCount,
			FollowingCount: u.FollowingCount,
			VideosCount:    u.VideoCount,
			LikesCount:     u.LikeCount,
			Phone:          u.Phone,
			Birthday:       u.Birthday.Format("2006-01-02"),
			Email:          u.Email,
			CreatedAt:      u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return userInfos, api.CodeSuccess, nil
}
