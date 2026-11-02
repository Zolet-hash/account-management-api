package telemetry

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// holds different prometheus metrics
type Metrics struct {
	Requests, Errors *prometheus.CounterVec
	Duration         *prometheus.HistogramVec
	Transfers        *prometheus.CounterVec
}

// for monitoring postgres conn pool
type poolCollector struct {
	pool *pgxpool.Pool
	desc *prometheus.Desc
}

func (collector *poolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

// Collect collects postgresql statistics and exposes them as a prometheus metrics
func (collector *poolCollector) Collect(ch chan<- prometheus.Metric) {
	// get database conn pool statistics
	stats := collector.pool.Stat()

	for state, value := range map[string]float64{
		"total":    float64(stats.TotalConns()),
		"acquired": float64(stats.AcquiredConns()),
		"idle":     float64(stats.IdleConns()),
		"max":      float64(stats.MaxConns()),
	} {

		// send metrics to prometheous
		ch <- prometheus.MustNewConstMetric(
			collector.desc,
			prometheus.GaugeValue,
			value,
			state,
		)
	}
}

// NewMetrics initializes & registers application and postgresql conn pool wuth prometheus
func NewMetrics(db *pgxpool.Pool) *Metrics {

	// tracks total no of HTTP requests handled by the application
	metrics := &Metrics{

		Requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "HTTP requests",
			},
			[]string{"method", "route", "status"},
		),

		// tracks total no of HTTP errs
		Errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "HTTP errors",
			},
			[]string{"route", "status"}),

		// measures how long http methods take to complete
		Duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "route"}),

		//tracks total number of transactions processed
		Transfers: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "financial_transactions_total",
				Help: "Financial transactions",
			},
			[]string{"type", "result"})}

	// register all metrics with prometheous
	prometheus.MustRegister(
		metrics.Requests,
		metrics.Errors,
		metrics.Duration,
		metrics.Transfers,
		&poolCollector{
			pool: db,
			desc: prometheus.NewDesc(
				"database_pool_connections",
				"PostgreSQL pool connections by state",
				[]string{"state"},
				nil,
			),
		},
	)
	return metrics
}
