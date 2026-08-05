package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, taskHandler *TaskHandler, userHandler *UserHandler, authHandler *AuthHandler, authMiddleware gin.HandlerFunc) *gin.Engine {
	// Public routes (no auth required)
	public := r.Group("/api")
	public.POST("/register", authHandler.HandleRegister)
	public.POST("/login", authHandler.HandleLogin)
	public.POST("/logout", authHandler.HandleLogout)
	public.POST("/refresh", authHandler.HandleRefreshToken)

	// Protected routes (require valid JWT token via auth microservice)
	protected := r.Group("/api")
	protected.Use(authMiddleware)

	// User endpoints (profile management only)
	protected.GET("/user/id/:id", userHandler.HandleGetUserByID)
	protected.GET("/user/name/:username", userHandler.HandleGetUserByUsername)
	protected.PATCH("/user/password/:id", userHandler.HandleChangePassword)
	protected.PUT("/user/profile", userHandler.HandleUpdateProfile)

	// Task endpoints (all CRUD operations)
	protected.POST("/task", taskHandler.HandleAddTask)
	protected.GET("/task", taskHandler.HandleGetTasks)
	protected.GET("/task/:id", taskHandler.HandleGetTaskById)
	protected.PUT("/task/:id", taskHandler.HandleUpdateTask)
	protected.DELETE("/task/:id", taskHandler.HandleDeleteTask)
	protected.PATCH("/task/complete/:id", taskHandler.HandleCompleteTask)

	return r
}
