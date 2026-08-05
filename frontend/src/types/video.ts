export type FeedType = 'follow' | 'recommend' | 'nearby' | 'friends' | 'topic' | 'profile'
export type TopNavValue = 'follow' | 'recommend' | 'nearby' | 'friends' | 'discover'

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

export interface CommentUser {
  id: string
  username: string
  nickname: string
  avatar: string
}

// 被回复的用户（回复里的 replyTo）
export interface ReplyToUser {
  id: string
  nickname: string
}

// 评论回复
export interface ReplyItem {
  id: string
  content: string
  user: CommentUser
  replyTo: ReplyToUser
  likeCount: number
  isLiked: boolean
  createdAt: string
}

export interface CommentItem {
  id: string
  content: string
  user: CommentUser
  likeCount: number
  replyCount: number
  isLiked: boolean
  isAuthored: boolean // 是否为视频作者
  replies: ReplyItem[]
  createdAt: string
}

export interface FeedCache {
  items: VideoItem[]
  cursor: string
  hasMore: boolean
  loading: boolean
  loaded: boolean
}

// 附近流所需的定位坐标
export interface GeoLocation {
  latitude: number
  longitude: number
}
