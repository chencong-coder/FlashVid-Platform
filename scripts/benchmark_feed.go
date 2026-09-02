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
	TotalDuration   time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	QPS             float64
}

type FeedResponse struct {
	Code int `json:"code"`
	Data struct {
		Videos []map[string]interface{} `json:"videos"`
	} `json:"data"`
}

// 获取推荐 Feed
func getRecommendFeed(token string) (time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/feed/recommend?limit=20", baseURL), nil)
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

	var feedResp FeedResponse
	if err := json.Unmarshal(body, &feedResp); err != nil {
		return duration, fmt.Errorf("parse response failed: %w", err)
	}

	if feedResp.Code != 0 {
		return duration, fmt.Errorf("business error: code=%d", feedResp.Code)
	}

	return duration, nil
}

// 获取关注 Feed
func getFollowFeed(token string) (time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/feed/follow?limit=20", baseURL), nil)
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

	var feedResp FeedResponse
	if err := json.Unmarshal(body, &feedResp); err != nil {
		return duration, fmt.Errorf("parse response failed: %w", err)
	}

	if feedResp.Code != 0 {
		return duration, fmt.Errorf("business error: code=%d", feedResp.Code)
	}

	return duration, nil
}

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
				latency, err := reqFunc()
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
	fmt.Println("开始 Feed 流压测...")
	fmt.Printf("Base URL: %s\n", baseURL)

	if token == "your_jwt_token_here" {
		fmt.Println("\n❌ 请先替换脚本中的 token")
		return
	}

	// 场景 1: 推荐 Feed
	fmt.Println("\n场景 1: 100 并发推荐 Feed (1000 次请求)")
	result1 := benchmark("推荐Feed", 100, 1000, func() (time.Duration, error) {
		return getRecommendFeed(token)
	})
	printResult("推荐Feed", result1)

	time.Sleep(2 * time.Second)

	// 场景 2: 关注 Feed
	fmt.Println("\n场景 2: 100 并发关注 Feed (1000 次请求)")
	result2 := benchmark("关注Feed", 100, 1000, func() (time.Duration, error) {
		return getFollowFeed(token)
	})
	printResult("关注Feed", result2)

	// 汇总
	fmt.Println("\n========== 汇总对比 ==========")
	fmt.Printf("推荐Feed QPS:      %.2f req/s\n", result1.QPS)
	fmt.Printf("关注Feed QPS:      %.2f req/s\n", result2.QPS)
	fmt.Printf("推荐Feed 平均延迟: %v\n", result1.AvgLatency)
	fmt.Printf("关注Feed 平均延迟: %v\n", result2.AvgLatency)
	fmt.Println("==============================")
}
