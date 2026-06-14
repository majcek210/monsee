package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the platform.
type Metrics struct {
	CheckDuration       *prometheus.HistogramVec
	MonitorStatusTotal  *prometheus.CounterVec
	ActiveIncidents     prometheus.Gauge
	QueueDepth          prometheus.Gauge
	HTTPDuration        *prometheus.HistogramVec
	RateLimiterFailOpen prometheus.Counter
}

// NewMetrics registers and returns all platform metrics.
func NewMetrics() *Metrics {
	return &Metrics{
		CheckDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "monitor_check_duration_seconds",
			Help:    "Duration of monitor checks in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"type"}),

		MonitorStatusTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "monitor_status_total",
			Help: "Total number of monitor check results by status.",
		}, []string{"status"}),

		ActiveIncidents: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "active_incidents_total",
			Help: "Current number of open incidents.",
		}),

		QueueDepth: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "queue_depth",
			Help: "Current number of tasks in the monitor check queue.",
		}),

		HTTPDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"route", "method", "status"}),

		RateLimiterFailOpen: promauto.NewCounter(prometheus.CounterOpts{
			Name: "rate_limiter_fail_open_total",
			Help: "Total number of requests allowed through because the rate limiter backend (Redis) was unavailable.",
		}),
	}
}
