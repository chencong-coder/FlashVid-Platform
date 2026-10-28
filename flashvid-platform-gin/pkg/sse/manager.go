package sse

import (
	"fmt"
	"sync"
	"time"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Client struct {
	UserID  int64
	Channel chan Event
	Done    chan struct{}
}

type Manager struct {
	clients sync.Map // map[int64]*Client
}

var GlobalManager *Manager

// InitManager 初始化全局 SSE 管理器
func InitManager() {
	GlobalManager = &Manager{}
}

// Register 注册新客户端连接
func (m *Manager) Register(userID int64) *Client {
	client := &Client{
		UserID:  userID,
		Channel: make(chan Event, 100),
		Done:    make(chan struct{}),
	}

	m.clients.Store(userID, client)
	return client
}

// Unregister 注销客户端连接
func (m *Manager) Unregister(userID int64) {
	if val, ok := m.clients.LoadAndDelete(userID); ok {
		client := val.(*Client)
		close(client.Done)
		close(client.Channel)
	}
}

// Send 向指定用户发送事件
func (m *Manager) Send(userID int64, event Event) error {
	val, ok := m.clients.Load(userID)
	if !ok {
		return fmt.Errorf("client not found: %d", userID)
	}

	client := val.(*Client)

	select {
	case client.Channel <- event:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send timeout")
	}
}

// GetOnlineCount 获取在线用户数
func (m *Manager) GetOnlineCount() int {
	count := 0
	m.clients.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
