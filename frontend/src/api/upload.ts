import http from './http'

import type { ApiResponse } from '@/types/api'

export type UploadFileType = 'image' | 'video' | 'audio'

// 后端返回 snake_case
export interface UploadResp {
  file_url: string
  file_size: number
  file_type: UploadFileType
  duration: number // 视频/音频时长（秒），本地存储暂固定为 0
}

/**
 * 上传文件（需登录），multipart/form-data。
 * @param file 文件对象（来自 input / vant Uploader）
 * @param fileType 文件类型：image / video / audio
 * @param onProgress 可选，上传进度回调 (0-100)
 */
export const uploadFile = (
  file: File,
  fileType: UploadFileType,
  onProgress?: (percent: number) => void,
) => {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('file_type', fileType)

  return http.post<ApiResponse<UploadResp>>('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded / e.total) * 100))
      }
    },
  })
}
