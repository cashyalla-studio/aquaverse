package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aquaverse_http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "aquaverse_http_request_duration_seconds",
		Help:    "HTTP request duration",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "aquaverse_active_connections",
		Help: "Active WebSocket + HTTP connections",
	})
)

func PrometheusMetrics() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			activeConnections.Inc()
			defer activeConnections.Dec()

			err := next(c)

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(c.Response().Status)
			path := c.Path() // 파라미터 치환된 경로 (예: /api/v1/fish/:id)

			httpRequestsTotal.WithLabelValues(c.Request().Method, path, status).Inc()
			httpRequestDuration.WithLabelValues(c.Request().Method, path).Observe(duration)

			return err
		}
	}
}
