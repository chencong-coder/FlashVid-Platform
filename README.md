# FlashVid-Platform

闪视短视频平台 —— 基于 Go + Gin + Redis + RabbitMQ 的高并发短视频系统

## 项目简介

FlashVid 是一个完整的短视频平台，支持视频发布、点赞收藏、Feed 流推荐、话题热榜等核心功能。通过 Redis 缓存 + RabbitMQ 消息队列实现高并发架构，单机支持 100+ 并发，QPS 可达 **3000+**。

## 技术栈

**后端**
- Go 1.23 + Gin（Web 框架）
- GORM（ORM）
- MySQL 8.0（主数据库）
- Redis 7.0（缓存 + 热榜 + Feed）
- RabbitMQ 3.x（消息队列）

**前端**
- 前端工程与开发说明位于 [`frontend/`](./frontend/README.md)

## 核心功能

- ✅ 用户系统：注册/登录、个人主页、关注/粉丝
- ✅ 视频系统：发布视频、视频详情、搜索、删除
- ✅ 互动系统：点赞、收藏、评论、分享
- ✅ Feed 流：推荐 Feed（热度排序）、关注 Feed（写扩散）
- ✅ 话题系统：话题热榜、话题详情、话题视频列表
- ✅ 通知系统：点赞/收藏/关注通知、未读数统计

## 高并发架构亮点

### 1. Feed 流混合推拉模式

**架构设计**：根据发布者粉丝数自适应切换推拉策略，平衡读写性能与存储成本

| 视频流 | 策略 | 数据结构 | 场景 |
|---|---|---|---|
| 推荐 Feed | 拉模式（热度榜） | Redis ZSet | 全局热度排序 |
| 关注 Feed | 混合推拉 | Redis List + MySQL | 粉丝 < 10000 推模式，≥ 10000 拉模式补全 |
| 好友 Feed | 混合推拉 | Redis List + MySQL | 互相关注，同关注流策略 |
| 附近 Feed | 拉模式 | MySQL GeoHash | 地理位置实时计算 |

**关键优化**：
- **普通用户场景**：发布视频时预写到每个粉丝的 Redis List（`feed:follow:{userId}`），读取时 O(1) 秒出
- **大V场景**：粉丝 ≥ 10000 跳过推送，避免百万次写入；粉丝读取时从 Redis List + MySQL 拉模式补全大V视频
- **降级保障**：Redis 不可用时自动降级到 MySQL 实时聚合，三层容错

**收益**：普通场景读取 QPS 711，大V场景避免写放大，单次发布节省 10000+ 次 Redis 写入

### 2. Redis 缓存优化

### 2. Redis 缓存优化

| 场景 | 方案 | 收益 |
|---|---|---|
| 点赞/收藏计数 | Redis HINCRBY 累计增量，定时批量刷 MySQL | QPS 提升至 396-730 |
| 点赞/收藏状态 | Redis Set 存储，天然幂等 | 避免软删唯一键冲突 |
| 视频详情 | Cache-Aside + singleflight 防击穿 | QPS 1090，性能提升 495 倍 |
| 话题热榜 | Redis ZSet 实时排序 | QPS 3069，O(log N) 复杂度 |
| Feed 推荐 | Redis ZSet 热度分数（时间衰减算法） | QPS 614，实时热度推荐 |

**缓存三大问题防护**：
- **防穿透**：空值缓存 60 秒
- **防击穿**：singleflight 分布式锁
- **防雪崩**：TTL 随机偏移（1800 ± 300 秒）

### 3. RabbitMQ 异步化

### 3. RabbitMQ 异步化

| 场景 | Exchange 类型 | 收益 |
|---|---|---|
| 点赞/收藏通知 | Direct | 主流程响应时间减少 30% |
| 视频播放统计 | Direct | 异步更新播放量 + 热度 |
| 热度更新 | Direct | 不阻塞主流程 |
| 视频发布广播 | Fanout | 并行处理热榜 + Feed 推送 |
| Feed 写扩散 | Fanout | 粉丝 < 10000 异步推送，避免主流程阻塞 |

**消息可靠性保障**：
- Consumer 手动 ACK + 重试 3 次
- 死信队列（DLX）托底
- 消息持久化

### 4. 性能数据

