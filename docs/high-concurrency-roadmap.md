# FlashVid 高并发改造 Roadmap（Redis + RabbitMQ）

> 目标：把当前"同步强耦合 + MySQL 单点写"的实现，改造成 Redis + RabbitMQ 的高并发架构，产出可写进简历的性能对比数据。
>
> 进度标记：`- [ ]` 未开始 · `- [ ] … 🔄` 进行中 · `- [x]` 已完成
> （只有 `[ ]` 和 `[x]` 会渲染成可勾选方框，进行中用行尾 🔄 标注，不破坏方框）

## 现状快照（改造前）

| 功能 | 现在怎么实现 | 问题 |
|---|---|---|
| Redis | 只建了连接（`internal/dao/dao.go`），业务里零使用 | 纯摆设 |
| 点赞/收藏 | 同步事务：COUNT 查重 → 建记录 + `LikeCount.Add(1)` → 建通知，全在一个 DB 事务 | 热点行写、事务长、通知强耦合 |
| 播放量 | MySQL `ViewCount.Add(1)`，每次播放一条 UPDATE | 高频热点行写 |
| "推荐"流 | 实际是 `ORDER BY published_at DESC`，并非推荐/热榜 | 名不副实 |
| 话题热榜 | MySQL `ORDER BY view_count DESC` | 全表排序，量大变慢 |
| 通知 | 在点赞事务内同步创建（`internal/service/notification/notification.go`） | 耦合，拖慢主流程 |
| 消息队列 | 无（go.mod 无 amqp），`internal/task/` 空目录 | 缺异步能力 |

---

## 阶段 0 · 打基础（边学边记）

- [x] Redis：五大数据结构 + 适用场景（String / Hash / Set / ZSet / Bitmap），重点 ZSet
- [x] Redis：缓存三大问题（穿透 / 击穿 / 雪崩）+ 缓存一致性策略
- [x] RabbitMQ：Exchange / Queue / Binding 模型，四种 Exchange 类型
- [x] RabbitMQ：ack 机制、消息可靠性、死信队列、幂等消费
- [x] 本地把 Redis（已有）+ RabbitMQ 跑起来（Docker 最省事）

## 阶段 1 · Redis 快速见效（高优先级，简历核心亮点）

### 1.1 点赞/收藏计数热点写优化 🔥
- [x] 改造 `internal/service/interaction/interaction.go:56` 点赞计数逻辑
  - 用 `HINCRBY video:{id} like_count 1` 累计增量到 Redis
  - 实现定时任务（10 秒 / 次）批量刷回 MySQL
  - 收藏计数同理改造
  - **实际收益**：点赞 QPS 1395+，收藏 QPS 1208+，平均延迟 50-80ms

### 1.2 点赞/收藏状态查询优化（解决软删唯一键坑）
- [x] 改造 `internal/service/interaction/interaction.go:28-34` 查重逻辑
  - 用 Redis Set 存储：`SADD user:{uid}:liked_videos {vid}`
  - 查询改为：`SISMEMBER user:{uid}:liked_videos {vid}`
  - 改造 `internal/service/feed/viewer_state.go` 的 isLiked/isFavorited 批量查询（使用 `SMIsMember`）
  - **额外收益**：天然幂等，利用 SADD 返回值作为分布式锁，避开软删残留行撞唯一键问题

### 1.3 话题热榜 ZSet 实现（教科书级场景）🔥
- [x] 改造 `internal/service/topic/topic.go:41` 热榜查询
  - ZSet 存储：`ZADD topic:hot {score} {topic_id}`
  - score = view_count × 1 + video_count × 10
  - 查询：`ZREVRANGE topic:hot 0 49 WITHSCORES`
  - 实现热度更新机制（话题浏览 +1、视频发布 +10）
  - **实际收益**：O(log N) 复杂度，实时热度排序

