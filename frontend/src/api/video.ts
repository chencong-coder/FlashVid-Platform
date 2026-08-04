import http from './http'

import type { ApiResponse } from '@/types/api'
import type { CommentItem, ReplyItem } from '@/types/video'
import type { FeedVideo } from './feed'
import type { Pagination } from './user'

// ===== 视频接口 =====
export interface CreateVideoPayload {
  title: string // 必填，最长 100
  coverUrl: string // 必填，已上传的封面 URL
  videoUrl: string // 必填，已上传的视频 URL
  duration: number // 必填，秒，最小 1
  description?: string // 最长 500
  width?: number
  height?: number
  musicId?: string // 背景音乐 ID
  city?: string
  location?: string
  latitude?: number // [-90, 90]
  longitude?: number // (-180, 180]
  topicNames?: string[] // 最多 5 个，每个 1-20 字符
}

export interface CreateVideoResp {
  video_id: string // 后端返回 snake_case
  status: number // 1-审核中 2-成功 3-未通过 4-下架
}

export interface DeleteVideoResp {
  message: string
}

export interface SearchVideosParams {
  keyword: string
  page?: number
  pageSize?: number // 最大 50
}

export interface SearchVideosResp {
  videos: FeedVideo[]
  pagination: Pagination
}

// 发布视频（需登录）
export const createVideo = (payload: CreateVideoPayload) =>
  http.post<ApiResponse<CreateVideoResp>>('/videos', payload)

// 删除视频（需登录）
export const deleteVideo = (videoId: string) =>
  http.delete<ApiResponse<DeleteVideoResp>>(`/videos/${videoId}`)

// 获取视频详情（公开）
export const getVideoDetail = (videoId: string) =>
  http.get<ApiResponse<{ video: FeedVideo }>>(`/videos/${videoId}`)

// 搜索视频（公开）
export const searchVideos = (params: SearchVideosParams) =>
  http.get<ApiResponse<SearchVideosResp>>('/videos/search', { params })

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

// 发表评论 / 回复的返回：一级评论返回 comment，回复返回 reply
export interface CreateCommentResp {
  comment?: CommentItem
  reply?: ReplyItem
}

export interface LikeCommentResp {
  isLiked: boolean
  likeCount: number
}

export interface DeleteCommentResp {
  [key: string]: never
}

// 获取视频评论列表（公开，游标分页）
export const getVideoComments = (videoId: string, params?: CommentCursorParams) =>
  http.get<ApiResponse<CommentPageResp>>(`/videos/${videoId}/comments`, { params })

// 获取某条评论的回复列表（公开）
export const getCommentReplies = (commentId: string) =>
  http.get<ApiResponse<{ replies: ReplyItem[] }>>(`/comments/${commentId}/replies`)

// 发表评论 / 回复（需登录）；parentId=0 为一级评论，>0 为回复
export const postComment = (
  videoId: string,
  content: string,
  parentId = '0',
  replyToUserId = '0',
) =>
  http.post<ApiResponse<CreateCommentResp>>(`/videos/${videoId}/comments`, {
    content,
    parentId,
    replyToUserId,
  })

// 删除评论（需登录）
export const deleteComment = (commentId: string) =>
  http.delete<ApiResponse<DeleteCommentResp>>(`/comments/${commentId}`)

// 点赞 / 取消点赞评论（需登录）
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
