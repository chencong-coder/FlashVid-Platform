# Docker 部署指南

## 🚀 快速开始

### 前置要求
- Docker 20.10+
- Docker Compose 2.0+

### 一键启动

**Linux/Mac:**
```bash
chmod +x start.sh
./start.sh
```

**Windows:**
```bash
start.bat
```

或者手动启动：
```bash
docker-compose up -d --build
```

---

## 📦 服务列表

| 服务 | 端口 | 说明 |
|------|------|------|
| **flashvid-api** | 8089 | API Server（HTTP 接口） |
| **flashvid-worker** | - | Worker 进程（MQ 消费者） |
| **mysql** | 3306 | MySQL 8.0 数据库 |
| **redis** | 6379 | Redis 7.0 缓存 |
| **rabbitmq** | 5672, 15672 | RabbitMQ 消息队列 + 管理界面 |

---

## 🔧 常用命令

### 查看服务状态
```bash
docker-compose ps
```

### 查看日志
```bash
# 所有服务
docker-compose logs -f

# 单个服务
docker-compose logs -f api
docker-compose logs -f worker
docker-compose logs -f mysql
```

### 重启服务
```bash
# 重启所有服务
docker-compose restart

# 重启单个服务
docker-compose restart api
docker-compose restart worker
```

### 停止服务
```bash
docker-compose down
```

### 停止并删除数据卷
```bash
docker-compose down -v
```

---

## 🗄️ 数据持久化

数据卷挂载：
- **MySQL 数据**：`mysql_data` 卷 + `./flashvid-platform-gin/docs/sql` 初始化脚本
- **Redis 数据**：`redis_data` 卷
- **RabbitMQ 数据**：`rabbitmq_data` 卷
- **上传文件**：`./flashvid-platform-gin/uploads`
- **日志文件**：`./flashvid-platform-gin/log`

---

## 🔍 访问地址

- **API Server**: http://localhost:8089
- **RabbitMQ 管理界面**: http://localhost:15672
  - 用户名：`admin`
  - 密码：`password`

---

## 🐛 故障排查

### 服务启动失败

**1. 检查端口占用**
```bash
# Linux/Mac
lsof -i :8089
lsof -i :3306
lsof -i :6379
lsof -i :5672

# Windows
netstat -ano | findstr :8089
netstat -ano | findstr :3306
```

**2. 查看容器日志**
```bash
docker-compose logs api
docker-compose logs worker
```

**3. 检查健康检查状态**
```bash
docker-compose ps
# 如果状态是 "unhealthy"，说明服务未就绪
```

### MySQL 连接失败

等待 MySQL 完全启动（健康检查通过后）：
```bash
docker-compose logs mysql | grep "ready for connections"
```

### RabbitMQ 连接失败

等待 RabbitMQ 完全启动：
```bash
docker-compose logs rabbitmq | grep "Server startup complete"
```

---

## 🏗️ 架构说明

### 多阶段构建
- **Builder 阶段**：编译 Go 二进制文件（API + Worker）
- **API 镜像**：运行 API Server
- **Worker 镜像**：运行 Worker 消费者

### 配置文件
- 本地开发：`config/config.yaml`（localhost）
- Docker 部署：`config/config.docker.yaml`（服务名）

### 健康检查
所有依赖服务（MySQL、Redis、RabbitMQ）都配置了健康检查，API 和 Worker 会等待依赖服务就绪后再启动。

---

## 📝 开发模式

如果需要本地开发（不使用 Docker）：

1. 启动依赖服务：
```bash
docker-compose up -d mysql redis rabbitmq
```

2. 本地运行 Go 程序：
```bash
# API Server
go run cmd/server/api/main.go

# Worker
go run cmd/server/worker/main.go
```

---

## 🔄 更新部署

```bash
# 拉取最新代码
git pull

# 重新构建并启动
docker-compose up -d --build

# 或使用启动脚本
./start.sh  # Linux/Mac
start.bat   # Windows
```
