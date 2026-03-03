package main

import (
	"context" // undefined: context
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql" // 匿名导入 MySQL 驱动
	"github.com/segmentio/kafka-go"
)

// 1. 定义数据的“形状”：设备 ID 和 传感器数值
type DeviceData struct {
	DeviceID string  `json:"device_id"`
	Value    float64 `json:"value"`
}

// 2. 初始化 Kafka 写入器 (生产者)
var kafkaWriter = &kafka.Writer{
	Addr:         kafka.TCP("localhost:9092"),
	Topic:        "device_data",
	Balancer:     &kafka.LeastBytes{},
	WriteTimeout: 10 * time.Second, // 设置超时，防止卡死
}

func handleData(w http.ResponseWriter, r *http.Request) {
	// 只允许 POST 请求（模拟设备提交数据）
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	// var data DeviceData
	// // 2. 解析从设备发来的 JSON 数据
	// err := json.NewDecoder(r.Body).Decode(&data)
	// if err != nil {
	// 	http.Error(w, "数据格式错误", http.StatusBadRequest)
	// 	return
	// }

	// // 3. 打印接收到的数据（后期这里会改成发往 Kafka）
	// fmt.Printf("收到设备 [%s] 的数据: %.2f\n", data.DeviceID, data.Value)

	// // 4. 给设备回一个“收到”的消息
	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte("数据已接收"))

	//使用kafka
	var data DeviceData
	json.NewDecoder(r.Body).Decode(&data)
	msg, _ := json.Marshal(data)
	// 写入kafka中
	err := kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Value: msg,
		},
	)

	if err != nil {
		fmt.Println("发送到 Kafka 失败:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Kafka 写入失败"))
		return
	}

	// --- 检查这里：打印的内容是否变了？ ---
	fmt.Printf("成功将 [%s] 数据存入 Kafka\n", data.DeviceID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK: Data Buffered"))
}

func main() {
	// 注册路由
	http.HandleFunc("/upload", handleData)

	fmt.Println("🚀 物联网接收服务器 (生产者) 已启动")
	fmt.Println("📍 监听端口: :8080")
	fmt.Println("🔗 目标 Kafka: localhost:9092")

	// 启动 HTTP 服务
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("❌ 服务器启动失败: %v\n", err)
	}
}
