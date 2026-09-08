# FlashVid 后端复习 Roadmap（分阶段学习）

> 按照从基础到进阶的顺序，分阶段系统复习后端架构，每个阶段独立完成。
>
> 进度标记：`- [ ]` 未开始 · `- [ ] … 🔄` 进行中 · `- [x]` 已完成

---

## 📋 学习路径

```
阶段 0: 环境与配置 (1-2 小时)
  ↓
阶段 1: 数据层基础 (2-3 小时)
  ↓
阶段 2: 业务层核心 (3-4 小时)
  ↓
阶段 3: 接口层实现 (2-3 小时)
  ↓
阶段 4: 消息队列与异步 (3-4 小时)
  ↓
阶段 5: 高并发优化 (4-5 小时)
  ↓
阶段 6: 完整项目串联 (2-3 小时)
```

**总计：17-24 小时**

---

## 阶段 0: 环境与配置 (1-2 小时)

### 目标
理解项目如何启动、配置如何加载、依赖如何管理

### 0.1 项目结构梳理
- [ ] 阅读 `flashvid-platform-gin/` 目录结构
  - `cmd/server/main.go` - 入口文件
  - `internal/` - 内部业务代码
  - `pkg/` - 公共工具包
  - `api/` - API 定义
  - `config.yaml` - 配置文件

**核心文件**：
```
flashvid-platform-gin/
├── cmd/server/main.go       # 程序入口
├── config.yaml              # 配置
├── internal/
│   ├── config/              # 配置加载
│   ├── dao/                 # 数据访问
│   ├── model/               # 数据模型
│   ├── service/             # 业务逻辑
│   ├── handler/             # HTTP 处理
│   ├── middleware/          # 中间件
│   ├── router/              # 路由
│   ├── mq/                  # 消息队列
│   └── consumer/            # MQ 消费者
└── pkg/
    ├── hotrank/             # 热度算法
    ├── jwt/                 # JWT 工具
    └── upload/              # 文件上传
```

### 0.2 配置文件解读
- [ ] 阅读 `config.yaml`
  - MySQL 连接配置
  - Redis 连接配置
  - RabbitMQ 连接配置
  - JWT 配置
  - 上传配置

**复习要点**：
```yaml
# 记住每个配置的作用
mysql:
  host: "localhost"
  port: 3306
  dbname: "flashvid"

redis:
  host: "localhost"
  port: 6379
  db: 0

rabbitmq:
  host: "localhost"
  port: 5672

jwt:
  secret: "xxx"
  expire_hours: 168  # 7天
```

### 0.3 启动流程分析
- [ ] 阅读 `cmd/server/main.go`
  - 配置加载顺序
  - 数据库初始化顺序
  - 消费者启动方式
  - HTTP 服务启动

**关键代码**：
```go
func main() {
    // 1. 加载配置
    cfg, _ := config.Load("config.yaml")
    
    // 2. 初始化 MySQL
    dao.InitMySQL(dsn)
    
    // 3. 初始化 Redis
    dao.InitRedis(addr, password, db)
    
    // 4. 初始化 RabbitMQ
    mq.MustInitRabbitMQ(url)
    mq.MustDeclareInfrastructure()
    
    // 5. 启动消费者
    go consumer.ConsumeNotification()
    go consumer.ConsumeHotrankUpdate()
    
    // 6. 启动 HTTP 服务
    r := router.Setup()
    r.Run(":8080")
}
```

**面试问题**：
- Q: 为什么消费者要用 `go` 启动？
- A: 异步非阻塞，允许多个消费者并行运行，不影响 HTTP 服务启动

### 0.4 依赖管理
- [ ] 阅读 `go.mod`
  - Gin 框架
  - GORM ORM
  - Redis 客户端
  - RabbitMQ 客户端
  - JWT 库

**核心依赖**：
```
github.com/gin-gonic/gin           # Web 框架
gorm.io/gorm                       # ORM
github.com/redis/go-redis/v9       # Redis
github.com/rabbitmq/amqp091-go     # RabbitMQ
github.com/golang-jwt/jwt/v5       # JWT
go.uber.org/zap                    # 日志
```

---

## 阶段 1: 数据层基础 (2-3 小时)

### 目标
掌握数据库操作、Redis 操作、数据模型设计

### 1.1 数据库模型
- [ ] 阅读 `internal/model/` 下所有模型文件
  - `user.go` - 用户表
  - `video.go` - 视频表
  - `like.go` - 点赞表
  - `favorite.go` - 收藏表
  - `comment.go` - 评论表
  - `notification.go` - 通知表
  - `follow.go` - 关注表
  - `topic.go` - 话题表

**核心字段记忆**：
```go
// Video 表
type Video struct {
    ID            int64     // 主键
    UserID        int64     // 作者 ID (索引)
    Title         string    // 标题
    VideoURL      string    // 视频地址
    CoverURL      string    // 封面地址
    LikeCount     int32     // 点赞数
    FavoriteCount int32     // 收藏数
    CommentCount  int32     // 评论数
    ViewCount     int32     // 观看数
    Status        int8      // 0=审核中 1=通过 2=拒绝
    PublishedAt   time.Time // 发布时间
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     gorm.DeletedAt  // 软删除
}
```

**面试要点**：
- 为什么要软删除？（数据恢复、审计）
- 索引设计原则？（查询频率高的字段、唯一约束）

### 1.2 DAO 层实现
- [ ] 阅读 `internal/dao/dao.go`
  - MySQL 初始化
  - Redis 初始化
  - 连接池配置

**关键代码**：
```go
// MySQL 初始化
func InitMySQL(dsn string) error {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        // PrepareStmt: true,  // SQL 预编译
    })
    DB = db
    return err
}

// Redis 初始化
func InitRedis(addr, password string, db int) error {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
    RedisClient = client
    return nil
}
```

### 1.3 GORM 查询操作
- [ ] 阅读 `internal/dao/query/` 生成的查询代码
- [ ] 学习常用查询模式

