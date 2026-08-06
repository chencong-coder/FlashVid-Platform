import { defineStore } from 'pinia'

import {
  getFollowFeed,
  getFriendsFeed,
  getNearbyFeed,
  getRecommendFeed,
  mapFeedVideo,
  type FeedParams,
  type FeedResponse,
  type FeedVideo,
} from '@/api/feed'
import { followUser, unfollowUser } from '@/api/user'
import { favoriteVideo, likeVideo, shareVideo, unfavoriteVideo, unlikeVideo } from '@/api/video'
import { getTopicVideos } from '@/api/topic'
import type { FeedCache, FeedType, GeoLocation, VideoItem } from '@/types/video'
import { useUserStore } from '@/store/user'

interface VideoState {
  activeVideoId: string
  muted: boolean
  location: GeoLocation | null
  topicId: string
  topicSort: 'popular' | 'latest'
  feeds: Record<FeedType, FeedCache>
  // 最近一次关注状态变更，供其他组件（如推荐面板）订阅实时更新
  lastFollowChange: { authorId: string; followed: boolean } | null
}

const createFeed = (): FeedCache => ({
  items: [],
  cursor: '',
  hasMore: true,
  loading: false,
  loaded: false,
})

// 按 feed 类型请求后端一页数据（nearby 需要定位，topic 需要 topicId）
const fetchFeedPage = async (
  feed: FeedType,
  params: FeedParams,
  location: GeoLocation | null,
  topic: { id: string; sort: 'popular' | 'latest' },
): Promise<FeedResponse> => {
  const count = 10
  if (feed === 'recommend') return (await getRecommendFeed({ ...params, count })).data.data
  if (feed === 'follow') return (await getFollowFeed({ ...params, count })).data.data
  if (feed === 'friends') return (await getFriendsFeed({ ...params, count })).data.data
  if (feed === 'profile') return { videos: [], nextCursorToken: '', hasMore: false }
  if (feed === 'topic') {
    if (!topic.id) return { videos: [], nextCursorToken: '', hasMore: false }
    return (await getTopicVideos(topic.id, { ...params, count, sort: topic.sort })).data.data
  }
  // nearby
  if (!location) return { videos: [], nextCursorToken: '', hasMore: false }
  return (
    await getNearbyFeed({
      ...params,
      count,
      latitude: location.latitude,
      longitude: location.longitude,
      distance: 10,
    })
  ).data.data
}

