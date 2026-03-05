package main

import (
	"fmt"

	"iot_platform/internal/config"
	"iot_platform/internal/producer"
)

func main() {
	fmt.Println("启动生产者（兼容入口），内部实现位于 internal/producer")
	cfg := config.Load()
	if err := producer.RunProducer(cfg); err != nil {
		fmt.Printf("服务器退出: %v\n", err)
	}
}
