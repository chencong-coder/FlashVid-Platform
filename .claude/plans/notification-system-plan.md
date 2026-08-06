# 通知系统实现方案

## 需求概述

在消息页（MessagesView）顶部有4个快捷入口图标：
- **粉丝**：谁关注了我
- **赞和收藏**：谁点赞/收藏了我的视频
- **@我的**：谁在评论中@了我
- **评论**：谁评论了我的视频

**目标**：
1. 图标旁边显示未读数（如 "+3"）
2. 点击图标跳转到该类通知的详情页
3. 详情页展示具体的通知列表（头像、昵称、操作描述、时间）
4. 标记已读功能

---

## 一、数据库设计

### 1.1 创建 `notifications` 表

```sql
CREATE TABLE `notifications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `user_id` BIGINT NOT NULL COMMENT '通知接收者ID',
  `actor_id` BIGINT NOT NULL COMMENT '触发者ID（谁做的操作）',
  `action_type` TINYINT NOT NULL COMMENT '1=关注 2=点赞视频 3=收藏视频 4=评论视频 5=回复评论',
  `target_type` TINYINT NOT NULL COMMENT '1=用户 2=视频 3=评论',
  `target_id` BIGINT NOT NULL COMMENT '目标对象ID',
  `content` VARCHAR(500) DEFAULT '' COMMENT '附加内容（评论文本预览等）',
  `is_read` TINYINT DEFAULT 0 COMMENT '0=未读 1=已读',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX `idx_user_created` (`user_id`, `created_at`),
  INDEX `idx_user_action_read` (`user_id`, `action_type`, `is_read`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知表';
```

**字段说明**：
- `action_type` 枚举：
  - 1 = 关注（映射到"粉丝"入口）
  - 2 = 点赞视频（映射到"赞和收藏"）
  - 3 = 收藏视频（映射到"赞和收藏"）
  - 4 = 评论视频（映射到"评论"）
  - 5 = 回复评论（映射到"评论"）
  - 6 = @提及（映射到"@我的"）— 后续扩展

**索引策略**：
- `idx_user_created`：支持按时间倒序拉取用户通知列表
- `idx_user_action_read`：支持按类型+未读状态快速统计

---

## 二、后端实现

### 2.1 修改现有服务层，插入通知记录

#### **关注操作** - `internal/service/user/follow.go`

在 `FollowUser()` 的事务中增加：
```go
// 创建通知（被关注者收到）
notification := &model.Notification{
    UserID:     followUserId,        // 接收者
    ActorID:    loginUserId,         // 触发者
    ActionType: 1,                   // 关注
    TargetType: 1,                   // 目标是用户
    TargetID:   followUserId,
}
if err := tx.Notification.Create(notification); err != nil {
    return err
}
```

#### **点赞视频** - `internal/service/interaction/interaction.go`

在 `LikeVideo()` 的事务中增加：
```go
// 只给视频作者发通知，不给自己发
if video.UserID != userId {
    notification := &model.Notification{
        UserID:     video.UserID,
        ActorID:    userId,
        ActionType: 2,  // 点赞视频
        TargetType: 2,  // 视频
        TargetID:   videoId,
    }
    if err := tx.Notification.Create(notification); err != nil {
        return err
    }
}
```

#### **收藏视频** - `internal/service/interaction/interaction.go`

在 `FavoriteVideo()` 的事务中增加（逻辑同点赞）：
```go
if video.UserID != userId {
    notification := &model.Notification{
        UserID:     video.UserID,
        ActorID:    userId,
        ActionType: 3,  // 收藏视频
        TargetType: 2,
        TargetID:   videoId,
    }
    // ...
}
```

#### **评论视频** - `internal/service/comment/comment.go`

在 `CreateComment()` 中增加：
```go
// 给视频作者发通知（非自己评论自己）
if video.UserID != comment.UserID {
    notification := &model.Notification{
        UserID:     video.UserID,
        ActorID:    comment.UserID,
        ActionType: 4,  // 评论视频
        TargetType: 3,  // 评论
        TargetID:   comment.ID,
        Content:    comment.Content[:min(100, len(comment.Content))], // 预览前100字
    }
    // ...
}

// 如果是回复评论，还要给被回复者发通知
if comment.ParentID > 0 && comment.ReplyToUserID > 0 && comment.ReplyToUserID != comment.UserID {
    notification := &model.Notification{
        UserID:     comment.ReplyToUserID,
        ActorID:    comment.UserID,
        ActionType: 5,  // 回复评论
        TargetType: 3,
        TargetID:   comment.ID,
        Content:    comment.Content[:min(100, len(comment.Content))],
    }
    // ...
}
```

---

### 2.2 新增通知服务层

**文件**：`internal/service/notification/notification.go`

```go
package notification

import (
    "context"
    "flashvid-platform-gin/internal/dao/query"
    "flashvid-platform-gin/internal/model"
)

// GetNotifications 获取通知列表（支持按 action_type 筛选）
func GetNotifications(ctx context.Context, userId int64, actionTypes []int32, page, pageSize int) (
    notifications []*model.NotificationInfo, total int64, err error,
) {
    q := query.Notification.WithContext(ctx).Where(query.Notification.UserID.Eq(userId))
    
    if len(actionTypes) > 0 {
        q = q.Where(query.Notification.ActionType.In(actionTypes...))
    }
    
    total, err = q.Count()
    if err != nil {
        return nil, 0, err
    }
    
    offset := (page - 1) * pageSize
    results, err := q.Order(query.Notification.CreatedAt.Desc()).
        Offset(offset).Limit(pageSize).Find()
    if err != nil {
        return nil, 0, err
    }
    
    // 批量加载 actor 和 target 信息（用户、视频）
    // 转换为 NotificationInfo DTO
    // ...
    
    return notifications, total, nil
}

// GetUnreadCountByType 按类型统计未读数
func GetUnreadCountByType(ctx context.Context, userId int64) (map[int32]int64, error) {
    // SELECT action_type, COUNT(*) FROM notifications 
    // WHERE user_id=? AND is_read=0 GROUP BY action_type
    // ...
}

// MarkAsRead 标记已读
func MarkAsRead(ctx context.Context, userId int64, notificationIds []int64) error {
    _, err := query.Notification.WithContext(ctx).
        Where(query.Notification.UserID.Eq(userId)).
        Where(query.Notification.ID.In(notificationIds...)).
        Update(query.Notification.IsRead, 1)
    return err
}
```

---

### 2.3 新增 Handler 层

**文件**：`internal/handler/notification/notification.go`

```go
package notification

import (
    "github.com/gin-gonic/gin"
    "flashvid-platform-gin/api"
    v1 "flashvid-platform-gin/api/notification/v1"
    "flashvid-platform-gin/internal/service/notification"
)

// GetNotificationsHandler 获取通知列表
func GetNotificationsHandler(c *gin.Context) {
    var req v1.GetNotificationsReq
    if err := c.ShouldBindQuery(&req); err != nil {
        api.ResponseError(c, api.CodeInvalidParam)
        return
    }
    
    userId := c.GetInt64("user_id")
    notifications, total, err := notification.GetNotifications(
        c, userId, req.ActionTypes, req.Page, req.PageSize,
    )
    if err != nil {
        api.ResponseError(c, api.CodeInternalError)
        return
    }
    
    api.ResponseSuccess(c, v1.GetNotificationsResp{
        List: notifications, Total: total,
    })
}

// GetUnreadCountsHandler 获取各类型未读数
func GetUnreadCountsHandler(c *gin.Context) {
    userId := c.GetInt64("user_id")
    counts, err := notification.GetUnreadCountByType(c, userId)
    if err != nil {
        api.ResponseError(c, api.CodeInternalError)
        return
    }
    
    api.ResponseSuccess(c, v1.UnreadCountsResp{
        Followers:     counts[1],  // 粉丝
        LikesAndFavs:  counts[2] + counts[3],  // 赞+收藏
        Mentions:      counts[6],  // @我的
        Comments:      counts[4] + counts[5],  // 评论+回复
    })
}

// MarkAsReadHandler 标记已读
func MarkAsReadHandler(c *gin.Context) {
    var req v1.MarkAsReadReq
    if err := c.ShouldBindJSON(&req); err != nil {
        api.ResponseError(c, api.CodeInvalidParam)
        return
    }
    
    userId := c.GetInt64("user_id")
    if err := notification.MarkAsRead(c, userId, req.NotificationIDs); err != nil {
        api.ResponseError(c, api.CodeInternalError)
        return
    }
    
    api.ResponseSuccess(c, nil)
}
```

---

### 2.4 API 定义

**文件**：`api/notification/v1/notification.go`

```go
package v1

type GetNotificationsReq struct {
    ActionTypes []int32 `form:"action_types"` // 筛选类型，空=全部
    Page        int     `form:"page" binding:"required,min=1"`
    PageSize    int     `form:"page_size" binding:"required,min=1,max=100"`
}

type NotificationInfo struct {
    ID         int64  `json:"id"`
    ActorID    int64  `json:"actorId"`
    ActorName  string `json:"actorName"`
    ActorAvatar string `json:"actorAvatar"`
    ActionType int32  `json:"actionType"`  // 1=关注 2=点赞 3=收藏 4=评论 5=回复
    TargetID   int64  `json:"targetId"`    // 视频ID或评论ID
    TargetTitle string `json:"targetTitle"` // 视频标题
    TargetCover string `json:"targetCover"` // 视频封面
    Content    string `json:"content"`     // 评论内容预览
    IsRead     int32  `json:"isRead"`
    CreatedAt  string `json:"createdAt"`
}

type GetNotificationsResp struct {
    List  []*NotificationInfo `json:"list"`
    Total int64               `json:"total"`
}

type UnreadCountsResp struct {
    Followers    int64 `json:"followers"`    // 粉丝
    LikesAndFavs int64 `json:"likesAndFavs"` // 赞和收藏
    Mentions     int64 `json:"mentions"`     // @我的
    Comments     int64 `json:"comments"`     // 评论
}

type MarkAsReadReq struct {
    NotificationIDs []int64 `json:"notificationIds" binding:"required"`
}
```

---

### 2.5 路由注册

**文件**：`internal/server/route.go`

```go
// 通知路由组
notificationR := apiV1.Group("/notifications").Use(middleware.Auth())
{
    notificationR.GET("", notification.GetNotificationsHandler)
    notificationR.GET("/unread-counts", notification.GetUnreadCountsHandler)
    notificationR.PUT("/read", notification.MarkAsReadHandler)
}
```

---

## 三、前端实现

### 3.1 新增 API 文件

**文件**：`frontend/src/api/notification.ts`

```typescript
import type { ApiResponse } from './types'
import http from './axios'

export interface NotificationItem {
  id: number
  actorId: number
  actorName: string
  actorAvatar: string
  actionType: number  // 1=关注 2=点赞 3=收藏 4=评论 5=回复
  targetId: number
  targetTitle?: string
  targetCover?: string
  content?: string
  isRead: number
  createdAt: string
}

export interface UnreadCounts {
  followers: number
  likesAndFavs: number
  mentions: number
  comments: number
}

export const getNotifications = (params: {
  actionTypes?: number[]
  page: number
  pageSize: number
}) => http.get<ApiResponse<{ list: NotificationItem[]; total: number }>>('/notifications', { params })

export const getUnreadCounts = () =>
  http.get<ApiResponse<UnreadCounts>>('/notifications/unread-counts')

export const markAsRead = (notificationIds: number[]) =>
  http.put<ApiResponse<null>>('/notifications/read', { notificationIds })
```

---

### 3.2 新增 Store

**文件**：`frontend/src/store/notification.ts`

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUnreadCounts, type UnreadCounts } from '@/api/notification'

export const useNotificationStore = defineStore('notification', () => {
  const unreadCounts = ref<UnreadCounts>({
    followers: 0,
    likesAndFavs: 0,
    mentions: 0,
    comments: 0,
  })

  const fetchUnreadCounts = async () => {
    try {
      const res = await getUnreadCounts()
      if (res.data.code === 0 && res.data.data) {
        unreadCounts.value = res.data.data
      }
    } catch {
      // 静默失败
    }
  }

  const totalUnread = computed(() => {
    return (
      unreadCounts.value.followers +
      unreadCounts.value.likesAndFavs +
      unreadCounts.value.mentions +
      unreadCounts.value.comments
    )
  })

  return {
    unreadCounts,
    totalUnread,
    fetchUnreadCounts,
  }
})
```

---

### 3.3 修改 MessagesView.vue

**目标**：
1. 在 `shortcuts` 数组中动态绑定未读数
2. 添加点击事件跳转到通知详情页

```vue
<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '@/store/notification'

const router = useRouter()
const notificationStore = useNotificationStore()

const shortcuts = computed(() => [
  {
    icon: 'fa-user-plus',
    label: '粉丝',
    color: 'bg-blue-500',
    count: notificationStore.unreadCounts.followers,
    route: '/notifications/followers',
  },
  {
    icon: 'fa-heart',
    label: '赞和收藏',
    color: 'bg-orange-500',
    count: notificationStore.unreadCounts.likesAndFavs,
    route: '/notifications/likes-favs',
  },
  {
    icon: 'fa-at',
    label: '@我的',
    color: 'bg-cyan-500',
    count: notificationStore.unreadCounts.mentions,
    route: '/notifications/mentions',
  },
  {
    icon: 'fa-comment',
    label: '评论',
    color: 'bg-green-500',
    count: notificationStore.unreadCounts.comments,
    route: '/notifications/comments',
  },
])

onMounted(() => {
  notificationStore.fetchUnreadCounts()
})

const goToNotifications = (route: string) => {
  router.push(route)
}
</script>

<template>
  <!-- ... -->
  <div class="grid grid-cols-4 gap-4 border-b border-white/5 p-4">
    <button
      v-for="item in shortcuts"
      :key="item.label"
      type="button"
      class="relative flex flex-col items-center gap-2"
      @click="goToNotifications(item.route)"
    >
      <div
        class="flex h-12 w-12 items-center justify-center rounded-full text-white"
        :class="item.color"
      >
        <i :class="`fa-solid ${item.icon}`" />
      </div>
      <span class="text-xs text-neutral-400">{{ item.label }}</span>
      <!-- 未读角标 -->
      <span
        v-if="item.count > 0"
        class="absolute right-0 top-0 rounded-full bg-red-500 px-1.5 py-0.5 text-xs font-semibold text-white"
      >
        {{ item.count > 99 ? '99+' : `+${item.count}` }}
      </span>
    </button>
  </div>
  <!-- ... -->
</template>
```

---

### 3.4 新增通知详情页

**文件**：`frontend/src/views/NotificationListView.vue`

```vue
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getNotifications, markAsRead, type NotificationItem } from '@/api/notification'
import { useNotificationStore } from '@/store/notification'

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()

const type = computed(() => route.params.type as string)
const title = computed(() => {
  const map: Record<string, string> = {
    followers: '粉丝',
    'likes-favs': '赞和收藏',
    mentions: '@我的',
    comments: '评论',
  }
  return map[type.value] || '通知'
})

// 类型映射
const actionTypeMap: Record<string, number[]> = {
  followers: [1],         // 关注
  'likes-favs': [2, 3],   // 点赞+收藏
  mentions: [6],          // @提及
  comments: [4, 5],       // 评论+回复
}

const notifications = ref<NotificationItem[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = 20

const loadNotifications = async () => {
  loading.value = true
  try {
    const res = await getNotifications({
      actionTypes: actionTypeMap[type.value],
      page: page.value,
      pageSize,
    })
    if (res.data.code === 0) {
      notifications.value = res.data.data?.list ?? []
      // 标记为已读
      const unreadIds = notifications.value.filter(n => n.isRead === 0).map(n => n.id)
      if (unreadIds.length > 0) {
        await markAsRead(unreadIds)
        notificationStore.fetchUnreadCounts()
      }
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadNotifications()
})

const getActionText = (item: NotificationItem) => {
  const map: Record<number, string> = {
    1: '关注了你',
    2: '赞了你的视频',
    3: '收藏了你的视频',
    4: '评论了你的视频',
    5: '回复了你',
  }
  return map[item.actionType] || ''
}

const handleItemClick = (item: NotificationItem) => {
  if (item.actionType === 1) {
    // 关注 → 跳转到用户主页
    router.push(`/profile/${item.actorId}`)
  } else if ([2, 3, 4, 5].includes(item.actionType)) {
    // 点赞/收藏/评论 → 跳转到视频详情
    router.push(`/video/${item.targetId}`)
  }
}
</script>

<template>
  <div class="min-h-screen bg-[#0f0f0f] text-white">
    <!-- 标题栏 -->
    <header class="flex items-center gap-3 border-b border-white/5 px-4 py-3">
      <button type="button" @click="router.back()">
        <i class="fa-solid fa-arrow-left text-lg" />
      </button>
      <h1 class="text-lg font-semibold">{{ title }}</h1>
    </header>

    <!-- 加载中 -->
    <div v-if="loading" class="flex justify-center py-10">
      <i class="fa-solid fa-circle-notch animate-spin text-2xl text-neutral-500" />
    </div>

    <!-- 列表 -->
    <div v-else-if="notifications.length > 0" class="divide-y divide-white/5">
      <button
        v-for="item in notifications"
        :key="item.id"
        type="button"
        class="flex w-full gap-3 px-4 py-3 text-left transition-colors hover:bg-white/5"
        @click="handleItemClick(item)"
      >
        <img
          :src="item.actorAvatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${item.actorId}`"
          :alt="item.actorName"
          class="h-12 w-12 shrink-0 rounded-full object-cover"
        />
        <div class="min-w-0 flex-1">
          <p class="mb-1 text-sm">
            <span class="font-semibold">{{ item.actorName }}</span>
            <span class="text-neutral-400"> {{ getActionText(item) }}</span>
          </p>
          <p v-if="item.content" class="mb-1 truncate text-xs text-neutral-500">
            {{ item.content }}
          </p>
          <p class="text-xs text-neutral-600">{{ item.createdAt }}</p>
        </div>
        <img
          v-if="item.targetCover"
          :src="item.targetCover"
          class="h-12 w-12 shrink-0 rounded object-cover"
        />
      </button>
    </div>

    <!-- 空状态 -->
    <div v-else class="py-20 text-center text-sm text-neutral-500">
      暂无通知
    </div>
  </div>
</template>
```

---

### 3.5 路由配置

**文件**：`frontend/src/router/index.ts`

```typescript
{
  path: '/notifications/:type',
  name: 'notification-list',
  component: () => import('@/views/NotificationListView.vue'),
  meta: {
    requiresAuth: true,
    hideBottomNav: true,
  },
}
```

---

## 四、实施步骤

### Phase 1: 数据库 + 后端核心（优先级最高）
1. ✅ 创建 `notifications` 表（编写 SQL migration）
2. ✅ 用 gorm gen 生成 model 和 query 代码
3. ✅ 修改 4 个服务层写通知记录（follow/like/favorite/comment）
4. ✅ 实现 notification service 层（查询+统计+已读）
5. ✅ 实现 notification handler 层
6. ✅ API 定义文件
7. ✅ 路由注册

### Phase 2: 前端基础（次优先级）
1. ✅ 新增 `api/notification.ts`
2. ✅ 新增 `store/notification.ts`
3. ✅ 修改 `MessagesView.vue`（加未读数+跳转）

### Phase 3: 前端详情页（最后）
1. ✅ 新增 `NotificationListView.vue`
2. ✅ 路由配置
3. ✅ 测试完整流程

---

## 五、技术细节

### 5.1 去重策略
**问题**：用户 A 短时间内多次点赞同一视频，会产生多条通知吗？

**方案**：在现有点赞/收藏/关注服务中已有幂等检查，重复操作直接返回成功，不会执行事务，因此不会产生重复通知。

### 5.2 通知聚合
**问题**：A、B、C 三人点赞了同一视频，是否需要聚合为"A、B、C 和其他人赞了你的视频"？

**方案**：**暂不聚合**，保持简单。每个操作独立一条通知，按时间倒序展示。后续可扩展聚合逻辑。

### 5.3 性能优化
- **批量查询**：通知列表需要 JOIN users 和 videos 表，service 层应批量加载避免 N+1 查询
- **索引**：`idx_user_created` 支持高效分页查询，`idx_user_action_read` 支持快速统计未读数

### 5.4 软删除
notifications 表**不使用软删除**，用户无法删除通知（只能标记已读），简化逻辑。

---

## 六、用户体验优化

### 6.1 实时更新
- 用户点赞/关注后，通知接收者的未读数**不会**实时更新（需要刷新或重新进入消息页）
- 后续可接入 WebSocket 推送实时通知

### 6.2 底部导航 Badge
- 消息 tab 图标上显示**总未读数**（私信未读 + 通知未读）
- 在 `BottomNav.vue` 中订阅 `notificationStore.totalUnread`

### 6.3 通知过期清理
- 建议后台定时任务删除 **超过 30 天且已读** 的通知，控制表大小

---

## 七、测试用例

1. **关注通知**：A 关注 B → B 的"粉丝"未读数 +1 → B 点击进入看到"A 关注了你"
2. **点赞通知**：A 点赞 B 的视频 → B 的"赞和收藏"未读数 +1 → 点击跳转到视频详情
3. **自己操作不产生通知**：A 点赞自己的视频 → 不产生通知
4. **幂等性**：A 重复关注 B → 只产生一条通知
5. **标记已读**：进入通知列表后，未读通知自动标记为已读，角标消失

---

## 八、风险与注意事项

### 8.1 数据迁移
- 新增 notifications 表后，**历史操作不会补通知**
- 上线后从零开始积累通知记录

### 8.2 通知爆炸
- 热门视频可能收到大量点赞通知，导致作者通知列表被刷屏
- 后续可考虑聚合策略："100+ 人赞了你的视频"

### 8.3 评论 @ 提及功能
- 当前方案预留了 `action_type=6` 的 @提及通知
- 需要前端在评论输入框实现 @用户 解析逻辑
- 后端在创建评论时解析 content 中的 @用户名，创建额外的通知记录

---

## 九、后续扩展

1. **WebSocket 实时推送**：用户在线时实时收到通知弹窗
2. **通知设置**：允许用户关闭某些类型的通知
3. **邮件/短信通知**：重要通知（被大 V 关注）发送邮件提醒
4. **通知聚合**：相同视频的多个点赞聚合展示
5. **@提及功能**：评论中 @用户 触发通知

---

**END**
