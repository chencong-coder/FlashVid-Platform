import http from './http'

import type { ApiResponse } from '@/types/api'
import type { CommentItem } from '@/types/video'

// ===== 评论接口 =====
export interface CommentCursorParams {
  cursor?: string
  count?: number
}

export interface CommentPageResp {
  comments: CommentItem[]
  nextCursorToken: string
  hasMore: boolean
}

export interface LikeCommentResp {
  isLiked: boolean
  likeCount: number
}

export const getVideoComments = (videoId: string, params?: CommentCursorParams) =>
  http.get<ApiResponse<CommentPageResp>>(`/videos/${videoId}/comments`, { params })

export const postComment = (videoId: string, content: string) =>
  http.post<ApiResponse<{ comment: CommentItem }>>(`/videos/${videoId}/comments`, {
    content,
    parentId: 0,
  })

export const likeComment = (commentId: string) =>
  http.post<ApiResponse<LikeCommentResp>>(`/comments/${commentId}/like`)

export const unlikeComment = (commentId: string) =>
  http.delete<ApiResponse<LikeCommentResp>>(`/comments/${commentId}/like`)

// ===== 互动接口 =====
export interface LikeVideoResp {
  isLiked: boolean
  likeCount: number
}

export interface FavoriteVideoResp {
  isFavorited: boolean
  favoriteCount: number
}

export interface ShareVideoResp {
  shareUrl: string
  shareCount: number
}

export type SharePlatform = 'wechat' | 'qq' | 'weibo' | 'link'

export const likeVideo = (videoId: string) =>
  http.post<ApiResponse<LikeVideoResp>>(`/videos/${videoId}/like`)

export const unlikeVideo = (videoId: string) =>
  http.delete<ApiResponse<LikeVideoResp>>(`/videos/${videoId}/like`)

export const favoriteVideo = (videoId: string) =>
  http.post<ApiResponse<FavoriteVideoResp>>(`/videos/${videoId}/favorite`)

export const unfavoriteVideo = (videoId: string) =>
  http.delete<ApiResponse<FavoriteVideoResp>>(`/videos/${videoId}/favorite`)

export const shareVideo = (videoId: string, platform: SharePlatform = 'link') =>
  http.post<ApiResponse<ShareVideoResp>>(`/videos/${videoId}/share`, { platform })
