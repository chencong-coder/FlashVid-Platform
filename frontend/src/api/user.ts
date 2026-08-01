import http from './http'

import type { ApiResponse } from '@/types/api'
import type { FeedVideo } from './feed'

// 登录 / 注册 / 刷新 Token 见 @/api/auth

// ===== 用户信息 =====
export interface UserInfo {
  userId: number
  username: string
  nickname: string
  avatar: string
  bio: string
  city: string
  followersCount: number
  followingCount: number
  videosCount: number
  likesCount: number
  phone: string
  gender: number
  birthday: string
  email: string
  createdAt: string
}

export interface Pagination {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

export interface VideoListResp {
  videos: FeedVideo[]
  pagination: Pagination
}

export interface UpdateUserPayload {
  nickname?: string
  avatar?: string
  bio?: string
  city?: string
  gender?: number
  birthday?: string
  email?: string
  phone?: string
}

export interface FollowResp {
  isFollowing: boolean
}

export interface PageParams {
  page?: number
  pageSize?: number
}

// ===== 用户资料 =====
export const getUserInfo = (userId: string | number) =>
  http.get<ApiResponse<UserInfo>>(`/user/${userId}`)

export const updateUserInfo = (payload: UpdateUserPayload) =>
  http.put<ApiResponse<UserInfo>>('/user/profile', payload)

export const getUserVideos = (userId: string | number, params: PageParams = {}) =>
  http.get<ApiResponse<VideoListResp>>(`/user/${userId}/videos`, { params })

// 我的点赞列表（私有）
export const getMyLikes = (params: PageParams = {}) =>
  http.get<ApiResponse<VideoListResp>>('/user/profile/likes', { params })

// 我的收藏列表（私有）
export const getMyFavorites = (params: PageParams = {}) =>
  http.get<ApiResponse<VideoListResp>>('/user/profile/favorites', { params })

// ===== 播放列表 =====
export interface PlaylistInfo {
  id: number
  title: string
  description: string
  coverUrl: string
  videoCount: number
  createdAt: string
}

export interface GetPlaylistsResp {
  playlists: PlaylistInfo[]
}

export interface PlaylistVideosResp {
  playlist: PlaylistInfo
  videos: FeedVideo[]
  pagination: Pagination
}

export const getMyPlaylists = () =>
  http.get<ApiResponse<GetPlaylistsResp>>('/playlists')

export const createPlaylist = (payload: { title: string; description?: string; coverUrl?: string }) =>
  http.post<ApiResponse<{ playlist: PlaylistInfo }>>('/playlists', payload)

export const updatePlaylist = (
  playlistId: number,
  payload: { title?: string; description?: string; coverUrl?: string },
) => http.put<ApiResponse<unknown>>(`/playlists/${playlistId}`, payload)

export const deletePlaylist = (playlistId: number) =>
  http.delete<ApiResponse<unknown>>(`/playlists/${playlistId}`)

export const getPlaylistVideos = (playlistId: number, params: PageParams = {}) =>
  http.get<ApiResponse<PlaylistVideosResp>>(`/playlists/${playlistId}/videos`, { params })

export const addVideoToPlaylist = (playlistId: number, videoId: number) =>
  http.post<ApiResponse<unknown>>(`/playlists/${playlistId}/videos`, { videoId })

export const removeVideoFromPlaylist = (playlistId: number, videoId: number) =>
  http.delete<ApiResponse<unknown>>(`/playlists/${playlistId}/videos/${videoId}`)

// ===== 关注 =====
export const followUser = (userId: string | number) =>
  http.post<ApiResponse<FollowResp>>(`/user/${userId}/follow`)

export const unfollowUser = (userId: string | number) =>
  http.delete<ApiResponse<FollowResp>>(`/user/${userId}/follow`)

export const getUserFollowers = (userId: string | number, params: PageParams = {}) =>
  http.get<ApiResponse<{ followers: UserInfo[]; pagination: Pagination }>>(
    `/user/${userId}/followers`,
    { params },
  )

export const getUserFollowing = (userId: string | number, params: PageParams = {}) =>
  http.get<ApiResponse<{ following: UserInfo[]; pagination: Pagination }>>(
    `/user/${userId}/followings`,
    { params },
  )
