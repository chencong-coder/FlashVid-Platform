import type { CommentItem, FeedType, VideoItem } from '@/types/video'

const posters = [
  'https://images.unsplash.com/photo-1578662996442-48f60103fc96?q=85&w=1080&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1536440136628-849c177e76a1?q=85&w=1080&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?q=85&w=1080&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1492684223066-81342ee5ff30?q=85&w=1080&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1500534314209-a25ddb2bd429?q=85&w=1080&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1514525253161-7a46d19cd819?q=85&w=1080&auto=format&fit=crop',
]

const avatars = [
  'https://images.unsplash.com/photo-1535713875002-d1d0cf3774f2?q=80&w=200&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1544005313-94ddf0286df2?q=80&w=200&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=200&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?q=80&w=200&auto=format&fit=crop',
]

const videoSources = [
  'https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4',
  'https://media.w3.org/2010/05/sintel/trailer.mp4',
  'https://media.w3.org/2010/05/bunny/trailer.mp4',
  'https://media.w3.org/2010/05/video/movie_300.mp4',
]

const descriptions = [
  '第27集：中了一张高级住宅住房券，刚搬进去就发现了不对劲...全文已完结，宝子们放心观看哦',
  '海边日落实况，沉浸式感受晚风，把今天的不开心都留在这里。',
  '城市夜行记录，转过街角刚好遇到一场浪漫的演出。',
  '周末出逃计划，沿着山路一直走，就能看见云海。',
]

const names = ['凌晨睡不着', '户外旅行家', '可乐要加冰', '阿川的镜头']

export const createMockVideos = (feed: FeedType, page = 1, count = 6): VideoItem[] =>
  Array.from({ length: count }, (_, index) => {
    const seed = (page - 1) * count + index
    const nameIndex = seed % names.length
    const authorName = names[nameIndex] ?? '闪视用户'
    const avatar = avatars[nameIndex] ?? avatars[0] ?? ''
    const poster = posters[seed % posters.length] ?? posters[0] ?? ''
    return {
      id: `${feed}-${page}-${index}`,
      author: {
        id: `author-${nameIndex}`,
        nickname: authorName,
        avatar,
        followed: feed === 'follow' || seed % 4 === 0,
      },
      description: descriptions[nameIndex] ?? descriptions[0] ?? '',
      topics:
        feed === 'nearby'
          ? ['同城生活', '城市漫游']
          : ['一口气看完系列', seed % 2 === 0 ? '全文已完结' : '日常vlog'],
      music: `原声 - ${authorName}`,
      poster,
      source: videoSources[seed % videoSources.length] ?? videoSources[0] ?? '',
      discCover: poster,
      stats: {
        likes: 68000 - seed * 917,
        comments: 1385 + seed * 23,
        favorites: 22000 - seed * 163,
        shares: 9539 - seed * 71,
      },
      liked: false,
      favorited: false,
      city: feed === 'nearby' ? '上海 · 2.4km' : undefined,
    }
  })

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
