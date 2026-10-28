package notification

import (
	"encoding/json"
	"fmt"
	"time"

	"flashvid-platform-gin/pkg/sse"
	notifSvc "flashvid-platform-gin/internal/service/notification"

	"github.com/gin-gonic/gin"
)

// NotificationStreamHandler SSE 推送通知流
func NotificationStreamHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 注册客户端
	client := sse.GlobalManager.Register(userID)
	defer sse.GlobalManager.Unregister(userID)

	// 发送初始未读数
	counts, err := notifSvc.GetUnreadCountsWithCache(c, userID)
	if err == nil {
		eventData := map[string]interface{}{
			"type":    "unread_count_update",
			"user_id": userID,
			"data":    counts,
		}
		sendSSEEvent(c, eventData)
	}

	// 心跳定时器（每 30 秒）
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// 超时定时器（60 秒无活动断开）
	timeout := time.NewTimer(60 * time.Second)
	defer timeout.Stop()

	ctx := c.Request.Context()

	for {
		select {
		case <-ctx.Done():
			return

		case <-client.Done:
			return

		case event := <-client.Channel:
			if err := sendSSEEvent(c, event); err != nil {
				return
			}
			timeout.Reset(60 * time.Second)

		case <-heartbeat.C:
			c.Writer.WriteString(":heartbeat\n\n")
			c.Writer.Flush()

		case <-timeout.C:
			return
		}
	}
}

func sendSSEEvent(c *gin.Context, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	if err != nil {
		return err
	}

	c.Writer.Flush()
	return nil
}
