package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	transactionsStaged = promauto.NewCounter(prometheus.CounterOpts{
		Name: "worker_transactions_staged_total",
		Help: "Total number of transactions staged by the worker",
	})

	fetchLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "worker_transaction_fetch_duration_seconds",
		Help:    "Duration of transaction fetching by the worker",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0},
	})

	dbWriteLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "worker_transaction_processing_duration_seconds",
		Help:    "Duration of transaction processing by the worker",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0},
	})
)
