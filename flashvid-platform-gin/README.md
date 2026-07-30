# FlashVid Platform - 后端服务

基于 **Gin + GORM Gen** 的短视频平台后端服务，采用清晰的四层架构。

基于 [Qimi 的 gin-base-layout](https://github.com/q1mi/gin-base-layout) 脚手架改造。

## 📁 项目结构

```bash
├── api/                          # 请求 / 响应类型定义
│   ├── api.go                   # ResponseSuccess / ResponseError
│   ├── code.go                  # 业务错误码
│   ├── auth/v1/
│   ├── user/v1/
│   ├── video/v1/
│   ├── feed/v1/
│   ├── topic/v1/
│   ├── comment/v1/
│   ├── interaction/v1/
│   ├── music/v1/
│   └── upload/v1/
├── cmd/
│   ├── gen/                     # GORM Gen 代码生成工具
│   └── server/                  # 服务启动入口
├── config/
│   └── config.yaml              # 服务配置（数据库 / JWT / 存储等）
├── docs/
│   ├── api.yaml                 # OpenAPI 3.0 规范
│   ├── api-design.md            # API 设计文档
│   ├── database-design.md       # 数据库设计文档
│   └── schema.sql               # 建表 SQL
├── internal/
│   ├── conf/                    # 配置结构体
│   ├── dao/                     # 数据库初始化（MySQL + Redis）
│   │   └── query/              # ⚠️ GORM Gen 自动生成，勿手改
│   ├── handler/                 # HTTP 处理层（参数绑定 / 校验）
│   ├── middleware/              # 中间件（JWT 认证）
│   ├── model/                   # 业务输出模型 + Gen 生成的数据库模型
│   ├── repository/              # 复杂查询（多表联查 / 推荐算法）
│   ├── server/                  # 路由注册（route.go）
│   ├── service/                 # 业务逻辑层
│   └── task/                    # 定时任务
├── log/                          # 日志文件
├── pkg/
│   ├── jwt/                     # JWT 签发 / 解析
│   ├── logging/                 # Zap 日志封装
│   ├── snowflake/               # 雪花 ID 生成
│   └── storage/                 # 本地文件存储
├── scripts/
├── test/
└── uploads/                     # 上传文件目录（已 gitignore）
```

## 🏗️ 架构设计

### 四层架构

```
Request
   ↓
Handler（HTTP层）       参数绑定 / 校验 / 错误响应
   ↓
Service（业务逻辑层）   核心逻辑 / 权限判断 / 数据组装
   ↓
GORM Gen Query         类型安全的自动生成查询（简单 CRUD）
Repository             手写复杂查询（多表 JOIN / 推荐算法）
   ↓
DAO（MySQL + Redis）
```

每个模块遵循固定目录映射：

| 层级 | 路径 | 职责 |
|------|------|------|
| 类型定义 | `api/<module>/v1/*.go` | Req / Resp 结构体 |
| HTTP 处理 | `internal/handler/<module>/` | 绑定参数、调 Service、返回响应 |
| 业务逻辑 | `internal/service/<module>/` | 查库、组装数据、返回 model 输出 |
| 输出模型 | `internal/model/<module>.go` | Handler ↔ Service 之间的传输结构 |
| 路由注册 | `internal/server/route.go` | 所有路由统一在此注册 |

### 响应格式

```go
// 成功
api.ResponseSuccess(c, resp)
// → {"code": 0, "message": "success", "data": {...}}

// 失败
api.ResponseError(c, api.CodeXxx)
// → {"code": 10001, "message": "用户不存在", "data": null}
```

## 📦 已实现模块

| 模块 | 路由前缀 | 功能 |
|------|---------|------|
| 认证 | `/api/v1/auth` | 注册、登录、刷新 Token |
| 用户 | `/api/v1/user` | 个人信息、修改资料、作品列表、点赞/收藏列表 |
| 关注 | `/api/v1/user/:id` | 关注、取消关注、关注列表、粉丝列表 |
| 视频 | `/api/v1/videos` | 发布、详情、删除、搜索 |
| 视频流 | `/api/v1/feed` | 推荐流、关注流、话题视频流 |
| 话题 | `/api/v1/topics` | 列表、详情、搜索、话题视频 |
| 评论 | `/api/v1/videos/:id/comments` | 发布、列表、删除、点赞 |
| 互动 | `/api/v1/videos/:id` | 点赞、收藏、分享 |
| 音乐 | `/api/v1/music` | 列表（热门/最新）、关键词搜索 |
| 上传 | `/api/v1/upload` | 图片 / 视频 / 音频文件上传 |
| 静态资源 | `/static` | 上传文件的 HTTP 访问 |

## 🚀 快速开始

### 1. 初始化数据库

```bash
mysql -u root -p < docs/schema.sql
```

### 2. 修改配置

```bash
cp config/config.yaml config/config.local.yaml
# 按实际环境修改 MySQL / Redis / JWT 等配置
```

### 3. 生成 GORM Gen 代码

```bash
# 连接数据库后生成（仅表结构变更时需要重新运行）
go run cmd/gen/generate.go
```

### 4. 启动服务

```bash
go run cmd/server/main.go
# 默认监听 :8089
```

### 5. 验证

```bash
# 注册
curl -X POST http://localhost:8089/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456","phone":"13800138000","code":"123456"}'

# 上传图片（需先登录获取 token）
curl -X POST http://localhost:8089/api/v1/upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/image.jpg" \
  -F "file_type=image"

# 访问上传的文件
# http://localhost:8089/static/image/xxx.jpg
```

> 开发环境短信验证码固定为 `123456`。

## ⚙️ 配置说明

`config/config.yaml` 关键配置项：

```yaml
server:
  name: "flashvid-platform"
  mode: "dev"                  # dev / release
  port: 8089

snowflake:
  start_time: "2026-07-19"     # 雪花 ID 起始时间
  machine_id: 1

jwt:
  access_secret: "..."         # access token 签名密钥
  refresh_secret: "..."        # refresh token 签名密钥
  access_expire_seconds: 3600  # access token 有效期，默认 1h
  refresh_expire_seconds: 86400  # refresh token 有效期，默认 1d

mysql:
  host: 127.0.0.1
  port: 3306
  user: "root"
  password: "..."
  dbname: "flashvid_platform"

redis:
  host: 127.0.0.1
  port: 6379
  db: 0

storage:
  local_path: "./uploads"                     # 文件保存目录
  base_url: "http://localhost:8089/static"    # 文件访问 URL 前缀
  max_image_size: 10485760    # 10 MB
  max_video_size: 524288000   # 500 MB
  max_audio_size: 52428800    # 50 MB
```

## 📝 开发规范

### 新增接口步骤

1. `api/<module>/v1/` — 定义 `XxxReq` / `XxxResp` 结构体
2. `internal/model/` — 定义 Service 输出模型（如 `XxxInfo`、`XxxListOutput`）
3. `internal/service/<module>/` — 实现业务逻辑，返回 `(*model.XxxOutput, api.ResCode, error)`
4. `internal/handler/<module>/` — 绑定参数、调 Service、返回响应
5. `internal/server/route.go` — 注册路由

### 数据访问选择

| 场景 | 用法 |
|------|------|
| 单表 CRUD、简单 WHERE | `query.XXX.WithContext(ctx).Where(...).Find()` |
| 同一字段 OR 条件 | `query.XXX.WithContext(ctx).Where(...).UnderlyingDB().Where("col1 LIKE ? OR col2 LIKE ?", v1, v2).Find(...)` |
| 多表 JOIN / 聚合统计 / 推荐算法 | `internal/repository/` 手写 SQL |

### 认证中间件

```go
// 需要登录的路由组
r.Use(middleware.Auth())

// Handler 内取当前用户 ID
userID := c.MustGet(middleware.CtxKeyUserID).(int64)
```

## 🔧 技术栈

| 组件 | 版本 / 说明 |
|------|------------|
| Web 框架 | Gin |
| ORM | GORM + GORM Gen（类型安全代码生成） |
| 配置管理 | Viper |
| 日志 | Zap |
| 认证 | JWT（Access + Refresh 双 Token） |
| ID 生成 | 雪花算法 |
| 数据库 | MySQL 8.0+ + Redis |
| 文件存储 | 本地磁盘（`pkg/storage`），`/static` 路由提供访问 |

## 📚 相关文档

- [API 设计文档](docs/api-design.md)
- [OpenAPI 3.0 规范](docs/api.yaml)
- [数据库设计文档](docs/database-design.md)

## 📄 许可证

MIT License
