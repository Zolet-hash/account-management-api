package telemetry

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
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

// Middleware collects HTTP request, error, latency, and financial transaction metrics for Prometheus
func (metrics *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()

		if route == "" {
			route = "unmatched"
		}

		status := http.StatusText(c.Writer.Status())
		metrics.Requests.WithLabelValues(
			c.Request.Method,
			route,
			status,
		).Inc()

		metrics.Duration.WithLabelValues(
			c.Request.Method,
			route,
		).Observe(time.Since(start).Seconds())

		if c.Writer.Status() >= 400 {
			metrics.Errors.WithLabelValues(route, status).Inc()
		}

		if route == "/api/v1/accounts/:accountId/deposit" || route == "/api/v1/accounts/:accountId/withdrawal" {
			typ := "deposit"
			if route == "/api/v1/accounts/:accountId/withdrawal" {
				typ = "withdrawal"
			}
			result := "success"

			if c.Writer.Status() >= 400 {
				result = "failure"
			}
			metrics.Transfers.WithLabelValues(typ, result).Inc()
		}
	}
}

// Init Initializes OTEl distributed tracing and exposes prometheus metrics through an HTTP handler
func Init(ctx context.Context, endpoint string) (func(context.Context) error, error) {

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	if endpoint == "" {
		otel.SetTracerProvider(
			sdktrace.NewTracerProvider(
				sdktrace.WithResource(
					resource.NewWithAttributes(
						semconv.SchemaURL,
						semconv.ServiceName("account-management"),
					),
				),
			),
		)
		return func(context.Context) error {
			return nil
		}, nil
	}

	// an endpoint was provided
	exp, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)

	// handle exporter initialization error
	if err != nil {
		return nil, err
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName("account-management"),
			),
		),
	)

	//register the tracer provider globaly
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Shutdown, nil
}

// exposes prometheus metrics through a http handler
func Handler() http.Handler {
	return promhttp.Handler()
}
