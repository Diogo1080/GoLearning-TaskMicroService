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
	protected.GET("/user/id/:id", h.HandleGetUserByID)
	protected.GET("/user/name/:username", h.HandleGetUserByUsername)
	protected.PATCH("/user/passwordChange", h.HandlePasswordChange)

	protected.POST("/task", h.HandleAddTask)
	protected.GET("/task", h.HandleGetTasks)
	protected.GET("/task/:id", h.HandleGetTaskById)
	protected.PUT("/task/:id", h.HandleUpdateTask)
	protected.DELETE("/task/:id", h.HandleDeleteTask)
	protected.PATCH("/task/complete/:id", h.HandleCompleteTask)

	return r
}
