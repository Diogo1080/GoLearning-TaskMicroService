package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ReadinessChecker func(context.Context) error

func RegisterRoutes(r *gin.Engine, taskHandler *TaskHandler, authMiddleware gin.HandlerFunc, readiness ...ReadinessChecker) *gin.Engine {
	r.GET("/livez", Live)
	r.GET("/readyz", readinessHandler(readiness...))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	public := r.Group("/api")
	public.GET("/health", Health)

	// Protected routes (require valid JWT token via auth microservice)
	protected := r.Group("/api")
	protected.Use(authMiddleware)

	// Task endpoints (all CRUD operations)
	protected.POST("/task", taskHandler.HandleAddTask)
	protected.GET("/task", taskHandler.HandleGetTasks)
	protected.GET("/task/:id", taskHandler.HandleGetTaskById)
	protected.PUT("/task/:id", taskHandler.HandleUpdateTask)
	protected.DELETE("/task/:id", taskHandler.HandleDeleteTask)
	protected.PATCH("/task/complete/:id", taskHandler.HandleCompleteTask)

	return r
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func readinessHandler(checkers ...ReadinessChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		for _, checker := range checkers {
			if err := checker(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
