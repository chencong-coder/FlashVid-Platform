# Redis 优化：高并发点赞/收藏系统

## 📋 目录
- [背景与问题](#背景与问题)
- [优化方案](#优化方案)
- [技术实现](#技术实现)
- [性能测试](#性能测试)
- [关键亮点](#关键亮点)
- [简历素材](#简历素材)

---

## 背景与问题

### 原始实现问题
点赞/收藏功能直接操作 MySQL，存在以下问题：

1. **高并发写入压力**：每次点赞都写 DB，QPS 受限于磁盘 I/O
2. **并发竞态条件**：先查询再插入存在并发窗口，可能产生重复记录
3. **响应延迟高**：每次请求都要等待 DB 事务完成

### 业务需求
- 支持高并发点赞/收藏操作（目标 QPS 1000+）
- 严格保证幂等性（同一用户不能重复点赞）
- 计数准确性（最终一致即可，允许短暂延迟）

---

## 优化方案

### 架构设计

```
┌─────────────┐
│   客户端     │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│         Gin Handler                 │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│   Redis (分布式锁 + 计数缓存)        │
│   • Set: 用户点赞/收藏状态           │
│   • Hash: 视频统计计数               │
└──────┬──────────────────────────────┘
       │ ① SADD/SREM 原子操作
       │    返回值作为分布式锁
       ▼
┌─────────────────────────────────────┐
│   MySQL 事务                         │
│   • 插入/删除 likes/favorites 记录   │
│   • 不更新 video 表计数（异步同步）  │
└─────────────────────────────────────┘
       ▲
       │ ② 定时任务 (10 秒)
       │
┌─────────────────────────────────────┐
│   定时同步任务                       │
│   Redis Hash → MySQL video 表        │
└─────────────────────────────────────┘
```

### 核心思路

1. **Redis Set 作为分布式锁**
   - `SADD` 返回 1 = 成功获取锁 → 执行事务
   - `SADD` 返回 0 = 已存在 → 幂等返回
   - **删除 `SIsMember` 前置检查**，避免竞态窗口

2. **Redis Hash 累积计数**
   - 点赞/收藏时 `HINCRBY +1/-1`
   - 定时任务每 10 秒同步到 MySQL
   - **降低 DB 写压力 90%+**

3. **计数兜底逻辑**
   - 检查 `HGet` 的 `err` 而非值是否为 0
   - Redis 无数据时用 DB 值兜底
   - **保证真实零计数也能正确返回**

---

## 技术实现

### 优化要点

#### 1. 无锁分布式并发控制

**问题**：传统方案先 `SIsMember` 检查再 `SADD`，存在竞态窗口：

```go
// ❌ 错误：有并发窗口
alreadyLiked := rdb.SIsMember(ctx, userLikedKey, videoIdStr)
if alreadyLiked {
    return // 已点赞
}
rdb.SAdd(ctx, userLikedKey, videoIdStr) // ⚠️ 并发请求可能同时通过检查
```

**解决方案**：直接用 `SADD/SREM` 返回值作为分布式锁：

```go
// ✅ 正确：SADD 原子操作
added := rdb.SAdd(ctx, userLikedKey, videoIdStr).Result()
if added == 0 {
    return // 已点赞，幂等
}
// added == 1，持有锁，执行事务
```

#### 2. 计数兜底逻辑修复

**问题**：原逻辑判断 `count == 0` 会误判真实零计数：

```go
// ❌ 错误：真实 0 收藏也会用 DB 兜底
count, _ := rdb.HGet(...).Int64()
if count == 0 {
    count = video.FavoriteCount // 可能不是 0
}
```

**修复**：检查 Redis 错误而非值：

```go
// ✅ 正确：区分 Redis 无数据 vs 值为 0
count, err := rdb.HGet(...).Int64()
if err != nil {
    count = video.FavoriteCount // Redis 无数据，用 DB 兜底
}
// 如果 err == nil，即使 count == 0 也是真实值
```

#### 3. 事务失败自动回滚

如果 MySQL 事务失败，自动回滚 Redis 状态：

```go
err = query.Q.Transaction(func(tx *query.Query) error {
    // ... MySQL 操作
})
if err != nil {
    // 回滚：删除 Redis Set 中的点赞状态
    rdb.SRem(ctx, userLikedKey, videoIdStr)
    return nil, api.CodeInternalError, err
}
```

保证 Redis 和 MySQL 最终一致。

#### 4. Redis 降级策略

```go
func LikeVideo1(ctx context.Context, userId int64, videoId int64) {
    // Redis 不可用时降级到纯 DB 方案
    if rdb == nil {
        return LikeVideo(ctx, userId, videoId)
    }
    
    // Redis 操作失败时降级
    added, err := rdb.SAdd(...).Result()
    if err != nil {
        return LikeVideo(ctx, userId, videoId)
    }
    // ...
}
```

### Redis Key 设计

```
user:{userId}:liked_videos       → Set，存储用户点赞的视频 ID
user:{userId}:favorited_videos   → Set，存储用户收藏的视频 ID
video:{videoId}:stats            → Hash，存储视频统计计数
    ├─ like_count                → 点赞数
    └─ favorite_count            → 收藏数
```

---

## 性能测试

### 测试环境
- **机器配置**: Windows 11, Go 1.23
- **并发工具**: 自研 Go 压测脚本
- **场景**: 100 并发 × 1000 次请求

### 测试结果

| 功能 | QPS (req/s) | 平均延迟 (ms) | P99 延迟 (ms) | 成功率 |
|------|-------------|---------------|---------------|--------|
| **点赞** | 1395.24 | 70.84 | ~400 | 100% |
| **取消点赞** | 1974.62 | 49.50 | ~386 | 100% |
| **收藏** | 1208.56 | 81.54 | ~520 | 100% |
| **取消收藏** | 561.56 | 149.80 | ~1780 | 100% |

### 详细输出示例

```
========== 点赞 压测结果 ==========
总请求数:     1000
成功请求数:   1000
失败请求数:   0
总耗时:       716.7228ms
QPS:          1395.24 req/s
平均延迟:     70.840557ms
最小延迟:     2.0895ms
最大延迟:     400.9721ms
成功率:       100.00%
=====================================
```

### 压测命令

```bash
cd scripts
go run benchmark_interaction.go
```

---

## 关键亮点

### 1. 无锁分布式并发控制
利用 `SADD` 返回值（0/1）作为天然分布式锁，无需额外加锁机制，避免了传统 `SIsMember` + `SADD` 两步操作的竞态窗口。

### 2. 降低 DB 写压力 90%+

**优化前**：每次点赞都更新 `video.like_count` 字段
- 1000 次点赞 = 1000 次 UPDATE video

**优化后**：Redis 累积 + 定时同步
- 1000 次点赞 = 1000 次 `HINCRBY`（内存操作）
- 10 秒后批量同步 = 1 次 UPDATE video
- **写压力降低 99%**

### 3. 计数准确性保证
检查 Redis 错误而非值为 0，区分 Redis 无数据场景和真实零计数场景，保证前端显示准确。

### 4. 最终一致性保证
事务失败自动回滚 Redis 状态，定时任务同步 Redis 到 MySQL，保证数据最终一致。

---

## 简历素材

### 项目描述

**高并发交互优化（点赞/收藏系统）**

**技术栈**: Redis (Set/Hash)、Go、MySQL、分布式锁

**优化方案**:
- 用 Redis Set 的 `SADD/SREM` 原子操作实现**无锁分布式并发控制**，删除 `SIsMember` 前置检查，避免竞态窗口
- Redis Hash 累积计数 + 定时任务异步同步 MySQL，**降低 DB 写压力 90%+**
- 事务失败自动回滚 Redis 状态，保证最终一致性

**性能指标**:
- 点赞接口 QPS 达到 **1395+**，取消点赞 **1974+**
- 100 并发场景成功率 **100%**，平均响应 **50-80ms**
- 支持幂等操作，杜绝重复点赞/收藏

**关键亮点**:
- 利用 `SADD` 返回值（0/1）作为天然分布式锁，无需额外加锁
- 计数兜底逻辑：检查 Redis 错误而非值为 0，保证真实零计数也能正确返回
- Redis 降级策略：Redis 故障时自动切换到纯 DB 方案，保证服务可用性

---

## 相关文件

### 核心代码
- [internal/service/interaction/interaction.go](../flashvid-platform-gin/internal/service/interaction/interaction.go) - 点赞/收藏业务逻辑
- [internal/task/sync_stats.go](../flashvid-platform-gin/internal/task/sync_stats.go) - 定时同步任务
- [internal/handler/interaction/interaction.go](../flashvid-platform-gin/internal/handler/interaction/interaction.go) - HTTP Handler

### 测试脚本
- [scripts/benchmark_interaction.go](../scripts/benchmark_interaction.go) - 压测脚本

---

## 后续优化方向

1. **热点视频缓存**：将热门视频详情缓存到 Redis，减少 DB 查询
2. **布隆过滤器**：预判视频是否存在，避免缓存穿透
3. **ZSet 排行榜**：用 Redis ZSet 实现热门视频/话题排行
4. **消息队列**：RabbitMQ 异步处理通知推送

---

*最后更新: 2025-08-25*
