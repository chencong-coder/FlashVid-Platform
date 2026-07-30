export interface UserProfile {
  id: string
  nickname: string
  avatar: string
  bio: string
  following: number
  followers: number
  likes: number
}

// 登录 / 注册相关类型已迁移至 @/api/auth（对齐后端 auth 模块）