**基础操作**：
```go
q := query.Use(dao.DB)

// 1. 单条查询
video, err := q.Video.WithContext(ctx).
    Where(q.Video.ID.Eq(videoId)).
    First()

// 2. 列表查询
videos, err := q.Video.WithContext(ctx).
    Where(q.Video.Status.Eq(1)).
    Order(q.Video.PublishedAt.Desc()).
    Limit(20).
    Find()

// 3. 条件查询
users, err := q.User.WithContext(ctx).
    Where(q.User.Username.Like("%" + keyword + "%")).
    Find()

// 4. 统计查询
count, err := q.Video.WithContext(ctx).
    Where(q.Video.UserID.Eq(userId)).
    Count()
```

### 1.4 事务处理
- [ ] 学习 GORM 事务操作

**示例代码**：
```go
err := query.Q.Transaction(func(tx *query.Query) error {
    // 操作 1
    if err := tx.Like.WithContext(ctx).Create(like); err != nil {
        return err
    }
    
    // 操作 2
    if err := tx.Video.WithContext(ctx).
        Where(tx.Video.ID.Eq(videoId)).
        UpdateSimple(tx.Video.LikeCount.Add(1)); err != nil {
        return err
    }
    
    return nil  // 返回 nil = 提交，返回 error = 回滚
})
```

**面试要点**：
- 事务的 ACID 特性
- 什么时候需要事务？（多表操作、需要原子性）
- 事务的性能影响？（锁竞争、时间不宜过长）

### 1.5 Redis 基础操作
- [ ] 掌握 Redis 五大数据结构的使用场景

**数据结构对照表**：
```go
// 1. String - 缓存对象
rdb.Set(ctx, "video:123", jsonData, 30*time.Minute)
rdb.Get(ctx, "video:123")

// 2. Hash - 缓存字段（统计数据）
rdb.HSet(ctx, "video:123:stats", "like_count", 100)
rdb.HIncrBy(ctx, "video:123:stats", "like_count", 1)
rdb.HGet(ctx, "video:123:stats", "like_count")

// 3. Set - 去重集合（点赞状态）
rdb.SAdd(ctx, "user:456:liked_videos", 123)
rdb.SIsMember(ctx, "user:456:liked_videos", 123)
rdb.SRem(ctx, "user:456:liked_videos", 123)

// 4. ZSet - 排行榜
rdb.ZAdd(ctx, "video:hot", redis.Z{Score: 1000, Member: "123"})
rdb.ZRevRange(ctx, "video:hot", 0, 49)  // Top 50

// 5. List - Feed 流
rdb.LPush(ctx, "feed:follow:456", 123)
rdb.LRange(ctx, "feed:follow:456", 0, 19)  // 最新 20 条
```

**实战练习**：
- [ ] 实现一个缓存函数：先查 Redis，miss 再查 DB
- [ ] 实现一个热榜查询：从 ZSet 读取 Top 50

---

## 阶段 2: 业务层核心 (3-4 小时)

### 目标
理解业务逻辑封装、缓存策略、事务控制、异步解耦

### 2.1 用户服务
- [ ] 阅读 `internal/service/user/`
  - 注册逻辑
  - 登录逻辑（密码校验 + JWT 生成）
  - 用户信息缓存

**核心代码**：
```go
// 登录
func Login(ctx, username, password string) (string, int, error) {
    // 1. 查询用户
    user, err := getUserByUsername(ctx, username)
    if err != nil {
        return "", api.CodeUserNotFound, err
    }
    
    // 2. 校验密码
    if !checkPassword(password, user.Password) {
        return "", api.CodeWrongPassword, errors.New("密码错误")
    }
    
    // 3. 生成 JWT
    token, err := jwt.GenerateToken(user.ID, user.Username)
    if err != nil {
        return "", api.CodeInternalError, err
    }
    
    return token, api.CodeSuccess, nil
}
```

**面试要点**：
- 密码如何存储？（bcrypt 哈希）
- JWT 如何生成和验证？（HS256 签名）
- Token 过期如何处理？（刷新 Token）

### 2.2 视频服务
- [ ] 阅读 `internal/service/video/video.go`
  - 视频详情查询（带缓存）
  - 视频发布
  - 视频列表查询

**缓存策略**：
```go
func GetVideoByIDWithCache(ctx, videoId, viewerId int64) (*model.VideoInfo, int, error) {
    // 1. 查询 Redis 缓存
    cacheKey := fmt.Sprintf("video:%d", videoId)
    cached, err := rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        // 缓存命中
        var video model.Video
        json.Unmarshal([]byte(cached), &video)
        return buildVideoInfo(&video, viewerId), api.CodeSuccess, nil
    }
    
    // 2. Singleflight 防击穿
    result, err, _ := sf.Do(cacheKey, func() (interface{}, error) {
        // 3. 查询数据库
        video, err := getVideoByID(ctx, videoId)
        if err != nil {
            // 防穿透：缓存空值
            rdb.Set(ctx, cacheKey, "null", 60*time.Second)
            return nil, err
        }
        
        // 4. 写入缓存（TTL + 随机偏移防雪崩）
        data, _ := json.Marshal(video)
        ttl := 1800 + rand.Intn(300)  // 30-35 分钟
        rdb.Set(ctx, cacheKey, data, time.Duration(ttl)*time.Second)
        
        return video, nil
    })
    
    if err != nil {
        return nil, api.CodeVideoNotFound, err
    }
    
    video := result.(*model.Video)
    
    // 5. 更新观看计数（Redis）
    statsKey := fmt.Sprintf("video:%d:stats", videoId)
    exists, _ := rdb.Exists(ctx, statsKey).Result()
    if exists == 0 {
        rdb.HSet(ctx, statsKey, "view_count", video.ViewCount+1)
    } else {
        rdb.HIncrBy(ctx, statsKey, "view_count", 1)
    }
    
    // 6. 发送 MQ 消息（异步更新热度 + MySQL）
    msg := mq.HotrankUpdateMessage{
        Action:  "update_video_view",
        VideoID: videoId,
    }
    body, _ := json.Marshal(msg)
    mq.Publish(ctx, "notification.exchange", "hotrank", body)
    
    return buildVideoInfo(video, viewerId), api.CodeSuccess, nil
}
```

