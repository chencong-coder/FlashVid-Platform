import http from './http'

import type { ApiResponse, PageQuery, PageResult } from '@/types/api'
import type { CommentItem, FeedType, VideoItem } from '@/types/video'

export const getVideoFeed = (feed: FeedType, params: PageQuery) =>
  http.get<ApiResponse<PageResult<VideoItem>>>('/videos/feed', { params: { ...params, feed } })

export const getVideoComments = (videoId: string, params: PageQuery) =>
  http.get<ApiResponse<PageResult<CommentItem>>>(`/videos/${videoId}/comments`, { params })

export const likeVideo = (videoId: string, liked: boolean) =>
  http.post<ApiResponse<null>>(`/videos/${videoId}/like`, { liked })