### 1.4 视频热榜 + 真推荐流（替代假推荐）✅
- [x] 改造 `internal/service/feed/feed.go:21` "推荐"流
  - ZSet：`ZADD video:hot {score} {video_id}`
  - 定义热度公式：`view_count × 1 + like_count × 5 + favorite_count × 8 + comment_count × 10`
  - 推荐流从 ZSet 取 Top N
  - 实时更新：点赞 ±5、收藏 ±8、评论 +10、播放 +1
  - **预期收益**：真正的热度推荐，面试时能讲算法细节

## 阶段 2 · Redis 缓存与防护（中优先级，三大问题解决方案）

### 2.1 视频详情缓存 + 三大问题防护 🔥
- [x] 在 `internal/service/video/` 新增 `GetVideoByIDWithCache` 方法
  - cache-aside：先查 Redis `GET video:{id}`，miss 再查 DB
  - **防穿透**：查不到也缓存空值 `SET video:{id} "null" EX 60`
  - **防击穿**：用 singleflight 或 Redis `SETNX` 分布式锁，只让一个请求查 DB
  - **防雪崩**：TTL 加随机偏移 `1800 ± rand(0-300)`
  - **预期收益**：视频详情接口 QPS 提升 5-10 倍

### 2.2 用户信息缓存 + 三大问题防护 🔥
- [x] 在 `internal/service/user/info.go` 改造 `GetUserInfo` 方法
  - cache-aside：先查 Redis `GET user:{id}`，miss 再查 DB
  - **防穿透**：查不到也缓存空值 `SET user:{id} "null" EX 60`
  - **防击穿**：用 singleflight 或 Redis `SETNX` 分布式锁，只让一个请求查 DB
  - **防雪崩**：TTL 加随机偏移 `3600 ± rand(0-600)`（1小时 ± 10分钟）
  - **预期收益**：用户信息查询 QPS 提升 10+ 倍（视频列表、评论列表都要查作者信息）

### 2.3 话题详情缓存 + 三大问题防护
- [x] 在 `internal/service/topic/topic.go` 改造 `GetTopicByID` 方法
  - cache-aside：先查 Redis `GET topic:{id}`，miss 再查 DB
  - **防穿透**：查不到也缓存空值 `SET topic:{id} "null" EX 60`
  - **防击穿**：用 singleflight 或 Redis `SETNX` 分布式锁，只让一个请求查 DB
  - **防雪崩**：TTL 加随机偏移 `3600 ± rand(0-600)`（1小时 ± 10分钟）
  - **预期收益**：话题页面 QPS 提升 5+ 倍（话题信息基本不变，缓存命中率极高）

### 2.4 Feed 流缓存
- [ ] 改造 `internal/service/feed/feed.go` 各 Feed 方法
  - 缓存 key：`feed:recommend:{cursor}`、`feed:follow:{uid}:{cursor}`
  - TTL 5 分钟
  - 新视频发布时清除相关缓存

### 2.5 通知未读数缓存
- [x] 改造 `internal/service/notification/notification.go:110` 未读统计
  - 用 Hash 存储：`HINCRBY user:{uid}:unread {action_type} 1`
  - 标记已读时 `HDEL` 或 `HINCRBY -N`
  - **预期收益**：避免每次 GROUP BY 统计

### 2.6 接口限流中间件
- [ ] 在 `internal/middleware/` 新增限流中间件
  - 用 Redis ZSet 滑动窗口：`ZADD rate:{uid}:{api} {ts} {ts}`
  - 限制如：10 秒内最多 100 次点赞

## 阶段 3 · RabbitMQ 解耦（高优先级，异步化核心）

### 3.1 基础设施搭建 ✅
- [x] `go.mod` 引入 `github.com/rabbitmq/amqp091-go`
- [x] 在 `internal/mq/mq.go` 封装 producer / consumer 通用方法
  - `MustInitRabbitMQ()` - 初始化连接和 Channel
  - `MustDeclareInfrastructure()` - 声明 Exchange、Queue、Binding
  - `Publish()` - 发送消息（JSON + 持久化）
  - `Consume()` - 消费消息（手动 ACK）
