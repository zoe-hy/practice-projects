package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// 定义和 API 一样的数据结构
type DeviceData struct {
	DeviceID string  `json:"device_id"`
	Value    float64 `json:"value"`
}

func main() {
	url := "http://localhost:8080/upload"
	totalRequests := 1000 // 总请求数
	concurrentLimit := 50 // 并发限制（同时有多少个“人”在发）

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrentLimit)

	start := time.Now()

	fmt.Printf("🚀 开始压测：发送 %d 个请求，并发数 %d...\n", totalRequests, concurrentLimit)

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 占用一个并发名额
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 构造随机数据
			data := DeviceData{
				DeviceID: fmt.Sprintf("Sensor_%03d", id),
				Value:    20.0 + (float64(id) * 0.1),
			}
			jsonData, _ := json.Marshal(data)

			// 发送 POST 请求
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("❌ 请求 %d 失败: %v\n", id, err)
				return
			}
			resp.Body.Close()
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Println("\n======================================")
	fmt.Printf("🏁 压测完成！\n")
	fmt.Printf("⏱️  总耗时: %v\n", duration)
	fmt.Printf("📈 平均速度: %.2f 请求/秒\n", float64(totalRequests)/duration.Seconds())
	fmt.Println("======================================")
}
