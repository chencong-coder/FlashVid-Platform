package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"flashvid-platform-gin/internal/conf"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/mq"
	"flashvid-platform-gin/internal/server"
	"flashvid-platform-gin/internal/task"
	"flashvid-platform-gin/pkg/jwt"
	"flashvid-platform-gin/pkg/logging"
	"flashvid-platform-gin/pkg/snowflake"
	"flashvid-platform-gin/pkg/storage"
)

var confPath = flag.String("conf", "./config/config.yaml", "配置文件路径")

func main() {
	// 加载配置
	flag.Parse()
	cfg := conf.Load(*confPath)

	// 初始化日志
	logger, err := logging.NewLogger(cfg)
	if err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	defer logger.Sync()

	dao.MustInitMySQL(cfg)  // 初始化 MySQL 连接
	dao.MustInitRedis(cfg)  // 初始化 Redis
	jwt.MustInit(cfg)       // 初始化 jwt
	snowflake.MustInit(cfg) // 初始化 snowflake
	storage.MustInit(cfg)   // 初始化本地文件存储

	// 初始化 RabbitMQ（仅用于发送消息，不声明队列，由 Worker 负责）
	mq.MustInitRabbitMQ(cfg)
	defer mq.Close()

	// 等待 RabbitMQ 完全就绪
	time.Sleep(1 * time.Second)

	// 启动定时任务：Redis 统计数据同步到 MySQL（每 10 秒）
	go func() {
		taskCtx := context.Background()
		task.SyncVideoStatsFromRedis(taskCtx, 10*time.Second)
	}()

	// 初始化路由
	r := server.SetupRoutes(cfg)
	// 启动 API 服务
	err = r.Run(fmt.Sprintf(":%d", cfg.GetInt("server.port")))
	if err != nil {
		panic(err)
	}
}