**面试要点**：
- 缓存三大问题如何解决？
  - **穿透**：空值缓存
  - **击穿**：Singleflight
  - **雪崩**：TTL 随机偏移
- 为什么观看计数不直接写 MySQL？（热点行写、性能问题）

### 2.3 交互服务（点赞/收藏）
- [ ] 阅读 `internal/service/interaction/interaction.go`
  - 点赞逻辑（幂等性保证）
  - 取消点赞
  - 收藏逻辑

**核心亮点：无锁分布式并发控制**：
```go
func LikeVideo1(ctx, userId, videoId int64) (*v1.LikeVideoResp, int, error) {
    // 1. 查询视频
    video, _ := getVideoByID(ctx, videoId)
    
    // 2. Redis Set 幂等检查（SADD 返回值作为分布式锁）
    userLikedKey := fmt.Sprintf("user:%d:liked_videos", userId)
    videoIdStr := strconv.FormatInt(videoId, 10)
    
    added, err := rdb.SAdd(ctx, userLikedKey, videoIdStr).Result()
    if err != nil {
        // Redis 故障降级到纯 DB 方案
        return LikeVideo(ctx, userId, videoId)
    }
    if added == 0 {
        // 已点赞，幂等返回
        count, _ := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "like_count").Int64()
        return &v1.LikeVideoResp{IsLiked: true, LikeCount: int32(count)}, api.CodeSuccess, nil
    }
    
    // 3. SADD 返回 1，持有分布式锁，执行事务
    like := &model.Like{
        UserID:     userId,
        TargetType: 1,
        TargetID:   videoId,
    }
    
    err = query.Q.Transaction(func(tx *query.Query) error {
        // 3.1 创建点赞记录
        if err := tx.Like.WithContext(ctx).Create(like); err != nil {
            return err
        }
        
        // 3.2 更新 Redis 计数
        statsKey := fmt.Sprintf("video:%d:stats", videoId)
        exists, _ := rdb.Exists(ctx, statsKey).Result()
        if exists == 0 {
            rdb.HSet(ctx, statsKey, "like_count", video.LikeCount+1)
        } else {
            rdb.HIncrBy(ctx, statsKey, "like_count", 1)
        }
        
        return nil
    })
    
    if err != nil {
        // 事务失败，回滚 Redis Set
        rdb.SRem(ctx, userLikedKey, videoIdStr)
        return nil, api.CodeInternalError, err
    }
    
    // 4. 发送 MQ 消息（异步通知 + 热度更新）
    notifMsg := mq.NotificationMessage{
        UserID:     video.UserID,
        ActorID:    userId,
        ActionType: 2,
        TargetType: 2,
        TargetID:   videoId,
    }
    body, _ := json.Marshal(notifMsg)
    mq.Publish(ctx, "notification.exchange", "notification", body)
    
    hotrankMsg := mq.HotrankUpdateMessage{
        Action:  "update_video_hot",
        VideoID: videoId,
    }
    body, _ = json.Marshal(hotrankMsg)
    mq.Publish(ctx, "notification.exchange", "hotrank", body)
    
    // 5. 返回结果
    count, _ := rdb.HGet(ctx, fmt.Sprintf("video:%d:stats", videoId), "like_count").Int64()
    return &v1.LikeVideoResp{IsLiked: true, LikeCount: int32(count)}, api.CodeSuccess, nil
}
```

**面试要点**：
- 为什么用 SADD 返回值作为锁？（原子操作、天然幂等）
- 传统方案的问题？（SIsMember + SADD 有竞态窗口）
- 事务失败如何回滚 Redis？（SRem 删除 Set 中的元素）
- 为什么不在事务里发 MQ？（MQ 失败不应回滚业务，异步解耦）

### 2.4 评论服务
- [ ] 阅读 `internal/service/comment/comment.go`
  - 创建评论（一级评论 + 二级回复）
  - 评论列表查询
  - 回复列表查询

**关键逻辑**：
```go
func CreateComment(ctx, userId, videoId, parentId, replyToUserId int64, content string) {
    // 1. 创建评论
    comment := &model.Comment{
        UserID:        userId,
        VideoID:       videoId,
        ParentID:      parentId,
        ReplyToUserID: replyToUserId,
        Content:       content,
    }
    
    // 2. 事务
    err := query.Q.Transaction(func(tx *query.Query) error {
        if err := tx.Comment.WithContext(ctx).Create(comment); err != nil {
            return err
        }
        
        // 更新父评论回复数 或 视频评论数
        if parentId > 0 {
            tx.Comment.WithContext(ctx).
                Where(tx.Comment.ID.Eq(parentId)).
                UpdateSimple(tx.Comment.ReplyCount.Add(1))
        } else {
            tx.Video.WithContext(ctx).
                Where(tx.Video.ID.Eq(videoId)).
                UpdateSimple(tx.Video.CommentCount.Add(1))
        }
        
        return nil
    })
    
    // 3. 一级评论：更新 Redis 计数 + 发送 MQ
    if parentId == 0 {
        statsKey := fmt.Sprintf("video:%d:stats", videoId)
        exists, _ := rdb.Exists(ctx, statsKey).Result()
        if exists == 0 {
            rdb.HSet(ctx, statsKey, "comment_count", video.CommentCount+1)
        } else {
            rdb.HIncrBy(ctx, statsKey, "comment_count", 1)
        }
        
        msg := mq.HotrankUpdateMessage{
            Action:  "update_video_comment",
            VideoID: videoId,
        }
        body, _ := json.Marshal(msg)
        mq.Publish(ctx, "notification.exchange", "hotrank", body)
    }
}
```

### 2.5 Feed 流服务
- [ ] 阅读 `internal/service/feed/feed.go`
  - 推荐流（热度排序）
  - 关注流（写扩散）

