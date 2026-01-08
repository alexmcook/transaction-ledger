package api

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"path", "method", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"path", "method", "status"})

	kafkaProducerLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "kafka_produce_latency_seconds",
		Help:    "Latency of producing messages to Kafka",
		Buckets: prometheus.DefBuckets,
	})

	kafkaTransactionsProduced = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_produced_total",
		Help: "Total number of messages produced to Kafka",
	})

	unmarshalLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "unmarshal_latency_seconds",
		Help:    "Latency of unmarshalling data",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0},
	}, []string{"endpoint"})
)

func prometheusMiddleware(c fiber.Ctx) error {
	start := time.Now()
	err := c.Next()

	status := strconv.Itoa(c.Response().StatusCode())
	path := c.Path()
	method := c.Method()

	httpRequestsTotal.WithLabelValues(path, method, status).Inc()
	httpRequestDuration.WithLabelValues(path, method, status).Observe(time.Since(start).Seconds())

	return err
}