export const useVideoStore = defineStore('video', {
  state: (): VideoState => ({
    activeVideoId: '',
    muted: true,
    location: null,
    topicId: '',
    topicSort: 'popular',
    feeds: {
      recommend: createFeed(),
      follow: createFeed(),
      nearby: createFeed(),
      friends: createFeed(),
      topic: createFeed(),
      profile: createFeed(),
    },
    lastFollowChange: null,
  }),
  actions: {
    setActiveVideo(videoId: string): void {
      this.activeVideoId = videoId
    },
    toggleMuted(): void {
      this.muted = !this.muted
    },
    setLocation(location: GeoLocation): void {
      this.location = location
    },
    // 首次加载（如已加载则跳过，除非 force）
    async initFeed(feed: FeedType, force = false): Promise<void> {
      const cache = this.feeds[feed]
      if (cache.loaded && !force) return
      cache.items = []
      cache.cursor = ''
      cache.hasMore = true
      cache.loaded = false
      await this.loadMore(feed)
      cache.loaded = true
    },
    async loadMore(feed: FeedType): Promise<void> {
      const cache = this.feeds[feed]
      if (cache.loading || !cache.hasMore) return
      cache.loading = true
      try {
        const data = await fetchFeedPage(feed, { cursor: cache.cursor }, this.location, {
          id: this.topicId,
          sort: this.topicSort,
        })
        const mapped = (data.videos ?? []).map((v) => mapFeedVideo(v, feed))
        cache.items = [...cache.items, ...mapped]
        cache.cursor = data.nextCursorToken ?? ''
        cache.hasMore = data.hasMore && cache.cursor !== ''
      } catch {
        cache.hasMore = false
      } finally {
        cache.loading = false
      }
    },
    // 从个人主页进入播放：预灌当前 tab 的视频列表
    startProfileFeed(videos: FeedVideo[], startIndex: number): void {
      this.feeds.profile = {
        items: videos.map((v) => mapFeedVideo(v, 'recommend')),
        cursor: '',
        hasMore: false,
        loading: false,
        loaded: true,
      }
    },
    // 从话题页进入播放：预灌已加载的视频 + 游标，避免重复请求，返回起始视频 id
    startTopicFeed(
      topicId: string,
      sort: 'popular' | 'latest',
      videos: VideoItem[],
      cursor: string,
      hasMore: boolean,
    ): void {
      this.topicId = topicId
      this.topicSort = sort
      this.feeds.topic = {
        items: [...videos],
        cursor,
        hasMore,
        loading: false,
        loaded: true,
      }
    },
    async toggleLike(feed: FeedType, videoId: string): Promise<void> {
      const video = this.findVideo(feed, videoId)
      if (!video) return
      const nextLiked = !video.liked
      // 乐观更新
      video.liked = nextLiked
      video.stats.likes += nextLiked ? 1 : -1
      try {
        const res = nextLiked ? await likeVideo(videoId) : await unlikeVideo(videoId)
        // 用后端返回值校准
        video.liked = res.data.data.isLiked
        video.stats.likes = res.data.data.likeCount
      } catch {
        // 回滚
        video.liked = !nextLiked
        video.stats.likes += nextLiked ? -1 : 1
      }
    },
    async toggleFavorite(feed: FeedType, videoId: string): Promise<void> {
      const video = this.findVideo(feed, videoId)
      if (!video) return
      const nextFav = !video.favorited
      video.favorited = nextFav
      video.stats.favorites += nextFav ? 1 : -1
      try {
        const res = nextFav ? await favoriteVideo(videoId) : await unfavoriteVideo(videoId)
        video.favorited = res.data.data.isFavorited
        video.stats.favorites = res.data.data.favoriteCount
      } catch {
        video.favorited = !nextFav
        video.stats.favorites += nextFav ? -1 : 1
      }
    },
    async shareCurrent(feed: FeedType, videoId: string): Promise<string> {
      const video = this.findVideo(feed, videoId)
      try {
        const res = await shareVideo(videoId, 'link')
        if (video) video.stats.shares = res.data.data.shareCount
        return res.data.data.shareUrl || window.location.href
      } catch {
        return window.location.href
      }
    },
    async toggleFollow(feed: FeedType, videoId: string): Promise<void> {
      const video = this.findVideo(feed, videoId)
      if (!video) return
      const authorId = video.author.id
      const nextFollowed = !video.author.followed
      // 乐观更新所有 feed 里同作者的视频
      const applyFollowed = (val: boolean): void => {
        Object.values(this.feeds).forEach((cache) => {
          cache.items.forEach((item) => {
            if (item.author.id === authorId) item.author.followed = val
          })
        })
      }
      applyFollowed(nextFollowed)
      try {
        const res = nextFollowed ? await followUser(authorId) : await unfollowUser(authorId)
        const isFollowing = res.data.data.isFollowing
        applyFollowed(isFollowing)
        // 同步更新当前登录用户的关注数（±1）
        const userStore = useUserStore()
        if (userStore.profile) {
          const delta = isFollowing ? 1 : -1
          userStore.profile.following = Math.max(0, userStore.profile.following + delta)
        }
        // 广播关注状态变更，供推荐面板等组件订阅
        this.lastFollowChange = { authorId, followed: isFollowing }
      } catch {
        applyFollowed(!nextFollowed)
      }
    },
    // 外部页面（个人主页/推荐面板）关注状态变更后，同步到所有 feed 里同作者的视频
    syncAuthorFollowed(authorId: string, followed: boolean): void {
      Object.values(this.feeds).forEach((cache) => {
        cache.items.forEach((item) => {
          if (item.author.id === authorId) item.author.followed = followed
        })
      })
    },
    findVideo(feed: FeedType, videoId: string): VideoItem | undefined {
      return this.feeds[feed].items.find((item) => item.id === videoId)
    },
  },
})