**推荐流（Redis ZSet）**：
```go
func GetRecommendFeed(ctx, cursor int64, pageSize int) ([]*model.VideoInfo, int64, error) {
    // 1. 从 Redis ZSet 取热度 Top N
    members, err := rdb.ZRevRangeWithScores(ctx, "video:hot", cursor, cursor+int64(pageSize)-1).Result()
    if err != nil {
        return nil, 0, err
    }
    
    // 2. 提取 videoId 列表
    var videoIds []int64
    for _, m := range members {
        videoId, _ := strconv.ParseInt(m.Member.(string), 10, 64)
        videoIds = append(videoIds, videoId)
    }
    
    // 3. 批量查询视频详情
    videos, _ := batchGetVideos(ctx, videoIds)
    
    // 4. 返回下一页游标
    nextCursor := cursor + int64(pageSize)
    return videos, nextCursor, nil
}
```

**关注流（Redis List）**：
```go
func GetFollowFeed(ctx, userId, cursor int64, pageSize int) ([]*model.VideoInfo, int64, error) {
    // 1. 从 Redis List 取用户的关注 Feed
    feedKey := fmt.Sprintf("feed:follow:%d", userId)
    videoIdStrs, err := rdb.LRange(ctx, feedKey, cursor, cursor+int64(pageSize)-1).Result()
    if err != nil {
        return nil, 0, err
    }
    
    // 2. 转换为 videoId 列表
    var videoIds []int64
    for _, idStr := range videoIdStrs {
        videoId, _ := strconv.ParseInt(idStr, 10, 64)
        videoIds = append(videoIds, videoId)
    }
    
    // 3. 批量查询视频详情
    videos, _ := batchGetVideos(ctx, videoIds)
    
    nextCursor := cursor + int64(pageSize)
    return videos, nextCursor, nil
}
```

**面试要点**：
- 推荐流和关注流的区别？
  - 推荐流：ZSet 全局热度排序，所有用户一样
  - 关注流：List 个性化 Feed，每个用户不同
- 写扩散的成本？（发布视频时推送给所有粉丝，大 V 成本高）
- 如何优化大 V 场景？（混合模式：小号写扩散，大号读扩散）

---

## 阶段 3: 接口层实现 (2-3 小时)

### 目标
掌握 HTTP 请求处理、参数校验、中间件使用、路由设计

### 3.1 Handler 基础
- [ ] 阅读 `internal/handler/video/video.go`

**标准 Handler 结构**：
```go
func GetVideoDetail(c *gin.Context) {
    // 1. 参数解析
    videoIdStr := c.Param("id")
    videoId, err := strconv.ParseInt(videoIdStr, 10, 64)
    if err != nil {
        c.JSON(200, gin.H{
            "code": api.CodeInvalidParams,
            "msg":  "视频ID格式错误",
        })
        return
    }
    
    // 2. 获取当前用户 ID（从中间件注入）
    userId, exists := c.Get("user_id")
    if !exists {
        userId = int64(0)  // 游客
    }
    viewerId := userId.(int64)
    
    // 3. 调用 Service
    result, code, err := video.GetVideoByIDWithCache(c.Request.Context(), videoId, viewerId)
    if err != nil {
        c.JSON(200, gin.H{
            "code": code,
            "msg":  err.Error(),
        })
        return
    }
    
    // 4. 返回响应
    c.JSON(200, gin.H{
        "code": api.CodeSuccess,
        "msg":  "success",
        "data": result,
    })
}
```

### 3.2 参数绑定与校验
- [ ] 阅读 `internal/handler/user/user.go`

**请求体绑定**：
```go
type LoginRequest struct {
    Username string `json:"username" binding:"required,min=3,max=20"`
    Password string `json:"password" binding:"required,min=6"`
}

func Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(200, gin.H{
            "code": api.CodeInvalidParams,
            "msg":  err.Error(),
        })
        return
    }
    
    // 调用 Service
    token, code, err := user.Login(c.Request.Context(), req.Username, req.Password)
    // ...
}
```

**常用 binding 标签**：
```go
binding:"required"           // 必填
binding:"min=3,max=20"       // 长度限制
binding:"email"              // 邮箱格式
binding:"numeric"            // 数字
binding:"oneof=1 2 3"        // 枚举值
```

### 3.3 中间件实现
- [ ] 阅读 `internal/middleware/auth.go`

**JWT 认证中间件**：
```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 获取 Token
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(200, gin.H{
                "code": api.CodeUnauthorized,
                "msg":  "未登录",
            })
            c.Abort()
            return
        }
        
        // 2. 验证 Token
        claims, err := jwt.ParseToken(token)
        if err != nil {
            c.JSON(200, gin.H{
                "code": api.CodeUnauthorized,
                "msg":  "Token 无效",
            })
            c.Abort()
            return
        }
        
        // 3. 注入用户信息到上下文
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        
        c.Next()
    }
}
```

**CORS 中间件**：
```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

**面试要点**：
- `c.Next()` 和 `c.Abort()` 的区别？
  - Next: 继续执行后续中间件和 Handler
  - Abort: 中止执行，直接返回响应
- 中间件执行顺序？（按注册顺序，CORS → Logger → Auth → Handler）

### 3.4 路由设计
- [ ] 阅读 `internal/router/router.go`

**RESTful API 设计**：
```go
func Setup() *gin.Engine {
    r := gin.Default()
    
    // 全局中间件
    r.Use(middleware.CORSMiddleware())
    r.Use(middleware.LoggerMiddleware())
    
    // 公开路由
    public := r.Group("/api/v1")
    {
        public.POST("/register", user.Register)
        public.POST("/login", user.Login)
        public.GET("/videos/:id", video.GetVideoDetail)
    }
    
    // 需要认证的路由
    auth := r.Group("/api/v1")
    auth.Use(middleware.AuthMiddleware())
    {
        // 用户
        auth.GET("/user/profile", user.GetProfile)
        auth.PUT("/user/profile", user.UpdateProfile)
        
        // 视频
        auth.POST("/videos", video.PublishVideo)
        auth.DELETE("/videos/:id", video.DeleteVideo)
        
        // 交互
        auth.POST("/videos/:id/like", interaction.LikeVideo)
        auth.DELETE("/videos/:id/like", interaction.UnlikeVideo)
        
        // Feed
        auth.GET("/feed/recommend", feed.GetRecommendFeed)
        auth.GET("/feed/follow", feed.GetFollowFeed)
    }
    
    return r
}
```

**RESTful 规范**：
- GET - 查询
- POST - 创建
- PUT - 更新（全量）
- PATCH - 更新（部分）
- DELETE - 删除

---

## 阶段 4: 消息队列与异步 (3-4 小时)

### 目标
理解 RabbitMQ 使用、消息生产与消费、异步解耦

### 4.1 MQ 基础设施
- [ ] 阅读 `internal/mq/mq.go`

**初始化流程**：
```go
// 1. 建立连接
func MustInitRabbitMQ(url string) {
    conn, err = amqp.Dial(url)
    channel, err = conn.Channel()
}

