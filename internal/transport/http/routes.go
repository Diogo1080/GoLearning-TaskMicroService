package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, taskHandler *TaskHandler, authMiddleware gin.HandlerFunc) *gin.Engine {

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