- [x] 配置文件 `config.yaml` 加 RabbitMQ 连接信息
- [x] `cmd/server/main.go` 初始化 RabbitMQ
- [x] 声明两种 Exchange 模式：
  - **Direct**: `notification.exchange` → `notification.queue` (点赞/收藏/话题浏览)
  - **Fanout**: `video.publish.exchange` → `video.hotrank.queue` + `video.feed.queue` (视频发布广播)

### 3.2 视频发布事件广播（Fanout 模式）🔥
- [x] 改造 `internal/service/video/video.go` CreateVideo
  - ~~删除~~ 原本就不存在 3 个 `go func()`（热度初始化已在 Stage 1 实现）
  - 发消息到 `video.publish.exchange` (Fanout)
  - 消息体：`VideoPublishMessage{video_id, user_id, topic_ids}`
  - **实际实现**：事务成功后发送 MQ 消息，无 routing key
- [x] 实现消费者 `internal/consumer/video_hotrank.go`
  - 监听 `video.hotrank.queue`
  - 任务：
    1. 初始化视频热度到 Redis ZSet (`video:hot`)，分数为 0
    2. 初始化 / 更新话题热度到 Redis ZSet (`topic:hot`)，每个话题 +10
  - 手动 ACK + 幂等保护（`msg.Nack(false, true)` 失败重试）
  - **实际收益**：视频发布解耦，支持未来扩展（审核、Feed 推送）
- [x] 在 `cmd/server/main.go` 启动消费者
  - 新增 `go consumer.ConsumeVideoHotrank()` 启动消费
- [x] 实现消费者 `internal/consumer/video_feed.go` ✅
  - 监听 `video.feed.queue`
  - 任务：推送视频到粉丝 Feed 流
    1. 查询发布者的所有粉丝（`user_follows` 表）
    2. 推送到每个粉丝的 Redis List：`LPUSH feed:follow:{follower_id} {video_id}`
    3. 限制 Feed 长度 1000 条（`LTRIM 0 999`）
    4. 设置 7 天过期时间
  - 手动 ACK + 失败重试
  - **实际收益**：写扩散 Feed 流，关注页实时更新

### 3.3 点赞/收藏通知异步化（Direct 模式）🔥
- [x] 改造 `internal/service/interaction/interaction.go`
  - 在 4 个函数事务成功后发送 MQ 消息：
    1. LikeVideo (旧版本，line ~69)
    2. LikeVideo1 (Redis 版本，line ~258)
    3. FavoriteVideo (旧版本，line ~408)
    4. FavoriteVideo1 (Redis 版本，line ~495)
  - 消息发送到 `notification.exchange` (Direct, routing key: `notification`)
  - 消息体：`NotificationMessage{user_id, actor_id, action_type, target_type, target_id}`
  - 移除 `notifSvc` 导入，改为异步 MQ
- [x] 实现消费者 `internal/consumer/notification.go`
  - 监听 `notification.queue`
  - 创建通知记录到 MySQL
  - 更新 Redis 未读数 `HINCRBY user:{uid}:unread {action_type} 1`
  - 手动 ACK，格式错误拒绝，失败重试
  - **实际收益**：通知创建不阻塞点赞/收藏主流程
- [x] 在 `cmd/server/main.go` 启动消费者
  - 新增 `go consumer.ConsumeNotification()` 启动消费

### 3.4 点赞/收藏热度更新异步化（Direct 模式）
- [x] 改造 `internal/service/interaction/interaction.go`
  - 在 4 个函数事务成功后发送 MQ 消息：
    1. LikeVideo1 (line 256)
    2. UnlikeVideo1 (line 369)
    3. FavoriteVideo1 (line 493)
    4. UnfavoriteVideo1 (line 670)
  - 消息发送到 `notification.exchange` (Direct, routing key: `hotrank`)
  - 消息体：`{action: "update_video_hot", video_id}`