// 2. 声明 Exchange
channel.ExchangeDeclare(
    "notification.exchange", // name
    "direct",                // type
    true,                    // durable
    false, false, false, nil,
)

// 3. 声明 Queue
channel.QueueDeclare(
    "notification.queue",
    true, false, false, false, nil,
)

// 4. 绑定 Queue 到 Exchange
channel.QueueBind(
    "notification.queue",    // queue
    "notification",          // routing key
    "notification.exchange", // exchange
    false, nil,
)
```

**Exchange 类型**：
- **Direct**: 精确匹配 routing key（本项目使用）
- **Fanout**: 广播到所有绑定的队列
- **Topic**: 模式匹配 routing key
- **Headers**: 根据消息头匹配

### 4.2 消息生产
- [ ] 阅读 `internal/mq/mq.go` 的 `Publish` 方法

**发送消息（Confirm 模式）**：
```go
func Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
    // 1. 开启 Confirm 模式
    if err := channel.Confirm(false); err != nil {
        return err
    }
    
    // 2. 发送消息
    err := channel.PublishWithContext(
        ctx,
        exchange,
        routingKey,
        false, false,
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,  // 持久化
        },
    )
    if err != nil {
        return err
    }
    
    // 3. 等待 ACK 确认
    select {
    case <-channel.NotifyPublish(make(chan amqp.Confirmation, 1)):
        return nil  // 成功
    case <-time.After(5 * time.Second):
        return fmt.Errorf("publish timeout")
    }
}
```

**消息体定义**：
```go
// 通知消息
type NotificationMessage struct {
    UserID     int64  `json:"user_id"`
    ActorID    int64  `json:"actor_id"`
    ActionType int    `json:"action_type"`
    TargetType int    `json:"target_type"`
    TargetID   int64  `json:"target_id"`
}

// 热度更新消息
type HotrankUpdateMessage struct {
    Action  string `json:"action"`
    VideoID int64  `json:"video_id"`
    TopicID int64  `json:"topic_id"`
}
```

### 4.3 消息消费
- [ ] 阅读 `internal/consumer/notification.go`

**消费者实现**：
```go
func ConsumeNotification() {
    msgs, err := channel.Consume(
        "notification.queue",
        "",    // consumer
        false, // auto-ack (关闭自动确认)
        false, false, false, nil,
    )
    
    for msg := range msgs {
        handleNotification(msg)
    }
}

func handleNotification(msg amqp.Delivery) {
    ctx := context.Background()
    
    // 1. 解析消息
    var notif mq.NotificationMessage
    if err := json.Unmarshal(msg.Body, &notif); err != nil {
        // 格式错误，拒绝不重试
        msg.Nack(false, false)
        return
    }
    
    // 2. 检查重试次数
    retryCount := GetRetryCount(msg)
    if retryCount >= 3 {
        // 重试超限，拒绝进入死信队列
        zap.L().Error("max retry exceeded", zap.Int64("user_id", notif.UserID))
        msg.Nack(false, false)
        return
    }
    
    // 3. 创建通知记录
    notification := &model.Notification{
        UserID:     notif.UserID,
        ActorID:    notif.ActorID,
        ActionType: int8(notif.ActionType),
        TargetType: int8(notif.TargetType),
        TargetID:   notif.TargetID,
    }
    
    q := query.Use(dao.DB)
    if err := q.Notification.WithContext(ctx).Create(notification); err != nil {
        // 失败重试
        msg.Nack(false, true)
        return
    }
    
    // 4. 更新 Redis 未读数
    rdb := dao.RedisClient
    unreadKey := fmt.Sprintf("user:%d:unread", notif.UserID)
    rdb.HIncrBy(ctx, unreadKey, strconv.Itoa(notif.ActionType), 1)
    
    // 5. ACK 确认
    msg.Ack(false)
}
```

**ACK 策略**：
- `msg.Ack(false)` - 成功处理，确认
- `msg.Nack(false, true)` - 失败重试，重新入队
- `msg.Nack(false, false)` - 失败不重试，丢弃/进入死信队列

### 4.4 热度更新消费者
- [ ] 阅读 `internal/consumer/hotrank_update.go`

**核心逻辑**：
```go
func handleHotrankUpdate(msg amqp.Delivery) {
    var event mq.HotrankUpdateMessage
    json.Unmarshal(msg.Body, &event)
    
    switch event.Action {
    case "update_video_view":
        // 更新热度分数
        hotrank.UpdateVideoHotScore(ctx, event.VideoID)
        // 同步统计数据到 MySQL
        syncVideoStatsToMySQL(ctx, event.VideoID)
        
    case "update_video_hot":
        hotrank.UpdateVideoHotScore(ctx, event.VideoID)
        syncVideoStatsToMySQL(ctx, event.VideoID)
        
    case "update_video_comment":
        hotrank.UpdateVideoHotScore(ctx, event.VideoID)
        syncVideoStatsToMySQL(ctx, event.VideoID)
        
    case "update_topic_view":
        hotrank.UpdateTopicViewCount(ctx, event.TopicID)
        syncTopicStatsToMySQL(ctx, event.TopicID)
    }
    
    msg.Ack(false)
}

