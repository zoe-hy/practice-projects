package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iot_platform/internal/config"
	"iot_platform/internal/db"
	"iot_platform/internal/logging"
	"iot_platform/internal/metrics"
	"iot_platform/internal/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

// RunConsumer 启动 Kafka 消费循环并阻塞，使用传入配置
func RunConsumer(cfg *config.Config) error {
	logging.Init()
	log.Info().Msg("=== 🚚 搬运工（消费者）正在启动 ===")

	sqlDB, err := db.OpenDB(cfg.MySQLDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("数据库连接配置错误")
		return err
	}

	if err = sqlDB.Ping(); err != nil {
		log.Fatal().Err(err).Msg("连不上 MySQL")
		return err
	}

	// 如果需要运行简单迁移
	if err := db.RunMigrations(sqlDB, "sql.sql"); err != nil {
		log.Error().Err(err).Msg("migration warning")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{cfg.KafkaBroker},
		Topic:       cfg.KafkaTopic,
		GroupID:     "db-saver-v2",
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	// metrics server
	go func() {
		metricsAddr := ":" + fmt.Sprintf("%d", 9100)
		_ = metrics.ServeMetrics(metricsAddr)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run {
		select {
		case <-stop:
			run = false
			break
		default:
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("读取 Kafka 失败")
				time.Sleep(1 * time.Second)
				continue
			}

			var data model.DeviceData
			if err = json.Unmarshal(m.Value, &data); err != nil {
				log.Error().Err(err).Str("raw", string(m.Value)).Msg("解析数据失败")
				continue
			}

			metrics.MessagesConsumed.Inc()
			if data.Value > 30 {
				log.Warn().Str("device", data.DeviceID).Float64("value", data.Value).Msg("temperature high")
			}

			query := "INSERT INTO device_logs (device_id, value) VALUES (?, ?)"
			if _, err := db.ExecPrepared(sqlDB, query, data.DeviceID, data.Value); err != nil {
				log.Error().Err(err).Msg("入库 SQL 报错")
			} else {
				log.Info().Str("device", data.DeviceID).Float64("value", data.Value).Msg("inserted")
			}
		}
	}

	// 关闭数据库
	_ = sqlDB.Close()
	log.Info().Msg("consumer exiting")
	return nil
}
