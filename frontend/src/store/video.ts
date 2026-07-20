import { defineStore } from 'pinia'

import { createMockVideos } from '@/data/mockVideos'
import type { FeedCache, FeedType, VideoItem } from '@/types/video'

interface VideoState {
  activeVideoId: string
  muted: boolean
  feeds: Record<FeedType, FeedCache>
}

const createFeed = (type: FeedType): FeedCache => ({
  items: createMockVideos(type),
  page: 1,
  hasMore: true,
})

export const useVideoStore = defineStore('video', {
  state: (): VideoState => ({
    activeVideoId: '',
    muted: true,
    feeds: {
      recommend: createFeed('recommend'),
      follow: createFeed('follow'),
      nearby: createFeed('nearby'),
    },
  }),
  actions: {
    setActiveVideo(videoId: string): void {
      this.activeVideoId = videoId
    },
    toggleMuted(): void {
      this.muted = !this.muted
    },
    toggleLike(feed: FeedType, videoId: string): void {
      const video = this.findVideo(feed, videoId)
      if (!video) return
      video.liked = !video.liked
      video.stats.likes += video.liked ? 1 : -1
    },
    toggleFavorite(feed: FeedType, videoId: string): void {
      const video = this.findVideo(feed, videoId)
      if (!video) return
      video.favorited = !video.favorited
      video.stats.favorites += video.favorited ? 1 : -1
    },
    toggleFollow(feed: FeedType, videoId: string): void {
      const video = this.findVideo(feed, videoId)
      if (!video) return
      const followed = !video.author.followed
      Object.values(this.feeds).forEach((cache) => {
        cache.items.forEach((item) => {
          if (item.author.id === video.author.id) item.author.followed = followed
        })
      })
    },
    loadMore(feed: FeedType): void {
      const cache = this.feeds[feed]
      if (!cache.hasMore) return
      const nextPage = cache.page + 1
      cache.items.push(...createMockVideos(feed, nextPage))
      cache.page = nextPage
      cache.hasMore = nextPage < 5
    },
    findVideo(feed: FeedType, videoId: string): VideoItem | undefined {
      return this.feeds[feed].items.find((item) => item.id === videoId)
    },
  },
})