// 同步视频统计到 MySQL
func syncVideoStatsToMySQL(ctx context.Context, videoId int64) {
    statsKey := fmt.Sprintf("video:%d:stats", videoId)
    
    // 从 Redis 读取最新计数
    likeCount, _ := rdb.HGet(ctx, statsKey, "like_count").Int64()
    favoriteCount, _ := rdb.HGet(ctx, statsKey, "favorite_count").Int64()
    commentCount, _ := rdb.HGet(ctx, statsKey, "comment_count").Int64()
    viewCount, _ := rdb.HGet(ctx, statsKey, "view_count").Int64()
    
    // 批量更新 MySQL
    q := query.Use(dao.DB)
    q.Video.WithContext(ctx).
        Where(q.Video.ID.Eq(videoId)).
        Updates(map[string]interface{}{
            "like_count":     likeCount,
            "favorite_count": favoriteCount,
            "comment_count":  commentCount,
            "view_count":     viewCount,
        })
}
```

**面试要点**：
- 为什么统计数据要异步同步 MySQL？（避免热点行写、削峰填谷）
- 如何保证消息不丢失？（Confirm 模式 + 手动 ACK + 持久化）
- 如何保证幂等性？（消费者内部去重逻辑）

---

## 阶段 5: 高并发优化 (4-5 小时)

### 目标
掌握缓存优化、分布式锁、热度算法、Feed 流设计

### 5.1 缓存三大问题详解

**问题 1: 缓存穿透（查询不存在的数据）**
```go
// 场景：恶意查询不存在的 videoId，每次都打到 DB
// 解决方案：空值缓存
func GetVideoByID(ctx, videoId int64) (*model.Video, error) {
    // 1. 查询缓存
    cached, err := rdb.Get(ctx, fmt.Sprintf("video:%d", videoId)).Result()
    if err == nil {
        if cached == "null" {
            // 缓存的空值
            return nil, errors.New("视频不存在")
        }
        var video model.Video
        json.Unmarshal([]byte(cached), &video)
        return &video, nil
    }
    
    // 2. 查询数据库
    video, err := queryDB(videoId)
    if err != nil {
        // 缓存空值，TTL 较短
        rdb.Set(ctx, fmt.Sprintf("video:%d", videoId), "null", 60*time.Second)
        return nil, err
    }
    
    // 3. 缓存真实数据
    data, _ := json.Marshal(video)
    rdb.Set(ctx, fmt.Sprintf("video:%d", videoId), data, 30*time.Minute)
    return video, nil
}
```

**问题 2: 缓存击穿（热点 key 过期，大量请求打到 DB）**
```go
// 解决方案：Singleflight
import "golang.org/x/sync/singleflight"

var sf singleflight.Group

