package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	baseURL = "http://localhost:8089"
	token   = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOjI3OTg1OTYwNDI3ODQ3NjksInVzZXJuYW1lIjoiYWxpY2VfbGluIiwiaXNzIjoibGl3ZW56aG91LmNvbSIsInN1YiI6InN1bmZsb3dlciIsImV4cCI6MTc4ODM1NDU4MiwibmJmIjoxNzg4MzUwOTgyLCJpYXQiOjE3ODgzNTA5ODJ9.V6h4ztBAWFLr3aws8D_KLT-VqNf_1Abp_QvOAqx2NCE"
	videoID = 1 // 先用 1 测试，如果不存在再换
)

// 全局 HTTP 客户端，复用连接池
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 500,
		IdleConnTimeout:     90 * time.Second,
	},
}

// 压测结果统计
type BenchmarkResult struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalDuration   time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	QPS             float64
}

// 发送点赞请求
func likeVideo(videoID int, token string) (time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/videos/%d/like", baseURL, videoID), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return duration, fmt.Errorf("status: %d, body: %s", resp.StatusCode, string(body))
	}

	return duration, nil
}

// 发送取消点赞请求
func unlikeVideo(videoID int, token string) (time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/videos/%d/like", baseURL, videoID), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return duration, fmt.Errorf("status: %d, body: %s", resp.StatusCode, string(body))
	}

	return duration, nil
}

// 发送收藏请求
func favoriteVideo(videoID int, token string) (time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/videos/%d/favorite", baseURL, videoID), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return duration, fmt.Errorf("status: %d, body: %s", resp.StatusCode, string(body))
	}

	return duration, nil
}

// 发送取消收藏请求
func unfavoriteVideo(videoID int, token string) (time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/videos/%d/favorite", baseURL, videoID), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return duration, fmt.Errorf("status: %d, body: %s", resp.StatusCode, string(body))
	}

	return duration, nil
}

// 并发压测
func benchmark(name string, concurrency int, requests int, reqFunc func() (time.Duration, error)) BenchmarkResult {
	var (
		successCount int64
		failedCount  int64
		totalLatency int64
		minLatency   = int64(1<<63 - 1)
		maxLatency   int64
		wg           sync.WaitGroup
		requestsChan = make(chan struct{}, requests)
		errorsMu     sync.Mutex
		errorCounts  = make(map[string]int)
	)

	// 填充请求队列
	for i := 0; i < requests; i++ {
		requestsChan <- struct{}{}
	}
	close(requestsChan)

	start := time.Now()

	// 启动并发 workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestsChan {
				latency, err := reqFunc()
				if err != nil {
					atomic.AddInt64(&failedCount, 1)
					// 收集错误信息
					errorsMu.Lock()
					errMsg := err.Error()
					if len(errMsg) > 100 {
						errMsg = errMsg[:100]
					}
					errorCounts[errMsg]++
					errorsMu.Unlock()
					continue
				}

				atomic.AddInt64(&successCount, 1)
				atomic.AddInt64(&totalLatency, int64(latency))

				// 更新最小延迟
				for {
					oldMin := atomic.LoadInt64(&minLatency)
					if int64(latency) >= oldMin {
						break
					}
					if atomic.CompareAndSwapInt64(&minLatency, oldMin, int64(latency)) {
						break
					}
				}

				// 更新最大延迟
				for {
					oldMax := atomic.LoadInt64(&maxLatency)
					if int64(latency) <= oldMax {
						break
					}
					if atomic.CompareAndSwapInt64(&maxLatency, oldMax, int64(latency)) {
						break
					}
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	// 打印错误统计
	if len(errorCounts) > 0 {
		fmt.Printf("\n错误统计:\n")
		for errMsg, count := range errorCounts {
			fmt.Printf("  [%d 次] %s\n", count, errMsg)
		}
	}

	result := BenchmarkResult{
		TotalRequests:   int64(requests),
		SuccessRequests: successCount,
		FailedRequests:  failedCount,
		TotalDuration:   duration,
		MinLatency:      time.Duration(minLatency),
		MaxLatency:      time.Duration(maxLatency),
		QPS:             float64(successCount) / duration.Seconds(),
	}

	if successCount > 0 {
		result.AvgLatency = time.Duration(totalLatency / successCount)
	}

	return result
}

func printResult(name string, result BenchmarkResult) {
	fmt.Printf("\n========== %s 压测结果 ==========\n", name)
	fmt.Printf("总请求数:     %d\n", result.TotalRequests)
	fmt.Printf("成功请求数:   %d\n", result.SuccessRequests)
	fmt.Printf("失败请求数:   %d\n", result.FailedRequests)
	fmt.Printf("总耗时:       %v\n", result.TotalDuration)
	fmt.Printf("QPS:          %.2f req/s\n", result.QPS)
	fmt.Printf("平均延迟:     %v\n", result.AvgLatency)
	fmt.Printf("最小延迟:     %v\n", result.MinLatency)
	fmt.Printf("最大延迟:     %v\n", result.MaxLatency)
	fmt.Printf("成功率:       %.2f%%\n", float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	fmt.Println("=====================================")
}

func main() {
	fmt.Println("开始压测...")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Video ID: %d\n", videoID)

	// 检查 token
	if token == "your_jwt_token_here" {
		fmt.Println("\n❌ 请先替换脚本中的 token 和 videoID")
		return
	}

	// 场景 1: 100 并发，1000 次点赞请求
	fmt.Println("\n场景 1: 100 并发点赞 (1000 次请求)")
	result1 := benchmark("点赞", 100, 1000, func() (time.Duration, error) {
		return likeVideo(videoID, token)
	})
	printResult("点赞", result1)

	time.Sleep(2 * time.Second)

	// 场景 2: 100 并发，1000 次取消点赞请求
	fmt.Println("\n场景 2: 100 并发取消点赞 (1000 次请求)")
	result2 := benchmark("取消点赞", 100, 1000, func() (time.Duration, error) {
		return unlikeVideo(videoID, token)
	})
	printResult("取消点赞", result2)

	time.Sleep(2 * time.Second)

	// 场景 3: 100 并发，1000 次收藏请求
	fmt.Println("\n场景 3: 100 并发收藏 (1000 次请求)")
	result3 := benchmark("收藏", 100, 1000, func() (time.Duration, error) {
		return favoriteVideo(videoID, token)
	})
	printResult("收藏", result3)

	time.Sleep(2 * time.Second)

	// 场景 4: 100 并发，1000 次取消收藏请求
	fmt.Println("\n场景 4: 100 并发取消收藏 (1000 次请求)")
	result4 := benchmark("取消收藏", 100, 1000, func() (time.Duration, error) {
		return unfavoriteVideo(videoID, token)
	})
	printResult("取消收藏", result4)

	// 汇总对比
	fmt.Println("\n========== 汇总对比 ==========")
	fmt.Printf("点赞 QPS:         %.2f req/s\n", result1.QPS)
	fmt.Printf("取消点赞 QPS:     %.2f req/s\n", result2.QPS)
	fmt.Printf("收藏 QPS:         %.2f req/s\n", result3.QPS)
	fmt.Printf("取消收藏 QPS:     %.2f req/s\n", result4.QPS)
	fmt.Printf("点赞平均延迟:     %v\n", result1.AvgLatency)
	fmt.Printf("取消点赞平均延迟: %v\n", result2.AvgLatency)
	fmt.Printf("收藏平均延迟:     %v\n", result3.AvgLatency)
	fmt.Printf("取消收藏平均延迟: %v\n", result4.AvgLatency)
	fmt.Println("==============================")
}
