package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flashvid-platform-gin/internal/conf"
	"flashvid-platform-gin/internal/consumer"
	"flashvid-platform-gin/internal/dao"
	"flashvid-platform-gin/internal/mq"
	"flashvid-platform-gin/pkg/logging"
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

	logger.Info("Worker starting...")

	dao.MustInitMySQL(cfg) // 初始化 MySQL 连接
	dao.MustInitRedis(cfg) // 初始化 Redis

	// 初始化 RabbitMQ
	mq.MustInitRabbitMQ(cfg)
	mq.MustDeclareInfrastructure()
	defer mq.Close()

	// 等待 RabbitMQ 完全就绪
	time.Sleep(1 * time.Second)

	// 启动 RabbitMQ 消费者：视频发布热度初始化
	go consumer.ConsumeVideoHotrank()
	logger.Info("Started consumer: ConsumeVideoHotrank")

	// 启动 RabbitMQ 消费者：视频发布 Feed 推送
	go consumer.ConsumeVideoFeed()
	logger.Info("Started consumer: ConsumeVideoFeed")

	// 启动 RabbitMQ 消费者：热度更新（点赞/收藏/播放触发）
	go consumer.ConsumeHotrankUpdate()
	logger.Info("Started consumer: ConsumeHotrankUpdate")

	// 启动 RabbitMQ 消费者：通知创建（点赞/收藏触发）
	go consumer.ConsumeNotification()
	logger.Info("Started consumer: ConsumeNotification")

	logger.Info("All consumers started, worker is running...")

	// 等待中断信号优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Worker shutting down...")
}
