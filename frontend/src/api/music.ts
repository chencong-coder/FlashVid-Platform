import http from './http'

import type { ApiResponse } from '@/types/api'

// ===== 音乐 =====
export interface MusicItem {
  id: number
  name: string
  artist: string
  album: string
  coverUrl: string
  musicUrl: string
  duration: number // 秒
  useCount: number // 被使用次数
  createdAt: string
}

// 扁平分页（音乐用 offset 分页，非游标）
export interface MusicListResp {
  list: MusicItem[]
  total: number
  page: number
  pageSize: number
}

export interface MusicListParams {
  sort?: 'hot' | 'latest' // 默认 hot
  page?: number
  pageSize?: number // 最大 100
}

// 获取音乐列表（公开）
export const getMusicList = (params: MusicListParams = {}) =>
  http.get<ApiResponse<MusicListResp>>('/music', { params })

// 搜索音乐（公开，匹配曲名或艺术家）
export const searchMusic = (keyword: string, page?: number, pageSize?: number) =>
  http.get<ApiResponse<MusicListResp>>('/music/search', {
    params: { keyword, page, pageSize },
  })
