package routes

import (
	handlers "backendGo/internal/server"
	"backendGo/internal/service"
	"backendGo/internal/store"
	"backendGo/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(taskService *service.TaskService, authService *service.AuthService, rds *store.Redis) *gin.Engine {
	r := gin.Default()

	h := handlers.NewHandlers(taskService, authService)

	r.POST("/api/login", h.HandleLogin(rds))
	r.POST("/api/task", middleware.AuthMiddleware(rds), h.HandleAddTask())
	return r
}
