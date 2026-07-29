import type { CommentItem } from '@/types/video'

const avatars = [
  'https://images.unsplash.com/photo-1535713875002-d1d0cf3774f2?q=80&w=200&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1544005313-94ddf0286df2?q=80&w=200&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=200&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?q=80&w=200&auto=format&fit=crop',
]

// TODO: 评论区暂用 mock，待评论接口接入后替换
export const mockComments: CommentItem[] = [
  {
    id: 'comment-1',
    userName: '星河入梦',
    avatar: avatars[2] ?? '',
    content: '这个氛围感太好了，已经循环看了好多遍。',
    time: '2小时前',
    likes: 1268,
  },
  {
    id: 'comment-2',
    userName: '慢慢生活',
    avatar: avatars[1] ?? '',
    content: '求背景音乐，画面和节奏都很舒服。',
    time: '昨天',
    likes: 389,
  },
  {
    id: 'comment-3',
    userName: '今天也要开心',
    avatar: avatars[3] ?? '',
    content: '已经收藏，周末也去打卡。',
    time: '3天前',
    likes: 96,
  },
]