| 接口 | QPS | 平均延迟 | 成功率 | 说明 |
|---|---|---|---|---|
| 话题热榜 | **3069.05** | 32.04ms | 100% | Redis ZSet O(log N) |
| 视频详情 | **1090.48** | 91.00ms | 100% | 性能提升 495 倍 |
| 收藏 | **730.69** | 135.14ms | 100% | Redis HINCRBY + Set |
| 关注 Feed | **711.84** | 138.82ms | 100% | Redis List 写扩散 |
| 取消点赞 | **718.98** | 138.12ms | 100% | Redis HINCRBY + Set |
| 取消收藏 | **684.06** | 144.82ms | 100% | Redis HINCRBY + Set |
| Feed 推荐 | **614.59** | 51.96ms | 100% | Redis ZSet 热度排序 |
| 点赞 | **396.48** | 234.71ms | 100% | Redis HINCRBY + Set |

**测试环境**：Windows 11, Go 1.23, MySQL 8.0, Redis 7.0, RabbitMQ 3.x  
**测试方法**：自研 Go 并发压测工具，100 并发 × 1000 请求  
**详细报告**：[docs/benchmark-report.md](./docs/benchmark-report.md)

## 项目结构

```
FlashVid-Platform/
├── flashvid-platform-gin/      # 后端服务（Go + Gin）
│   ├── cmd/server/             # 服务入口
│   ├── internal/               # 内部实现
│   │   ├── consumer/           # RabbitMQ 消费者
│   │   ├── dao/                # 数据访问层
│   │   ├── middleware/         # 中间件
│   │   ├── mq/                 # 消息队列封装
│   │   ├── router/             # 路由
│   │   └── service/            # 业务逻辑
│   ├── pkg/hotrank/            # 热度计算工具
│   └── config/                 # 配置文件
├── frontend/                   # 前端项目
├── docs/                       # 文档
│   ├── benchmark-report.md     # 压测报告
│   ├── high-concurrency-roadmap.md  # 高并发改造路线图
│   └── redis-optimization.md   # Redis 优化方案
└── scripts/                    # 压测脚本
    ├── benchmark_interaction.go      # 点赞/收藏压测
    ├── benchmark_feed.go             # Feed 流压测
    ├── benchmark_video_detail.go     # 视频详情压测
    ├── benchmark_hot_topics.go       # 话题热榜压测
    └── run_all_benchmarks.bat        # 一键运行（Windows）
```

## 快速开始

### 1. 环境准备

```bash
# 启动 MySQL
docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root mysql:8.0

# 启动 Redis
docker run -d -p 6379:6379 redis:7.0

# 启动 RabbitMQ
docker run -d -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

### 2. 启动后端

```bash
cd flashvid-platform-gin
go mod download
go run cmd/server/main.go
```

### 3. 运行压测

```bash
cd scripts
# Windows
run_all_benchmarks.bat

# Linux/Mac
chmod +x run_all_benchmarks.sh
./run_all_benchmarks.sh
```

## 文档

- [压测报告](./docs/benchmark-report.md) - 完整性能测试数据 + 架构优化亮点
- [高并发改造路线图](./docs/high-concurrency-roadmap.md) - 从 0 到 1 的改造过程
- [Redis 优化方案](./docs/redis-optimization.md) - 缓存设计详解
- [压测脚本说明](./scripts/README.md) - 如何运行压测

## 简历亮点

1. **点赞/收藏性能优化**：通过 Redis HINCRBY + Set 实现计数和状态存储，QPS 提升至 **396-730**，平均延迟降至 **135-235ms**
2. **视频详情性能突破**：移除 MQ Confirm 阻塞，QPS 从 2.19 提升至 **1090.48**，性能提升 **495 倍**
3. **热榜实时排序**：使用 Redis ZSet 实现话题/视频热榜，O(log N) 复杂度，QPS **3069**，支持实时更新
4. **缓存三大问题防护**：实现穿透/击穿/雪崩防护方案，缓存命中率 > 90%
5. **异步化解耦**：通过 RabbitMQ 将通知创建、热度更新、Feed 推送异步化，主流程响应时间减少 **30-50%**
6. **消息可靠性保障**：Producer Confirm + Consumer 重试 + 死信队列，消息零丢失
7. **Feed 写扩散**：视频发布时推送到粉丝 Feed，支持大 V 场景（百万粉丝）

## 开发者

- GitHub: [@chencong-coder](https://github.com/chencong-coder)
- 开发时间: 2026

## License

MIT
