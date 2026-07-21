package routes

import (
	"backendGo/internal/middleware"
	handlers "backendGo/internal/server"
	"backendGo/internal/store"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, taskHandler *handlers.TaskHandler, userHandler *handlers.UserHandler, authHandler *handlers.AuthHandler, rds *store.Redis) *gin.Engine {
	h := handlers.NewHandlers(taskHandler, userHandler, authHandler, rds)

	public := r.Group("/api")
	public.POST("/login", h.AuthHandler.HandleLogin)
	public.POST("/register", h.UserHandler.HandleCreateUser)

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(rds))
	protected.GET("/user/id/:id", h.UserHandler.HandleGetUserByID)
	protected.GET("/user/name/:username", h.UserHandler.HandleGetUserByUsername)
	protected.PATCH("/user/passwordChange", h.UserHandler.HandlePasswordChange)

	protected.POST("/task", h.TaskHandler.HandleAddTask)
	protected.GET("/task", h.TaskHandler.HandleGetTasks)
	protected.GET("/task/:id", h.TaskHandler.HandleGetTaskById)
	protected.PUT("/task/:id", h.TaskHandler.HandleUpdateTask)
	protected.DELETE("/task/:id", h.TaskHandler.HandleDeleteTask)
	protected.PATCH("/task/complete/:id", h.TaskHandler.HandleCompleteTask)

	return r
}
