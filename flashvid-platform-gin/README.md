# FlashVid Platform - 后端服务

基于 Gin + GORM Gen 的短视频平台后端服务，采用清晰的分层架构。

基于 [Qimi 的 gin-base-layout](https://github.com/q1mi/gin-base-layout) 脚手架改造。

## 📁 项目结构

```bash
├── api/                          # API 定义
├── cmd/
│   ├── gen/                     # GORM Gen 代码生成工具
│   └── server/                  # 服务启动入口
├── config/                       # 配置文件
├── docs/                         # 文档
├── internal/
│   ├── conf/                    # 配置结构
│   ├── dao/                     # 数据库初始化层
│   │   ├── mysql.go
│   │   ├── redis.go
│   │   └── query/              # GORM Gen 自动生成
│   ├── repository/              # 数据访问层（复杂查询）⭐
│   │   ├── user_repo.go
│   │   ├── video_repo.go
│   │   ├── feed_repo.go
│   │   └── stat_repo.go
│   ├── service/                 # 业务逻辑层
│   ├── handler/                 # HTTP 处理层
│   ├── middleware/              # 中间件
│   ├── model/                   # 数据模型（Gen 生成）
│   ├── server/                  # 服务器配置
│   └── task/                    # 定时任务
├── log/                          # 日志文件
├── pkg/                          # 公共包
│   ├── jwt/
│   ├── logging/
│   └── snowflake/
├── scripts/                      # 脚本文件
└── test/                         # 测试文件
```

## 🏗️ 架构设计

### 分层架构

```
Handler (HTTP层) → Service (业务逻辑层)
                       ↓
         ┌─────────────┴─────────────┐
         ↓                           ↓
    Query (Gen)              Repository
    简单CRUD 80%              复杂查询 20%
         ↓                           ↓
         └─────────────┬─────────────┘
                       ↓
                  DAO (数据库初始化)
                       ↓
                  MySQL/Redis
```

### 数据访问层

- **Query (GORM Gen)**: 自动生成的类型安全查询，处理 80% 的简单 CRUD
- **Repository**: 手写的复杂查询（多表联查、统计、算法）

## 🚀 快速开始

### 1. 初始化数据库

```bash
# 创建数据库并导入表结构
mysql -u root -p < ../docs/schema.sql
```

### 2. 生成 GORM Gen 代码

```bash
# 运行代码生成工具
go run cmd/gen/main.go
```

### 3. 配置环境变量

```bash
# 修改配置文件
vim config/config.yaml
```

### 4. 启动服务

```bash
# 开发环境
go run cmd/server/main.go
```

## 📝 开发规范

### 数据访问层使用原则

**简单查询用 Gen：**
```go
// 80% 的场景
u := dao.Q.User
user, err := u.Where(u.ID.Eq(userID)).First()
```

**复杂查询用 Repository：**
```go
// 20% 的场景：多表联查、统计、算法
videos, err := videoRepo.GetRecommendVideos(userID, 20)
```

### Repository 使用场景

在以下情况需要使用 Repository 层：

- ✅ 多表联查（JOIN 2个以上表）
- ✅ 聚合统计（GROUP BY, HAVING, COUNT, SUM）
- ✅ 复杂算法（推荐算法、权重计算）
- ✅ 地理位置查询（Haversine 公式）
- ✅ 性能优化（原生 SQL + 索引优化）

## 📚 相关文档

- [API 设计文档](../docs/api-design.md)
- [数据库设计文档](../docs/database-design.md)
- [OpenAPI 规范](../docs/api.yaml)

## 🔧 技术栈

- **Web 框架**: Gin
- **ORM**: GORM + GORM Gen
- **配置管理**: Viper
- **日志**: Zap
- **认证**: JWT
- **数据库**: MySQL 8.0+ + Redis

## 📄 许可证

MIT License