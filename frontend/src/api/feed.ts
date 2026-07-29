import http from './http'

import type { ApiResponse } from '@/types/api'
import type { FeedType, VideoItem } from '@/types/video'

// ===== 后端返回结构 =====
export interface FeedAuthor {
  id: number
  username: string
  avatar: string
  nickname: string
}

export interface FeedStats {
  viewCount: number
  likeCount: number
  commentCount: number
  shareCount: number
  favoriteCount: number
}

export interface FeedVideo {
  id: number
  title: string
  description: string
  coverUrl: string
  videoUrl: string
  duration: number
  width: number
  height: number
  musicId: number
  city: string
  topics: string[]
  author: FeedAuthor
  stats: FeedStats
  publishedAt: string
}

export interface FeedResponse {
  videos: FeedVideo[]
  nextCursorToken: string
  hasMore: boolean
}

export interface FeedParams {
  cursor?: string
  count?: number
}

export interface NearbyParams extends FeedParams {
  latitude: number
  longitude: number
  distance?: number
}

// ===== 接口调用 =====
export const getRecommendFeed = (params: FeedParams = {}) =>
  http.get<ApiResponse<FeedResponse>>('/feed/recommend', { params })

export const getFollowFeed = (params: FeedParams = {}) =>
  http.get<ApiResponse<FeedResponse>>('/feed/follow', { params })

export const getFriendsFeed = (params: FeedParams = {}) =>
  http.get<ApiResponse<FeedResponse>>('/feed/friends', { params })

export const getNearbyFeed = (params: NearbyParams) =>
  http.get<ApiResponse<FeedResponse>>('/feed/nearby', { params })

// ===== 后端 FeedVideo → 前端 VideoItem =====
export const mapFeedVideo = (v: FeedVideo, feed: FeedType): VideoItem => ({
  id: String(v.id),
  author: {
    id: String(v.author.id),
    nickname: v.author.nickname || v.author.username,
    avatar: v.author.avatar,
    // 关注流 / 好友流里的作者一定是已关注的
    followed: feed === 'follow' || feed === 'friends',
  },
  description: v.description || v.title,
  topics: v.topics ?? [],
  music: `原声 - ${v.author.nickname || v.author.username}`,
  poster: v.coverUrl,
  source: v.videoUrl,
  discCover: v.coverUrl,
  stats: {
    likes: v.stats.likeCount,
    comments: v.stats.commentCount,
    favorites: v.stats.favoriteCount,
    shares: v.stats.shareCount,
  },
  liked: false,
  favorited: false,
  city: feed === 'nearby' ? v.city : undefined,
})
