package routes

import (
	handlers "backendGo/internal/server"
	"backendGo/internal/service"
	"backendGo/internal/store"
	"backendGo/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, taskService *service.TaskService, authService *service.AuthService, rds *store.Redis) *gin.Engine {
	h := handlers.NewHandlers(taskService, authService, rds)

	public := r.Group("/api")
	public.POST("/login", h.HandleLogin)
	public.POST("/register", h.HandleCreateUser)

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(rds))
	protected.POST("/task", h.HandleAddTask)
	protected.GET("/task", h.HandleGetTasks)
	return r
}
