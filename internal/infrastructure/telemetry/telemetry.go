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

func (collector *poolCollector) Collect(ch chan<- prometheus.Metric) {

	stats := collector.pool.Stat()
	for state, value := range map[string]float64{"total": float64(stats.TotalConns()), "acquired": float64(stats.AcquiredConns()), "idle": float64(stats.IdleConns()), "max": float64(stats.MaxConns())} {
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, value, state)
	}
}
