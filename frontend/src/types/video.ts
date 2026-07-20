export type FeedType = 'follow' | 'recommend' | 'nearby'

export interface VideoAuthor {
  id: string
  nickname: string
  avatar: string
  verified?: boolean
  followed: boolean
}

export interface VideoStats {
  likes: number
  comments: number
  favorites: number
  shares: number
}

export interface VideoItem {
  id: string
  author: VideoAuthor
  description: string
  topics: string[]
  music: string
  poster: string
  source: string
  discCover: string
  stats: VideoStats
  liked: boolean
  favorited: boolean
  city?: string
}

export interface CommentItem {
  id: string
  userName: string
  avatar: string
  content: string
  time: string
  likes: number
}

export interface FeedCache {
  items: VideoItem[]
  page: number
  hasMore: boolean
}
