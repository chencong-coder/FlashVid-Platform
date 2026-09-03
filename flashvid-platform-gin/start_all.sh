#!/bin/bash

echo "Starting FlashVid Platform..."

echo ""
echo "[1/2] Starting API Server..."
go run cmd/server/api/main.go &
API_PID=$!

sleep 2

echo "[2/2] Starting Worker (Consumers)..."
go run cmd/server/worker/main.go &
WORKER_PID=$!

echo ""
echo "All services started!"
echo "- API Server (PID: $API_PID): http://localhost:8080"
echo "- Worker (PID: $WORKER_PID): Running in background"
echo ""
echo "Press Ctrl+C to stop all services"

# 优雅关闭
trap "kill $API_PID $WORKER_PID; exit" SIGINT SIGTERM

wait
