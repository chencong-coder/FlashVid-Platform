import http from './http'
import type { ApiResponse } from '@/types/api'
import type { FeedVideo } from './feed'

export interface TopicItem {
  id: string
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

export type TopicSort = 'hot' | 'latest'

export interface GetTopicsParams {
  sort?: TopicSort
  count?: number
  cursor?: string
}

export interface GetTopicVideosParams {
  sort?: 'popular' | 'latest'
  cursor?: string
  count?: number
}

export const getTopics = (params: GetTopicsParams = {}) =>
  http.get<ApiResponse<TopicsResponse>>('/topics', { params })

export const searchTopics = (keyword: string, cursor?: string, count?: number) =>
  http.get<ApiResponse<TopicsResponse>>('/topics/search', { params: { keyword, cursor, count } })

// 话题详情
export const getTopicById = (topicId: string) =>
  http.get<ApiResponse<{ topic: TopicItem }>>(`/topics/${topicId}`)

// 话题下的视频列表（popular / latest）
export const getTopicVideos = (topicId: string, params: GetTopicVideosParams = {}) =>
  http.get<ApiResponse<TopicVideosResponse>>(`/topics/${topicId}/videos`, { params })
