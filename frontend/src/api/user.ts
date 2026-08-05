import http from './http'

import type { ApiResponse } from '@/types/api'
import type { FeedVideo } from './feed'

// 登录 / 注册 / 刷新 Token 见 @/api/auth

// ===== 用户信息 =====
export interface UserInfo {
  userId: string
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
  isFollowing?: boolean  // 当前登录用户是否已关注（查看他人主页时由后端返回）
}

export interface UpdateUserResult extends Omit<UserInfo, 'createdAt'> {
  updatedAt: string
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
export const getUserInfo = (userId: string) => http.get<ApiResponse<UserInfo>>(`/user/${userId}`)

export const updateUserInfo = (payload: UpdateUserPayload) =>
  http.put<ApiResponse<UpdateUserResult>>('/user/profile', payload)

export const getUserVideos = (userId: string, params: PageParams = {}) =>
  http.get<ApiResponse<VideoListResp>>(`/user/${userId}/videos`, { params })

// 我的点赞列表（私有）
export const getMyLikes = (params: PageParams = {}) =>
  http.get<ApiResponse<VideoListResp>>('/user/profile/likes', { params })

// 我的收藏列表（私有）
export const getMyFavorites = (params: PageParams = {}) =>
  http.get<ApiResponse<VideoListResp>>('/user/profile/favorites', { params })

// ===== 播放列表 =====
export interface PlaylistInfo {
  id: string
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
  nextCursor: string
  hasMore: boolean
}

export interface PlaylistVideosParams {
  cursor?: string
  count?: number
}

export interface CreatePlaylistPayload {
  title: string
  description?: string
  coverUrl?: string
}

export interface UpdatePlaylistPayload {
  title?: string
  description?: string
  coverUrl?: string
}

interface PlaylistWire {
  id: string
  title: string
  description: string
  cover_url: string
  video_count: number
  created_at: string
}

interface PlaylistWriteWire {
  title?: string
  description?: string
  cover_url?: string
}

interface GetPlaylistsWireResp {
  playlists: PlaylistWire[] | null
}

interface PlaylistVideosWireResp {
  playlist: PlaylistWire
  videos: FeedVideo[]
  nextCursor: string
  hasMore: boolean
}

const mapPlaylist = (playlist: PlaylistWire): PlaylistInfo => ({
  id: playlist.id,
  title: playlist.title,
  description: playlist.description,
  coverUrl: playlist.cover_url,
  videoCount: playlist.video_count,
  createdAt: playlist.created_at,
})

const mapPlaylistPayload = (payload: UpdatePlaylistPayload): PlaylistWriteWire => ({
  ...(payload.title !== undefined ? { title: payload.title } : {}),
  ...(payload.description !== undefined ? { description: payload.description } : {}),
  ...(payload.coverUrl !== undefined ? { cover_url: payload.coverUrl } : {}),
})

export const getMyPlaylists = async () => {
  const response = await http.get<ApiResponse<GetPlaylistsWireResp>>('/playlists')
  return {
    ...response,
    data: {
      ...response.data,
      data: {
        playlists: (response.data.data.playlists ?? []).map(mapPlaylist),
      },
    },
  }
}

export const createPlaylist = async (payload: CreatePlaylistPayload) => {
  const response = await http.post<ApiResponse<{ playlist: PlaylistWire }>>('/playlists', {
    ...mapPlaylistPayload(payload),
    title: payload.title,
  })
  return {
    ...response,
    data: {
      ...response.data,
      data: { playlist: mapPlaylist(response.data.data.playlist) },
    },
  }
}

export const updatePlaylist = (playlistId: string, payload: UpdatePlaylistPayload) =>
  http.put<ApiResponse<string>>(`/playlists/${playlistId}`, mapPlaylistPayload(payload))

export const deletePlaylist = (playlistId: string) =>
  http.delete<ApiResponse<string>>(`/playlists/${playlistId}`)

export const getPlaylistVideos = async (playlistId: string, params: PlaylistVideosParams = {}) => {
  const response = await http.get<ApiResponse<PlaylistVideosWireResp>>(
    `/playlists/${playlistId}/videos`,
    { params },
  )
  return {
    ...response,
    data: {
      ...response.data,
      data: {
        ...response.data.data,
        playlist: mapPlaylist(response.data.data.playlist),
      },
    },
  }
}

export const addVideoToPlaylist = (playlistId: string, videoId: string) =>
  http.post<ApiResponse<string>>(`/playlists/${playlistId}/videos`, { videoId })

export const removeVideoFromPlaylist = (playlistId: string, videoId: string) =>
  http.delete<ApiResponse<string>>(`/playlists/${playlistId}/videos/${videoId}`)

// ===== 推荐用户 =====
export interface RecommendUser {
  userId: string
  nickname: string
  avatar: string
  bio: string
  followerCount: number
}

export const getRecommendUsers = (count = 5) =>
  http.get<ApiResponse<{ users: RecommendUser[] }>>('/user/recommend', { params: { count } })

// ===== 搜索用户 =====
export interface SearchUserItem {
  userId: string
  nickname: string
  avatar: string
  bio: string
  followerCount: number
}

export const searchUsers = (keyword: string, count = 6) =>
  http.get<ApiResponse<{ users: SearchUserItem[] }>>('/users/search', { params: { keyword, count } })

// ===== 关注 =====
export const followUser = (userId: string) =>
  http.post<ApiResponse<FollowResp>>(`/user/${userId}/follow`)

export const unfollowUser = (userId: string) =>
  http.delete<ApiResponse<FollowResp>>(`/user/${userId}/follow`)

export const getUserFollowers = (userId: string, params: PageParams = {}) =>
  http.get<ApiResponse<{ followers: UserInfo[]; pagination: Pagination }>>(
    `/user/${userId}/followers`,
    { params },
  )

export const getUserFollowing = (userId: string, params: PageParams = {}) =>
  http.get<ApiResponse<{ following: UserInfo[]; pagination: Pagination }>>(
    `/user/${userId}/followings`,
    { params },
  )
