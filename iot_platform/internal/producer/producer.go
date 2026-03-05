package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iot_platform/internal/config"
	"iot_platform/internal/logging"
	"iot_platform/internal/metrics"
	"iot_platform/internal/model"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

// RunProducer 启动生产者 HTTP 服务并阻塞，使用传入配置
func RunProducer(cfg *config.Config) error {
	logging.Init()

	// 初始化 Kafka writer（基于配置）
	kafkaWriter := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBroker),
		Topic:        cfg.KafkaTopic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
	}

	// HTTP handler
	handleData := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
			return
		}

		var data model.DeviceData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "数据格式错误", http.StatusBadRequest)
			return
		}

		msg, _ := json.Marshal(data)

		// 简单重试逻辑
		var err error
		for i := 0; i < 3; i++ {
			err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{Value: msg})
			if err == nil {
				break
			}
			time.Sleep(time.Duration(100*(1<<i)) * time.Millisecond)
		}
		if err != nil {
			log.Error().Err(err).Msg("发送到 Kafka 失败")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Kafka 写入失败"))
			return
		}

		metrics.MessagesProduced.Inc()
		log.Info().Str("device", data.DeviceID).Float64("value", data.Value).Msg("produced")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK: Data Buffered"))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", handleData)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	// metrics server on separate port (HTTP_PORT+1)
	go func() {
		metricsAddr := ":" + fmt.Sprintf("%d", atoi(cfg.HTTPPort)+1)
		_ = metrics.ServeMetrics(metricsAddr)
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("producer http server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("producer server error")
		}
	}()

	<-stop
	log.Info().Msg("shutting down producer server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func atoi(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
