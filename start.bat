@echo off
chcp 65001 >nul
echo 🚀 启动 FlashVid 平台...
echo.

REM 检查 Docker
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker 未安装，请先安装 Docker Desktop
    pause
    exit /b 1
)

REM 检查 Docker Compose
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    docker compose version >nul 2>&1
    if %errorlevel% neq 0 (
        echo ❌ Docker Compose 未安装
        pause
        exit /b 1
    )
)

REM 停止并删除旧容器
echo 🧹 清理旧容器...
docker-compose down

REM 构建并启动所有服务
echo 🔨 构建镜像并启动服务...
docker-compose up -d --build

REM 等待服务启动
echo ⏳ 等待服务启动...
timeout /t 10 /nobreak >nul

REM 检查服务状态
echo.
echo 📊 服务状态：
docker-compose ps

echo.
echo ✅ FlashVid 平台启动完成！
echo.
echo 📌 访问地址：
echo   - 前端应用: http://localhost
echo   - API Server: http://localhost:8089
echo   - RabbitMQ Management: http://localhost:15672 (admin/password)
echo.
echo 📝 查看日志：
echo   - 前端: docker-compose logs -f frontend
echo   - API Server: docker-compose logs -f api
echo   - Worker: docker-compose logs -f worker
echo   - 所有服务: docker-compose logs -f
echo.
echo 🛑 停止服务: docker-compose down
echo.
pause
