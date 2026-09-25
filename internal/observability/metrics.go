package observability

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	BuildInfo    *prometheus.GaugeVec
	HTTPRequests *prometheus.CounterVec
	HTTPDuration *prometheus.HistogramVec
	HTTPInFlight prometheus.Gauge
	HTTPErrors   *prometheus.CounterVec
	GRPCRequests *prometheus.CounterVec
	GRPCDuration *prometheus.HistogramVec
	GRPCInFlight prometheus.Gauge
	GRPCErrors   *prometheus.CounterVec
}

func NewMetrics(registerer prometheus.Registerer, metadata ...string) *Metrics {
	metrics := &Metrics{
		BuildInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "service_build_info",
			Help: "Service build metadata.",
		}, []string{"service", "version", "environment"}),
		HTTPRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		}, []string{"method", "route", "status"}),
		HTTPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP request latency in seconds.",
		}, []string{"method", "route", "status"}),
		HTTPInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of in-flight HTTP requests.",
		}),
		HTTPErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP requests with an error status.",
		}, []string{"method", "route", "status"}),
		GRPCRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "grpc_client_requests_total",
			Help: "Total number of outbound gRPC requests.",
		}, []string{"method", "code"}),
		GRPCDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "grpc_client_request_duration_seconds",
			Help: "Outbound gRPC request latency in seconds.",
		}, []string{"method", "code"}),
		GRPCInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "grpc_client_requests_in_flight",
			Help: "Current number of in-flight outbound gRPC requests.",
		}),
		GRPCErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "grpc_client_errors_total",
			Help: "Total number of failed outbound gRPC requests.",
		}, []string{"method", "code"}),
	}
	service, version, environment := "task-service", "dev", "unknown"
	if len(metadata) > 0 {
		service = metadata[0]
	}
	if len(metadata) > 1 {
		version = metadata[1]
	}
	if len(metadata) > 2 {
		environment = metadata[2]
	}
	metrics.BuildInfo.WithLabelValues(service, version, environment).Set(1)

	registerer.MustRegister(
		metrics.BuildInfo,
		metrics.HTTPRequests, metrics.HTTPDuration, metrics.HTTPInFlight, metrics.HTTPErrors,
		metrics.GRPCRequests, metrics.GRPCDuration, metrics.GRPCInFlight, metrics.GRPCErrors,
	)
	return metrics
}

func StatusCode(status int) string {
	return strconv.Itoa(status)
}

func (m *Metrics) ObserveHTTP(method, route string, status int, duration time.Duration) {
	statusCode := StatusCode(status)
	m.HTTPRequests.WithLabelValues(method, route, statusCode).Inc()
	m.HTTPDuration.WithLabelValues(method, route, statusCode).Observe(duration.Seconds())
	if status >= 400 {
		m.HTTPErrors.WithLabelValues(method, route, statusCode).Inc()
	}
}

func (m *Metrics) ObserveGRPC(method, code string, duration time.Duration) {
	m.GRPCRequests.WithLabelValues(method, code).Inc()
	m.GRPCDuration.WithLabelValues(method, code).Observe(duration.Seconds())
	if code != "OK" {
		m.GRPCErrors.WithLabelValues(method, code).Inc()
	}
}