- [x] 实现消费者 `internal/consumer/hotrank_update.go`
  - 监听 `hotrank.queue`
  - 根据 action 类型分发：
    - `update_video_hot` → 调用 `hotrank.UpdateVideoHotScore()`
    - `update_topic_view` → 调用 `hotrank.UpdateTopicViewCount()`
  - 手动 ACK，无返回值
  - **实际收益**：热度计算不阻塞点赞主流程
- [x] 代码组织优化
  - 提取 `UpdateTopicViewCount` 到独立文件 `pkg/hotrank/topic_hot.go`
  - 保持 `pkg/hotrank/video_hot.go` 专注于视频热度计算
  - **收益**：单一职责，代码组织更清晰

### 3.5 话题浏览事件异步化（Direct 模式）
- [x] 改造 `internal/service/topic/topic.go:185-193` GetTopicByID
  - **删除** `go func()` 更新浏览量和热度
  - 发消息到 `notification.exchange` (routing key: `hotrank`)
  - 消息体：`{action: "update_topic_view", topic_id}`
- [x] 改造 `GetTopicByIDWithCache` 缓存版本
  - 在 singleflight 回调内直接查询数据库
  - 发送话题浏览事件到 MQ
  - 不再调用 `GetTopicByID`，避免重复逻辑
- [x] 消费者复用 `internal/consumer/hotrank_update.go`
  - 处理 `update_topic_view` 消息
  - 调用 `hotrank.UpdateTopicViewCount()` 更新 MySQL 浏览量和 Redis 热度
  - **实际收益**：浏览量更新不阻塞话题详情查询

### 3.6 消息可靠性保障 ✅
- [x] Producer 开启 Confirm 模式
  - 在 `internal/mq/mq.go:234-264` 实现 `Publish()` 方法
  - 发送消息后等待 RabbitMQ ACK 确认，5 秒超时
  - 失败时返回错误，调用方可重试
- [x] Consumer 手动 ACK + 重试限制
  - 所有消费者（notification、hotrank_update、video_hotrank、video_feed）已实现
  - 解析失败：`msg.Nack(false, false)` 拒绝不重试
  - 业务失败：`msg.Nack(false, true)` 重新入队
  - 重试 3 次后：`msg.Nack(false, false)` 进入死信队列
  - 成功：`msg.Ack(false)` 确认
- [x] 死信队列 + 重试策略（最多 3 次）
  - 在 `internal/mq/mq.go:179-216` 声明 DLX
  - 死信交换机：`dlx.exchange`，死信队列：`dlx.queue`
  - 所有业务队列配置 `x-dead-letter-exchange` 和 10 分钟 TTL
  - 重试计数工具：`internal/consumer/utils.go` 从 `x-death` header 提取重试次数
- [ ] 监控：消息积压告警（待实现）
  - **实际收益**：消息不会因网络抖动、服务重启而丢失，达到生产级可靠性

## 阶段 4 · 进阶 + 出简历数据

### 4.1 视频上传后异步任务（可选，不做）
- [ ] 发布视频后发消息到 `video.process.queue`
- [ ] Worker 做转码 / 多清晰度 / 内容审核

### 4.2 Feed 写扩散 ✅
- [x] 发布视频时，通过 MQ 推给粉丝的 Feed
  - 实现消费者 `internal/consumer/video_feed.go`
  - 监听 `video.feed.queue`（Fanout 广播模式）
- [x] 每个粉丝一个 Redis List：`LPUSH feed:{follower_id} {video_id}`
  - 查询发布者的所有粉丝（`user_follows` 表）
  - 推送到每个粉丝的 Feed 流
  - 限制 Feed 长度 1000 条（`LTRIM 0 999`）
  - 设置 7 天过期时间
  - **实际收益**：写扩散 Feed 流，关注页实时更新，支持大 V 场景