func GetVideoByID(ctx, videoId int64) (*model.Video, error) {
    cacheKey := fmt.Sprintf("video:%d", videoId)
    
    // Singleflight: 多个并发请求只执行一次
    result, err, _ := sf.Do(cacheKey, func() (interface{}, error) {
        // 只有第一个请求会执行这里
        cached, err := rdb.Get(ctx, cacheKey).Result()
        if err == nil {
            var video model.Video
            json.Unmarshal([]byte(cached), &video)
            return &video, nil
        }
        
        // 查询数据库
        video, err := queryDB(videoId)
        if err != nil {
            return nil, err
        }
        
        // 写缓存
        data, _ := json.Marshal(video)
        rdb.Set(ctx, cacheKey, data, 30*time.Minute)
        return video, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    return result.(*model.Video), nil
}
```

**问题 3: 缓存雪崩（大量 key 同时过期）**
```go
// 解决方案：TTL 加随机偏移
func SetCache(ctx, key string, value interface{}, baseTTL time.Duration) {
    data, _ := json.Marshal(value)
    
    // 基础 TTL + 随机偏移（0-10% 范围）
    offset := rand.Intn(int(baseTTL.Seconds()) / 10)
    ttl := baseTTL + time.Duration(offset)*time.Second
    
    rdb.Set(ctx, key, data, ttl)
}

// 示例：30 分钟 ± 3 分钟
SetCache(ctx, "video:123", video, 30*time.Minute)
```

### 5.2 分布式锁实现

**方案 1: Redis SADD 作为锁（本项目使用）**
```go
// 优点：原子操作、天然幂等
added, err := rdb.SAdd(ctx, userLikedKey, videoIdStr).Result()
if added == 0 {
    // 已持有锁，幂等返回
    return
}
// added == 1，获取锁成功，执行业务
defer rdb.SRem(ctx, userLikedKey, videoIdStr)  // 释放锁
```

**方案 2: Redis SETNX 作为锁**
```go
// 设置锁
ok, err := rdb.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
if !ok {
    // 锁已被占用
    return errors.New("获取锁失败")
}

// 执行业务
defer rdb.Del(ctx, lockKey)  // 释放锁
```

**对比**：
- SADD: 适合幂等场景（点赞），锁即业务状态
- SETNX: 适合通用锁场景，需要设置过期时间防死锁

### 5.3 热度算法设计

**公式**：
```
热度分数 = 观看数 × 1 + 点赞数 × 5 + 收藏数 × 8 + 评论数 × 10
```

**实现**：
```go
// pkg/hotrank/video_hot.go
func UpdateVideoHotScore(ctx context.Context, videoId int64) {
    // 1. 从 Redis 读取最新统计
    statsKey := fmt.Sprintf("video:%d:stats", videoId)
    viewCount, _ := rdb.HGet(ctx, statsKey, "view_count").Int64()
    likeCount, _ := rdb.HGet(ctx, statsKey, "like_count").Int64()
    favoriteCount, _ := rdb.HGet(ctx, statsKey, "favorite_count").Int64()
    commentCount, _ := rdb.HGet(ctx, statsKey, "comment_count").Int64()
    
    // 2. 计算基础分数
    baseScore := float64(viewCount)*1 + float64(likeCount)*5 +
        float64(favoriteCount)*8 + float64(commentCount)*10
    
    // 3. 加时间戳实现稳定排序（同分数按时间倒序）
    video, _ := getVideoByID(ctx, videoId)
    timestamp := float64(video.PublishedAt.Unix()) / 1e13  // 微调因子
    finalScore := baseScore + timestamp
    
    // 4. 更新 Redis ZSet
    rdb.ZAdd(ctx, "video:hot", redis.Z{
        Score:  finalScore,
        Member: strconv.FormatInt(videoId, 10),
    })
}
```

**稳定排序解释**：
- 如果只用 baseScore，同分数视频排序不稳定（Redis 内部随机）
- 加上 timestamp/1e13（权重很小），同分数按发布时间倒序
- 例如：score=1000.0000001234 → 1000 是热度，0.0000001234 是时间戳影响

### 5.4 Feed 流设计

**推荐流（读扩散）**：
```go
// 优点：发布成本低（只更新一个 ZSet）
// 缺点：用户读取时需要实时计算

func PublishVideo(videoId int64) {
    // 初始化热度到 ZSet
    rdb.ZAdd(ctx, "video:hot", redis.Z{Score: 0, Member: strconv.FormatInt(videoId, 10)})
}

func GetRecommendFeed(ctx, cursor, pageSize int64) []int64 {
    // 从 ZSet 读取 Top N
    members, _ := rdb.ZRevRange(ctx, "video:hot", cursor, cursor+pageSize-1).Result()
    
    var videoIds []int64
    for _, m := range members {
        videoId, _ := strconv.ParseInt(m, 10, 64)
        videoIds = append(videoIds, videoId)
    }
    
    return videoIds
}
```

**关注流（写扩散）**：
```go
// 优点：用户读取快（直接从 List 读）
// 缺点：发布成本高（需要推送给所有粉丝）

func PublishVideo(userId, videoId int64) {
    // 1. 查询粉丝列表
    followers := getFollowers(userId)
    
    // 2. 推送到每个粉丝的 Feed
    for _, followerId := range followers {
        feedKey := fmt.Sprintf("feed:follow:%d", followerId)
        rdb.LPush(ctx, feedKey, videoId)
        rdb.LTrim(ctx, feedKey, 0, 999)  // 保留最新 1000 条
        rdb.Expire(ctx, feedKey, 7*24*time.Hour)  // 7 天过期
    }
}

func GetFollowFeed(ctx, userId, cursor, pageSize int64) []int64 {
    feedKey := fmt.Sprintf("feed:follow:%d", userId)
    videoIdStrs, _ := rdb.LRange(ctx, feedKey, cursor, cursor+pageSize-1).Result()
    
    var videoIds []int64
    for _, idStr := range videoIdStrs {
        videoId, _ := strconv.ParseInt(idStr, 10, 64)
        videoIds = append(videoIds, videoId)
    }
    
    return videoIds
}
```

**对比**：
- 推荐流：适合全局热榜，所有用户一样
- 关注流：适合个性化 Feed，每个用户不同

### 5.5 统一架构优化

**优化前问题**：
- 点赞/收藏在业务层更新 Redis
- 观看/评论在 MQ 消费者更新 Redis
- 架构不一致，职责不清

**优化后方案**：
- 所有计数更新都在业务层同步完成（实时性）
- MQ 消费者只负责热度更新 + MySQL 同步（异步任务）

**代码示例**：
```go
// 业务层：观看视频
func GetVideoByID(ctx, videoId int64) {
    // 1. 查询视频
    video := queryVideo(videoId)
    
    // 2. 更新 Redis 观看计数（同步）
    statsKey := fmt.Sprintf("video:%d:stats", videoId)
    exists, _ := rdb.Exists(ctx, statsKey).Result()
    if exists == 0 {
        rdb.HSet(ctx, statsKey, "view_count", video.ViewCount+1)
    } else {
        rdb.HIncrBy(ctx, statsKey, "view_count", 1)
    }
    
    // 3. 发送 MQ 消息（异步）
    msg := mq.HotrankUpdateMessage{
        Action:  "update_video_view",
        VideoID: videoId,
    }
    body, _ := json.Marshal(msg)
    mq.Publish(ctx, "notification.exchange", "hotrank", body)
}

// MQ 消费者：只负责热度更新 + MySQL 同步
func handleHotrankUpdate(msg amqp.Delivery) {
    var event mq.HotrankUpdateMessage
    json.Unmarshal(msg.Body, &event)
    
    // 更新热度分数
    hotrank.UpdateVideoHotScore(ctx, event.VideoID)
    
    // 从 Redis 读取最新计数并同步到 MySQL
    syncVideoStatsToMySQL(ctx, event.VideoID)
    
    msg.Ack(false)
}
```

**优势**：
- 架构统一：所有计数更新在业务层
- 实时性好：Redis 立即更新，用户立即看到
- 职责清晰：业务层负责计数，消费者负责热度和同步

---

## 阶段 6: 完整项目串联 (2-3 小时)

### 目标
通过完整案例串联所有知识点，理解请求全流程

### 6.1 案例：用户点赞视频

**完整流程**：
```
1. 前端请求
   POST /api/v1/videos/123/like
   Headers: Authorization: Bearer <token>

2. 路由匹配
   router.go → auth.POST("/videos/:id/like", interaction.LikeVideo)

3. 中间件执行
   - CORSMiddleware: 处理跨域
   - LoggerMiddleware: 记录日志
   - AuthMiddleware: 验证 JWT → 注入 user_id=456

4. Handler 处理
   interaction.LikeVideo(c)
   - 解析参数: videoId=123
   - 获取 user_id=456
   - 调用 Service

5. Service 业务逻辑
   interaction.LikeVideo1(ctx, 456, 123)
   
   5.1 查询视频
       video, _ := getVideoByID(ctx, 123)
   
   5.2 Redis Set 幂等检查
       added, _ := rdb.SAdd(ctx, "user:456:liked_videos", "123").Result()
       if added == 0 { return }  // 已点赞
   
   5.3 MySQL 事务
       - 创建 likes 记录
       - 更新 Redis: HIncrBy video:123:stats like_count 1
   
   5.4 事务成功，发送 MQ 消息
       - 通知消息 → notification.queue
       - 热度更新消息 → hotrank.queue
   
   5.5 返回结果
       {is_liked: true, like_count: 101}

6. Handler 返回响应
   c.JSON(200, {code: 0, msg: "success", data: {...}})

7. 异步任务（MQ 消费者）
   
   7.1 通知消费者
       - 创建通知记录到 MySQL
       - 更新 Redis 未读数
   
   7.2 热度更新消费者
       - 重新计算视频热度分数
       - 更新 Redis ZSet: video:hot
       - 同步计数到 MySQL video 表
```

**时序图**：
```
用户 -> 前端 -> Gin -> 中间件 -> Handler -> Service
                                              |
                                              v
                                         Redis Set (幂等)
                                              |
                                              v
                                         MySQL 事务
                                              |
                                              v
                                         Redis Hash (计数)
                                              |
                                              v
                                         RabbitMQ (异步)
                                              |
                +-----------+------------------+------------------+
                |           |                                     |
                v           v                                     v
          通知消费者    热度更新消费者                        返回响应
                |           |                                     |
                v           v                                     v
          创建通知记录  更新热度分数                            用户
          更新未读数    同步 MySQL
```

### 6.2 性能瓶颈分析

**问题 1: 数据库连接数不足**
```go
// 解决方案：配置连接池
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(100)      // 最大连接数
sqlDB.SetMaxIdleConns(10)       // 最大空闲连接
sqlDB.SetConnMaxLifetime(time.Hour)  // 连接最大存活时间
```

**问题 2: Redis 连接数不足**
```go
client := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    PoolSize:     100,               // 连接池大小
    MinIdleConns: 10,                // 最小空闲连接
    MaxRetries:   3,                 // 重试次数
})
```

**问题 3: MQ 消费速度慢**
```go
// 解决方案：多消费者并发
for i := 0; i < 5; i++ {
    go consumer.ConsumeNotification()
}
```

**问题 4: 热点数据查询慢**
```go
// 解决方案：本地缓存 (go-cache)
var localCache = cache.New(5*time.Minute, 10*time.Minute)

