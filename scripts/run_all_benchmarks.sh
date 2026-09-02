#!/bin/bash

# FlashVid 综合压测脚本
# 用法: ./run_all_benchmarks.sh

echo "======================================"
echo "FlashVid 平台综合压测"
echo "======================================"
echo ""

# 检查服务器是否运行
echo "检查服务器状态..."
if ! curl -s http://localhost:8089/health > /dev/null 2>&1; then
    echo "❌ 服务器未启动，请先启动服务器"
    echo "   cd flashvid-platform-gin && go run cmd/server/main.go"
    exit 1
fi
echo "✅ 服务器运行中"
echo ""

# 创建结果目录
mkdir -p benchmark_results
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULT_DIR="benchmark_results/${TIMESTAMP}"
mkdir -p "$RESULT_DIR"

echo "======================================"
echo "1/4 点赞/收藏压测"
echo "======================================"
go run benchmark_interaction.go 2>&1 | tee "${RESULT_DIR}/interaction.log"
echo ""

sleep 3

echo "======================================"
echo "2/4 Feed 流压测"
echo "======================================"
go run benchmark_feed.go 2>&1 | tee "${RESULT_DIR}/feed.log"
echo ""

sleep 3

echo "======================================"
echo "3/4 视频详情压测"
echo "======================================"
go run benchmark_video_detail.go 2>&1 | tee "${RESULT_DIR}/video_detail.log"
echo ""

sleep 3

echo "======================================"
echo "4/4 话题热榜压测"
echo "======================================"
go run benchmark_hot_topics.go 2>&1 | tee "${RESULT_DIR}/hot_topics.log"
echo ""

echo "======================================"
echo "压测完成！"
echo "======================================"
echo "结果已保存到: ${RESULT_DIR}/"
echo ""
echo "下一步："
echo "1. 查看各个日志文件获取详细数据"
echo "2. 将关键指标填写到 docs/high-concurrency-roadmap.md"
echo "3. 运行 ./generate_benchmark_report.sh 生成报告"
