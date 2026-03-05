package main

import (
	"fmt"

	"iot_platform/internal/config"
	"iot_platform/internal/consumer"
)

func main() {
	fmt.Println("启动消费者（兼容入口），内部实现位于 internal/consumer")
	cfg := config.Load()
	if err := consumer.RunConsumer(cfg); err != nil {
		fmt.Printf("消费者退出: %v\n", err)
	}
}