func GetVideoByID(videoId int64) *model.Video {
    // 1. 查询本地缓存
    if cached, found := localCache.Get(fmt.Sprintf("video:%d", videoId)); found {
        return cached.(*model.Video)
    }
    
    // 2. 查询 Redis
    // 3. 查询 MySQL
    // 4. 写入本地缓存
    localCache.Set(fmt.Sprintf("video:%d", videoId), video, cache.DefaultExpiration)
}
```

### 6.3 监控与日志

**日志记录**：
```go
// 使用 zap 结构化日志
zap.L().Info("user login",
    zap.Int64("user_id", userId),
    zap.String("username", username),
    zap.String("ip", c.ClientIP()),
)

zap.L().Error("database error",
    zap.Error(err),
    zap.String("sql", sql),
)
```

**性能监控**：
```go
// 中间件记录接口响应时间
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        
        c.Next()
        
        latency := time.Since(start)
        statusCode := c.Writer.Status()
        
        // 上报到监控系统（Prometheus）
        httpRequestDuration.WithLabelValues(path, strconv.Itoa(statusCode)).Observe(latency.Seconds())
    }
}
```

### 6.4 面试高频问题总结

**Q1: 你的项目用了哪些技术栈？**
- Gin 框架、GORM、Redis、RabbitMQ、JWT、Zap

**Q2: 为什么选择这些技术？**
- Gin: 轻量高性能、路由灵活
- GORM: ORM 开发效率高、支持事务
- Redis: 缓存热点数据、ZSet 实现排行榜
- RabbitMQ: 异步解耦、削峰填谷

**Q3: 如何保证接口性能？**
- Redis 缓存热点数据
- 数据库索引优化
- 异步处理非核心逻辑（MQ）
- 分布式锁避免竞态

**Q4: 如何保证数据一致性？**
- Redis 和 MySQL 最终一致（MQ 异步同步）
- 事务失败自动回滚 Redis
- Cache-Aside 策略

**Q5: 如何处理高并发场景？**
- Redis 计数器（避免热点行写）
- 分布式锁（SADD 幂等）
- 消息队列（削峰填谷）
- Singleflight（防缓存击穿）

**Q6: 项目的亮点是什么？**
- 无锁分布式并发控制（SADD 返回值作为锁）
- 统一架构设计（业务层同步更新 Redis + MQ 异步同步 MySQL）
- 缓存三大问题解决方案（穿透/击穿/雪崩）
- 热度算法 + 稳定排序（ZSet + 时间戳）
- 消息可靠性保证（Confirm + 手动 ACK + 死信队列）

---

## 复习建议

### 每个阶段的学习方法

**理论学习**：
1. 阅读对应章节的代码
2. 理解核心概念和设计思路
3. 记忆关键代码片段

**实践练习**：
1. 运行项目，调试断点
2. 修改代码，观察效果
3. 压测接口，分析性能

**面试准备**：
1. 准备 3 分钟项目介绍
2. 准备每个技术点的 1 分钟讲解
3. 准备性能数据和优化效果

### 时间安排

**集中复习（3-4 天）**：
- Day 1: 阶段 0-1（环境配置 + 数据层）
- Day 2: 阶段 2-3（业务层 + 接口层）
- Day 3: 阶段 4-5（MQ + 高并发优化）
- Day 4: 阶段 6（项目串联 + 面试准备）

**分散复习（1-2 周）**：
- 每天 2-3 小时，按阶段推进
- 周末集中练习和总结

---

## 附录：关键文件清单

### 必读文件（核心逻辑）
- [ ] `cmd/server/main.go` - 启动流程
- [ ] `internal/dao/dao.go` - 数据库初始化
- [ ] `internal/service/interaction/interaction.go` - 点赞/收藏逻辑
- [ ] `internal/service/video/video.go` - 视频服务
- [ ] `internal/service/feed/feed.go` - Feed 流
- [ ] `internal/consumer/hotrank_update.go` - 热度更新消费者
- [ ] `internal/mq/mq.go` - MQ 封装
- [ ] `pkg/hotrank/video_hot.go` - 热度算法

### 选读文件（扩展理解）
- [ ] `internal/middleware/auth.go` - JWT 认证
- [ ] `internal/handler/video/video.go` - Handler 示例
- [ ] `internal/service/comment/comment.go` - 评论逻辑
- [ ] `pkg/jwt/jwt.go` - JWT 工具

---

*最后更新: 2025-09-08*
