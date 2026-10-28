package task

import (
	"context"
	"fmt"
)

// InitAllHot 初始化所有热榜数据
func InitAllHot(ctx context.Context) error {
	// 初始化话题热榜
	fmt.Println("正在初始化话题热榜...")
	if err := InitTopicHot(ctx); err != nil {
		return fmt.Errorf("话题热榜初始化失败: %w", err)
	}
	fmt.Println("✓ 话题热榜初始化成功")

	// 初始化视频热榜
	fmt.Println("正在初始化视频热榜...")
	if err := InitVideoHot(ctx); err != nil {
		return fmt.Errorf("视频热榜初始化失败: %w", err)
	}
	fmt.Println("✓ 视频热榜初始化成功")

	fmt.Println("\n🎉 所有热榜初始化完成！")
	return nil
}
