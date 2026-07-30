import http from './http'

import type { ApiResponse } from '@/types/api'

// ===== 类型定义 =====

export interface MessageUserInfo {
  id: number
  username: string
  nickname: string
  avatar: string
}

export interface LastMessageInfo {
  id: number
  messageType: number
  content: string
  mediaUrl: string
  createdAt: string
}

export interface ConversationInfo {
  targetUser: MessageUserInfo
  lastMessage: LastMessageInfo
  unreadCount: number
  updatedAt: string
}

export interface ConversationListResp {
  list: ConversationInfo[]
  total: number
  page: number
  pageSize: number
}

export interface MessageInfo {
  id: number
  fromUserId: number
  toUserId: number
  messageType: number   // 1=文字 2=图片 3=视频
  content: string
  mediaUrl: string
  isRead: boolean
  createdAt: string
}

export interface MessageListResp {
  messages: MessageInfo[]
  nextCursorToken: string
  hasMore: boolean
}

export interface SendMessagePayload {
  toUserId: number
  messageType: number
  content?: string
  mediaUrl?: string
}

// ===== 接口调用 =====

/** 获取当前用户的会话列表（offset 分页） */
export const getConversations = (params: { page?: number; pageSize?: number } = {}) =>
  http.get<ApiResponse<ConversationListResp>>('/conversations', { params })

/** 获取与指定用户的消息列表（游标分页） */
export const getMessages = (userId: number | string, params: { cursor?: string; count?: number } = {}) =>
  http.get<ApiResponse<MessageListResp>>(`/conversations/${userId}/messages`, { params })

/** 标记与指定用户的会话为已读 */
export const markConversationRead = (userId: number | string) =>
  http.put<ApiResponse<{ readCount: number }>>(`/conversations/${userId}/read`)

/** 发送私信 */
export const sendMessage = (payload: SendMessagePayload) =>
  http.post<ApiResponse<MessageInfo>>('/messages', payload)

/** 删除私信 */
export const deleteMessage = (id: number | string) =>
  http.delete<ApiResponse<null>>(`/messages/${id}`)

/** 获取未读消息总数 */
export const getUnreadCount = () =>
  http.get<ApiResponse<{ unreadCount: number }>>('/messages/unread-count')
