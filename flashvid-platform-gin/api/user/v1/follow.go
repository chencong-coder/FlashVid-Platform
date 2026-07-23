package v1

import (
    "flashvid-platform-gin/internal/model"
)

// 关注响应
type FollowResp struct {
    IsFollowing bool `json:"isFollowing"` // 是否已关注
}



// 获取用户粉丝/关注列表请求
type GetUserFollowReq struct {
    Page     int `form:"page" json:"page"`         // 页码
    PageSize int `form:"pageSize" json:"pageSize"` // 每页数量
}

// 获取用户粉丝列表响应
type GetUserFollowersResp struct {
    Followers  []model.UserInfo `json:"followers"`  // 粉丝列表
    Pagination model.Pagination `json:"pagination"` // 分页信息
}


// 获取用户关注列表响应
type GetUserFollowingResp struct {
    Followings  []model.UserInfo `json:"following"`  // 关注列表
    Pagination model.Pagination `json:"pagination"` // 分页信息
}
