import type { FeedType } from './video'

export type BottomTab = 'home' | 'discover' | 'publish' | 'messages' | 'profile'
export type PopupName = 'search' | 'comments' | 'share' | null

export interface AppState {
  activeTopTab: FeedType
  activeBottomTab: BottomTab
  popup: PopupName
  globalLoading: boolean
}
