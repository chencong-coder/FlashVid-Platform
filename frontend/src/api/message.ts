import http from './http'

import type { ApiResponse } from '@/types/api'

// ===== 类型定义 =====

export type MessageType = 1 | 2 | 3

export interface MessageUserInfo {
  id: string
  username: string
  nickname: string
  avatar: string
}

export interface LastMessageInfo {
  id: string
  messageType: MessageType
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
  id: string
  fromUserId: string
  toUserId: string
  messageType: MessageType // 1=文字 2=图片 3=视频
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

interface SendMessagePayloadBase {
  toUserId: string
}

export interface SendTextMessagePayload extends SendMessagePayloadBase {
  messageType: 1
  content: string
  mediaUrl?: never
}

export interface SendMediaMessagePayload extends SendMessagePayloadBase {
  messageType: 2 | 3
  mediaUrl: string
  content?: string
}

export type SendMessagePayload = SendTextMessagePayload | SendMediaMessagePayload

// ===== 接口调用 =====

/** 获取当前用户的会话列表（offset 分页） */
export const getConversations = (params: { page?: number; pageSize?: number } = {}) =>
  http.get<ApiResponse<ConversationListResp>>('/conversations', { params })

/** 获取与指定用户的消息列表（游标分页） */
export const getMessages = (userId: string, params: { cursor?: string; count?: number } = {}) =>
  http.get<ApiResponse<MessageListResp>>(`/conversations/${userId}/messages`, { params })

/** 标记与指定用户的会话为已读 */
export const markConversationRead = (userId: string) =>
  http.put<ApiResponse<{ readCount: number }>>(`/conversations/${userId}/read`)

/** 发送私信 */
export const sendMessage = (payload: SendMessagePayload) =>
  http.post<ApiResponse<MessageInfo>>('/messages', payload)

/** 删除私信 */
export const deleteMessage = (id: string) =>
  http.delete<ApiResponse<Record<string, never>>>(`/messages/${id}`)

/** 获取未读消息总数 */
export const getUnreadCount = () =>
  http.get<ApiResponse<{ unreadCount: number }>>('/messages/unread-count')
