package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for RPC operations.
type Metrics struct {
	requestCount      *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec
	errorCount        *prometheus.CounterVec
	activeConnections *prometheus.GaugeVec
}

// NewMetrics creates a new Metrics instance with default configuration.
func NewMetrics() *Metrics {
	m := &Metrics{}

	m.requestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "rpc",
			Subsystem: "client",
			Name:      "requests_total",
			Help:      "Total number of RPC requests",
		},
		[]string{"direction", "service", "method"},
	)

	m.requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "rpc",
			Subsystem: "server",
			Name:      "request_duration_seconds",
			Help:      "RPC request duration distribution",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"direction", "service", "method"},
	)

	m.errorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "rpc",
			Subsystem: "client",
			Name:      "errors_total",
			Help:      "Total number of RPC errors",
		},
		[]string{"direction", "service", "method", "code"},
	)

	m.activeConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "rpc",
			Subsystem: "client",
			Name:      "active_connections",
			Help:      "Current active connections",
		},
		[]string{"client", "service"},
	)

	return m
}

// ReqFinished records a completed RPC request (success or error).
func (m *Metrics) ReqFinished(direction string, service string, method string, duration time.Duration, code string, isError bool) {
	m.requestCount.WithLabelValues(direction, service, method).Inc()
	m.requestDuration.WithLabelValues(direction, service, method).Observe(duration.Seconds())
	if isError {
		m.errorCount.WithLabelValues(direction, service, method, code).Inc()
	}
}

// ActiveConnections tracks server-side concurrent requests.
func (m *Metrics) ActiveConnections(service string) (release func()) {
	m.activeConnections.WithLabelValues("server", service).Inc()
	return func() {
		m.activeConnections.WithLabelValues("server", service).Dec()
	}
}

// ClientInc increments the active client connection counter.
func (m *Metrics) ClientInc(clientName string, service string) {
	m.activeConnections.WithLabelValues(clientName, service).Inc()
}

// ClientDec decrements the active client connection counter.
func (m *Metrics) ClientDec(clientName string, service string) {
	m.activeConnections.WithLabelValues(clientName, service).Dec()
}

// HTTPHandler returns an http.Handler for exposing metrics over HTTP.
func (m *Metrics) HTTPHandler() http.Handler {
	return promhttp.Handler()
}
