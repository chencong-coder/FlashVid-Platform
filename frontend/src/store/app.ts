import { defineStore } from 'pinia'

import type { AppState, BottomTab, PopupName } from '@/types/app'
import type { FeedType } from '@/types/video'

export const useAppStore = defineStore('app', {
  state: (): AppState => ({
    activeTopTab: 'recommend',
    activeBottomTab: 'home',
    popup: null,
    globalLoading: false,
  }),
  actions: {
    setTopTab(tab: FeedType): void {
      this.activeTopTab = tab
    },
    setBottomTab(tab: BottomTab): void {
      this.activeBottomTab = tab
    },
    openPopup(name: Exclude<PopupName, null>): void {
      this.popup = name
    },
    closePopup(): void {
      this.popup = null
    },
    setLoading(loading: boolean): void {
      this.globalLoading = loading
    },
  },
})
