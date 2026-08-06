import type { ApiResponse } from './types'
import http from './axios'

export interface NotificationItem {
  id: number
  actorId: number
  actorName: string
  actorAvatar: string
  actionType: number // 1=关注 2=点赞 3=收藏 4=评论 5=回复
  targetId: number
  targetTitle?: string
  targetCover?: string
  content?: string
  isRead: number
  createdAt: string
}

export interface UnreadCounts {
  followers: number
  likesAndFavs: number
  mentions: number
  comments: number
}

export const getNotifications = (params: {
  action_types?: number[]
  page: number
  page_size: number
}) => http.get<ApiResponse<{ list: NotificationItem[]; total: number }>>('/notifications', { params })

export const getUnreadCounts = () =>
  http.get<ApiResponse<UnreadCounts>>('/notifications/unread-counts')

export const markAsRead = (actionTypes?: number[]) =>
  http.put<ApiResponse<null>>('/notifications/read', { actionTypes: actionTypes ?? [] })
