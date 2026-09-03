#!/bin/bash

echo "🚀 启动 FlashVid 平台..."

# 检查 Docker 和 Docker Compose
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 停止并删除旧容器
echo "🧹 清理旧容器..."
docker-compose down

# 构建并启动所有服务
echo "🔨 构建镜像并启动服务..."
docker-compose up -d --build

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo ""
echo "📊 服务状态："
docker-compose ps

echo ""
echo "✅ FlashVid 平台启动完成！"
echo ""
echo "📌 访问地址："
echo "  - 前端应用: http://localhost"
echo "  - API Server: http://localhost:8089"
echo "  - RabbitMQ Management: http://localhost:15672 (admin/password)"
echo ""
echo "📝 查看日志："
echo "  - 前端: docker-compose logs -f frontend"
echo "  - API Server: docker-compose logs -f api"
echo "  - Worker: docker-compose logs -f worker"
echo "  - 所有服务: docker-compose logs -f"
echo ""
echo "🛑 停止服务: docker-compose down"
