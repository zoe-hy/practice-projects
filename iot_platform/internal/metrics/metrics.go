package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	MessagesProduced = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "iot_messages_produced_total",
		Help: "Total number of messages produced to Kafka",
	})
	MessagesConsumed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "iot_messages_consumed_total",
		Help: "Total number of messages consumed from Kafka",
	})
)

func init() {
	prometheus.MustRegister(MessagesProduced)
	prometheus.MustRegister(MessagesConsumed)
}

// ServeMetrics 在指定地址暴露 Prometheus metrics
func ServeMetrics(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}
