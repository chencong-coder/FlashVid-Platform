package main

import (
	"encoding/json"
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
	videoID = 1
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 500,
		IdleConnTimeout:     90 * time.Second,
	},
}

type BenchmarkResult struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	CacheHits       int64
	CacheMisses     int64
	TotalDuration   time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	QPS             float64
	CacheHitRate    float64
}

type VideoResponse struct {
	Code int `json:"code"`
	Data struct {
		Video map[string]interface{} `json:"video"`
	} `json:"data"`
}

// 获取视频详情
func getVideoDetail(videoID int, token string, cacheHits *int64, cacheMisses *int64) (time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/videos/%d", baseURL, videoID), nil)
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

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return duration, fmt.Errorf("status: %d, body: %s", resp.StatusCode, string(body))
	}

	var videoResp VideoResponse
	if err := json.Unmarshal(body, &videoResp); err != nil {
		return duration, fmt.Errorf("parse response failed: %w", err)
	}

	if videoResp.Code != 0 {
		return duration, fmt.Errorf("business error: code=%d", videoResp.Code)
	}

	// 简单判断缓存命中率（延迟 < 10ms 认为是缓存命中）
	if duration < 10*time.Millisecond {
		atomic.AddInt64(cacheHits, 1)
	} else {
		atomic.AddInt64(cacheMisses, 1)
	}

	return duration, nil
}

func benchmark(name string, concurrency int, requests int, videoID int, token string) BenchmarkResult {
	var (
		successCount int64
		failedCount  int64
		cacheHits    int64
		cacheMisses  int64
		totalLatency int64
		minLatency   = int64(1<<63 - 1)
		maxLatency   int64
		wg           sync.WaitGroup
		requestsChan = make(chan struct{}, requests)
		errorsMu     sync.Mutex
		errorCounts  = make(map[string]int)
	)

	for i := 0; i < requests; i++ {
		requestsChan <- struct{}{}
	}
	close(requestsChan)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestsChan {
				latency, err := getVideoDetail(videoID, token, &cacheHits, &cacheMisses)
				if err != nil {
					atomic.AddInt64(&failedCount, 1)
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

				for {
					oldMin := atomic.LoadInt64(&minLatency)
					if int64(latency) >= oldMin {
						break
					}
					if atomic.CompareAndSwapInt64(&minLatency, oldMin, int64(latency)) {
						break
					}
				}

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
		CacheHits:       cacheHits,
		CacheMisses:     cacheMisses,
		TotalDuration:   duration,
		MinLatency:      time.Duration(minLatency),
		MaxLatency:      time.Duration(maxLatency),
		QPS:             float64(successCount) / duration.Seconds(),
	}

	if successCount > 0 {
		result.AvgLatency = time.Duration(totalLatency / successCount)
		result.CacheHitRate = float64(cacheHits) / float64(successCount) * 100
	}

	return result
}

func printResult(name string, result BenchmarkResult) {
	fmt.Printf("\n========== %s 压测结果 ==========\n", name)
	fmt.Printf("总请求数:     %d\n", result.TotalRequests)
	fmt.Printf("成功请求数:   %d\n", result.SuccessRequests)
	fmt.Printf("失败请求数:   %d\n", result.FailedRequests)
	fmt.Printf("缓存命中:     %d\n", result.CacheHits)
	fmt.Printf("缓存未命中:   %d\n", result.CacheMisses)
	fmt.Printf("缓存命中率:   %.2f%%\n", result.CacheHitRate)
	fmt.Printf("总耗时:       %v\n", result.TotalDuration)
	fmt.Printf("QPS:          %.2f req/s\n", result.QPS)
	fmt.Printf("平均延迟:     %v\n", result.AvgLatency)
	fmt.Printf("最小延迟:     %v\n", result.MinLatency)
	fmt.Printf("最大延迟:     %v\n", result.MaxLatency)
	fmt.Printf("成功率:       %.2f%%\n", float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	fmt.Println("=====================================")
}

func main() {
	fmt.Println("开始视频详情压测...")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Video ID: %d\n", videoID)

	if token == "your_jwt_token_here" {
		fmt.Println("\n❌ 请先替换脚本中的 token 和 videoID")
		return
	}

	// 预热缓存：先请求 3 次
	fmt.Println("\n预热缓存...")
	for i := 0; i < 3; i++ {
		getVideoDetail(videoID, token, new(int64), new(int64))
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("预热完成")

	time.Sleep(1 * time.Second)

	// 场景 1: 100 并发，1000 次请求（测试缓存命中率）
	fmt.Println("\n场景 1: 100 并发视频详情 (1000 次请求)")
	result := benchmark("视频详情", 100, 1000, videoID, token)
	printResult("视频详情", result)

	// 汇总
	fmt.Println("\n========== 性能总结 ==========")
	fmt.Printf("QPS:          %.2f req/s\n", result.QPS)
	fmt.Printf("平均延迟:     %v\n", result.AvgLatency)
	fmt.Printf("缓存命中率:   %.2f%%\n", result.CacheHitRate)
	fmt.Println("==============================")
}
