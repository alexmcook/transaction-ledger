package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	transactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transactions_total",
			Help: "Total number of transactions processed",
		},
		[]string{"status"},
	)
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries",
			Buckets: []float64{.001, .002, .005, .01, .02, .05, .1},
		},
		[]string{"query"},
	)
)

var TransactionsSuccess = transactionsTotal.WithLabelValues("success")

func NewDBQueryTimer(queryName string) prometheus.Observer {
	return dbQueryDuration.WithLabelValues(queryName)
}
