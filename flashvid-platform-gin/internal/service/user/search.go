package user

import (
	"context"
	"fmt"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
)

// SearchUsers 按昵称或用户名模糊搜索用户
func SearchUsers(ctx context.Context, keyword string, count int) ([]model.UserInfo, api.ResCode, error) {
	pattern := fmt.Sprintf("%%%s%%", keyword)

	users, err := query.User.WithContext(ctx).
		Where(
			query.User.Nickname.Like(pattern),
		).
		Or(query.User.Username.Like(pattern)).
		Order(query.User.FollowerCount.Desc()).
		Limit(count).
		Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	infos := make([]model.UserInfo, 0, len(users))
	for _, u := range users {
		infos = append(infos, model.UserInfo{
			UserId:         u.ID,
			Nickname:       u.Nickname,
			Avatar:         u.Avatar,
			Bio:            u.Bio,
			FollowersCount: u.FollowerCount,
		})
	}
	return infos, api.CodeSuccess, nil
}
