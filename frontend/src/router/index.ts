import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import type { BottomTab } from '@/types/app'

declare module 'vue-router' {
  interface RouteMeta {
    title: string
    bottomTab?: BottomTab
    hideBottomNav?: boolean
    requiresAuth?: boolean
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'recommend',
    component: () => import('@/views/RecommendView.vue'),
    meta: { title: '推荐', bottomTab: 'home' },
  },
  {
    path: '/follow',
    name: 'follow',
    component: () => import('@/views/FollowView.vue'),
    meta: { title: '关注', bottomTab: 'home' },
  },
  {
    path: '/nearby',
    name: 'nearby',
    component: () => import('@/views/NearbyView.vue'),
    meta: { title: '同城', bottomTab: 'home' },
  },
  {
    path: '/friends',
    name: 'friends',
    component: () => import('@/views/FriendsView.vue'),
    meta: { title: '朋友', bottomTab: 'friends' },
  },
  {
    path: '/discover',
    name: 'discover',
    component: () => import('@/views/DiscoverView.vue'),
    meta: { title: '发现', bottomTab: 'discover' },
  },
  {
    path: '/topic/:id',
    name: 'topic',
    component: () => import('@/views/TopicView.vue'),
    meta: { title: '话题', bottomTab: 'discover' },
  },
  {
    path: '/topic/:id/play',
    name: 'topic-play',
    component: () => import('@/views/feed/TopicPlayView.vue'),
    meta: { title: '话题', bottomTab: 'discover', hideBottomNav: true },
  },
  {
    path: '/publish',
    name: 'publish',
    component: () => import('@/views/PublishView.vue'),
    meta: { title: '发布作品', bottomTab: 'publish', hideBottomNav: true },
  },
  {
    path: '/messages',
    name: 'messages',
    component: () => import('@/views/MessagesView.vue'),
    meta: { title: '消息', bottomTab: 'messages' },
  },
  {
    path: '/chat/:userId',
    name: 'chat',
    component: () => import('@/views/ChatView.vue'),
    meta: { title: '私信', hideBottomNav: true },
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('@/views/ProfileView.vue'),
    meta: { title: '我的', bottomTab: 'profile' },
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  document.title = `${to.meta.title} - 闪视`
  return true
})

export default router