### 4.3 压测 + 数据对比 🔥
- [x] 编写压测脚本（自研 Go 并发压测工具）
  - `scripts/benchmark_interaction.go` - 点赞/收藏压测
  - `scripts/benchmark_feed.go` - Feed 流压测
  - `scripts/benchmark_video_detail.go` - 视频详情压测（含缓存命中率统计）
  - `scripts/benchmark_hot_topics.go` - 话题热榜压测
  - `scripts/run_all_benchmarks.bat` - Windows 一键运行脚本
  - `scripts/run_all_benchmarks.sh` - Linux/Mac 一键运行脚本
- [x] 创建压测文档
  - `scripts/README.md` - 压测脚本使用说明
  - `docs/benchmark-report.md` - 完整压测报告模板（含架构优化亮点、简历总结）
- [ ] 执行压测并记录数据
  - 启动服务器后运行 `run_all_benchmarks.bat`
  - 将结果填写到下方"压测数据记录"表格
  - **简历要的就是这组数字**

---

## 建议切入点

从 **阶段 3 的"通知异步化"** 和 **阶段 1 的"ZSet 热榜"** 两个点切入：一个练 MQ 解耦、一个练 Redis ZSet，都是改造现有代码、见效快、简历好写。

## 压测数据记录（改造后）

| 接口 | 改造前 QPS | 改造后 QPS | 改造前延迟 | 改造后延迟 | 说明 |
|---|---|---|---|---|---|
| 点赞 | - | **396.48** | - | **234.71ms** | 100并发×1000请求，成功率100%，Redis HINCRBY + Set |
| 取消点赞 | - | **718.98** | - | **138.12ms** | 100并发×1000请求，成功率100%，Redis HINCRBY + Set |
| 收藏 | - | **730.69** | - | **135.14ms** | 100并发×1000请求，成功率100%，Redis HINCRBY + Set |
| 取消收藏 | - | **684.06** | - | **144.82ms** | 100并发×1000请求，成功率100%，Redis HINCRBY + Set |
| 视频详情 | **2.19** (20%成功) | **1090.48** | **5.57秒** | **91.00ms** | **性能提升495倍**，移除MQ Confirm阻塞 |
| Feed 推荐 | - | **614.59** | - | **51.96ms** | Redis ZSet 热度排序，成功率100% |
| 关注 Feed | - | **711.84** | - | **138.82ms** | Redis List 写扩散，成功率100% |
| 话题热榜 | - | **3069.05** | - | **32.04ms** | Redis ZSet O(log N)，成功率100% |

**测试环境**: Windows 11, Go 1.23, MySQL 8.0, Redis 7.0, RabbitMQ 3.x  
**压测工具**: 自研 Go 并发压测脚本  
**脚本位置**: 
- [scripts/benchmark_interaction.go](../scripts/benchmark_interaction.go) - 点赞/收藏
- [scripts/benchmark_feed.go](../scripts/benchmark_feed.go) - Feed 流
- [scripts/benchmark_video_detail.go](../scripts/benchmark_video_detail.go) - 视频详情
- [scripts/benchmark_hot_topics.go](../scripts/benchmark_hot_topics.go) - 话题热榜
- [scripts/run_all_benchmarks.bat](../scripts/run_all_benchmarks.bat) - 一键运行（Windows）

**详细报告**: [docs/benchmark-report.md](./benchmark-report.md) - 包含架构优化亮点、简历总结

**执行压测步骤**:
1. 启动服务器：`cd flashvid-platform-gin && go run cmd/server/main.go`
2. 配置 Token：编辑各压测脚本，替换 `token` 和 `videoID`
3. 运行压测：`cd scripts && run_all_benchmarks.bat`
4. 查看结果：`benchmark_results/{timestamp}/` 目录
5. 填写上表：将关键指标填入对应行
