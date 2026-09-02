# 压测脚本使用说明

本目录包含多个压测脚本，用于测试 FlashVid 平台的高并发性能。

## 脚本列表

### 1. benchmark_interaction.go - 点赞/收藏压测
测试点赞、取消点赞、收藏、取消收藏四个接口的性能。

**运行方式：**
```bash
cd scripts
go run benchmark_interaction.go
```

**测试场景：**
- 100 并发 × 1000 次点赞请求
- 100 并发 × 1000 次取消点赞请求
- 100 并发 × 1000 次收藏请求
- 100 并发 × 1000 次取消收藏请求

### 2. benchmark_feed.go - Feed 流压测
测试推荐 Feed 和关注 Feed 的性能。

**运行方式：**
```bash
cd scripts
go run benchmark_feed.go
```

**测试场景：**
- 100 并发 × 1000 次推荐 Feed 请求
- 100 并发 × 1000 次关注 Feed 请求

### 3. benchmark_video_detail.go - 视频详情压测
测试视频详情接口的性能和缓存命中率。

**运行方式：**
```bash
cd scripts
go run benchmark_video_detail.go
```

**测试场景：**
- 预热缓存 3 次
- 100 并发 × 1000 次视频详情请求
- 统计缓存命中率（延迟 < 10ms 视为命中）

### 4. benchmark_hot_topics.go - 话题热榜压测
测试话题热榜接口的性能。

**运行方式：**
```bash
cd scripts
go run benchmark_hot_topics.go
```

**测试场景：**
- 100 并发 × 1000 次话题热榜请求

## 配置说明

每个脚本都需要配置以下常量：

```go
const (
    baseURL = "http://localhost:8089"  // 服务器地址
    token   = "your_jwt_token_here"    // JWT Token
    videoID = 1                         // 视频ID（仅 interaction 和 video_detail 需要）
)
```

### 获取 Token

1. 登录系统获取 JWT Token
2. 将 Token 替换到脚本中的 `token` 常量
3. 确保 Token 未过期

### 选择测试视频

对于 `benchmark_interaction.go` 和 `benchmark_video_detail.go`：
- 选择一个存在的视频 ID
- 建议使用非自己发布的视频（避免不能给自己视频点赞的限制）

## 压测指标说明

每个脚本输出以下指标：

- **总请求数**：发送的总请求数量
- **成功请求数**：返回成功的请求数量
- **失败请求数**：返回失败的请求数量
- **总耗时**：压测总耗时
- **QPS**：每秒处理请求数（Query Per Second）
- **平均延迟**：所有成功请求的平均响应时间
- **最小延迟**：最快的响应时间
- **最大延迟**：最慢的响应时间
- **成功率**：成功请求占总请求的百分比

视频详情脚本额外输出：
- **缓存命中**：缓存命中次数
- **缓存未命中**：缓存未命中次数
- **缓存命中率**：缓存命中占成功请求的百分比

## 注意事项

1. **服务器准备**：确保服务器已启动并运行在 `baseURL` 指定的地址
2. **数据准备**：确保数据库有测试数据（视频、用户、话题等）
3. **Redis/RabbitMQ**：确保 Redis 和 RabbitMQ 服务正常运行
4. **并发调整**：可根据服务器性能调整并发数和请求数
5. **间隔时间**：脚本在不同场景间有 2 秒间隔，避免相互干扰

## 压测结果记录

将压测结果填写到 `docs/high-concurrency-roadmap.md` 的"压测数据记录"表格中。

## 示例输出

```
========== 点赞 压测结果 ==========
总请求数:     1000
成功请求数:   1000
失败请求数:   0
总耗时:       716.789ms
QPS:          1395.24 req/s
平均延迟:     71.679ms
最小延迟:     10.234ms
最大延迟:     412.567ms
成功率:       100.00%
=====================================
```
