package config

import (
	"os"
)

// Config 保存运行时配置，可从环境变量加载
type Config struct {
	KafkaBroker string // e.g. localhost:9092
	KafkaTopic  string // e.g. device_data
	HTTPPort    string // e.g. 8080
	MySQLDSN    string // e.g. root:root@tcp(127.0.0.1:3307)/iot_db?parseTime=true
}

// Load 从环境变量读取配置，若未设置则采用合理默认值
func Load() *Config {
	cfg := &Config{
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:  getEnv("KAFKA_TOPIC", "device_data"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		MySQLDSN:    getEnv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3307)/iot_db?parseTime=true"),
	}
	return cfg
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
