@echo off
REM FlashVid 综合压测脚本 (Windows)
REM 用法: run_all_benchmarks.bat

echo ======================================
echo FlashVid 平台综合压测
echo ======================================
echo.

REM 检查服务器是否运行
echo 检查服务器状态...
curl -s http://localhost:8089/health >nul 2>&1
if errorlevel 1 (
    echo ❌ 服务器未启动，请先启动服务器
    echo    cd flashvid-platform-gin ^&^& go run cmd/server/main.go
    exit /b 1
)
echo ✅ 服务器运行中
echo.

REM 创建结果目录
if not exist benchmark_results mkdir benchmark_results
set TIMESTAMP=%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set TIMESTAMP=%TIMESTAMP: =0%
set RESULT_DIR=benchmark_results\%TIMESTAMP%
mkdir "%RESULT_DIR%"

echo ======================================
echo 1/4 点赞/收藏压测
echo ======================================
go run benchmark_interaction.go 2>&1 | tee "%RESULT_DIR%\interaction.log"
echo.

timeout /t 3 >nul

echo ======================================
echo 2/4 Feed 流压测
echo ======================================
go run benchmark_feed.go 2>&1 | tee "%RESULT_DIR%\feed.log"
echo.

timeout /t 3 >nul

echo ======================================
echo 3/4 视频详情压测
echo ======================================
go run benchmark_video_detail.go 2>&1 | tee "%RESULT_DIR%\video_detail.log"
echo.

timeout /t 3 >nul

echo ======================================
echo 4/4 话题热榜压测
echo ======================================
go run benchmark_hot_topics.go 2>&1 | tee "%RESULT_DIR%\hot_topics.log"
echo.

echo ======================================
echo 压测完成！
echo ======================================
echo 结果已保存到: %RESULT_DIR%\
echo.
echo 下一步：
echo 1. 查看各个日志文件获取详细数据
echo 2. 将关键指标填写到 docs\high-concurrency-roadmap.md
echo 3. 生成性能对比报告
pause
