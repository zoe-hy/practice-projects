package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // 确保已安装: go get github.com/go-sql-driver/mysql
	"github.com/segmentio/kafka-go"    // 确保已安装: go get github.com/segmentio/kafka-go@v0.4.47
)

// 定义数据结构，必须与 main.go 发送的 JSON 格式对应
type DeviceData struct {
	DeviceID string  `json:"device_id"`
	Value    float64 `json:"value"`
}

func main() {
	fmt.Println("=== 🚚 搬运工（消费者）正在启动 ===")

	// 1. 连接 MySQL (由于 3306 被占用，这里务必确认使用 3307)
	// 格式: 用户名:密码@tcp(地址:端口)/数据库名
	dsn := "root:root@tcp(127.0.0.1:3307)/iot_db?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ 数据库连接配置错误: %v", err)
	}

	// 检查数据库是否真的在线
	err = db.Ping()
	if err != nil {
		log.Fatalf("❌ 连不上 MySQL! 请确认 Docker 容器已启动且端口为 3307: %v", err)
	}
	fmt.Println("✅ 数据库连接成功！")
	defer db.Close()

	// 2. 配置 Kafka 阅读器
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		Topic:       "device_data",
		GroupID:     "db-saver-v2",     // 使用新 GroupID 确保从头开始消费
		StartOffset: kafka.FirstOffset, // 即使程序重启，也会把之前没处理的数据补回来
	})
	defer reader.Close()

	fmt.Println("👂 正在监听 Kafka 消息...")

	for {
		// 3. 从 Kafka 抓取一条消息
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("❌ 读取 Kafka 失败: %v\n", err)
			break
		}

		// 4. 解析 JSON 数据
		var data DeviceData
		err = json.Unmarshal(m.Value, &data)
		if err != nil {
			fmt.Printf("❌ 解析数据失败: %v, 原始内容: %s\n", err, string(m.Value))
			continue
		}
		if data.Value > 30 {
			// \033[31m 是红色，\033[0m 是重置颜色
			fmt.Printf("\033[31m⚠️  警告：设备 [%s] 温度异常！当前数值: %.2f\033[0m\n",
				data.DeviceID, data.Value)
		}

		// 5. 写入数据库
		query := "INSERT INTO device_logs (device_id, value) VALUES (?, ?)"
		result, err := db.Exec(query, data.DeviceID, data.Value)

		if err != nil {
			fmt.Printf("❌ 入库 SQL 报错: %v\n", err)
		} else {
			lastID, _ := result.LastInsertId()
			fmt.Printf("✅ [%s] 入库成功! 自动编号: %d, 数值: %.2f\n",
				time.Now().Format("15:04:05"), lastID, data.Value)
		}
	}
}
