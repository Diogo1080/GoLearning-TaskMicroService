package middleware

import (
	"time"

	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/observability"
	"github.com/gin-gonic/gin"
)

func PrometheusMetrics(metrics *observability.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.HTTPInFlight.Inc()
		started := time.Now()
		c.Next()
		defer metrics.HTTPInFlight.Dec()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		metrics.ObserveHTTP(c.Request.Method, route, c.Writer.Status(), time.Since(started))
	}
}
