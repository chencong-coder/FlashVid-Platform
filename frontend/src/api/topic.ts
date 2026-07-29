import http from './http'
import type { ApiResponse } from '@/types/api'
import type { FeedVideo } from './feed'

export interface TopicItem {
  id: number
  name: string
  description: string
  coverUrl: string
  viewCount: number
  videoCount: number
  createdAt: string
}

export interface TopicsResponse {
  topics: TopicItem[]
  nextCursorToken: string
  hasMore: boolean
}

export interface TopicVideosResponse {
  videos: FeedVideo[]
  nextCursorToken: string
  hasMore: boolean
}

export const getTopics = (params: { sort?: string; count?: number; cursor?: string } = {}) =>
  http.get<ApiResponse<TopicsResponse>>('/topics', { params })

export const searchTopics = (keyword: string, cursor?: string, count?: number) =>
  http.get<ApiResponse<TopicsResponse>>('/topics/search', { params: { keyword, cursor, count } })

// 话题详情
export const getTopicById = (topicId: string | number) =>
  http.get<ApiResponse<{ topic: TopicItem }>>(`/topics/${topicId}`)

// 话题下的视频列表（popular / latest）
export const getTopicVideos = (
  topicId: string | number,
  params: { sort?: 'popular' | 'latest'; cursor?: string; count?: number } = {},
) => http.get<ApiResponse<TopicVideosResponse>>(`/topics/${topicId}/videos`, { params })
