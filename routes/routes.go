package routes

import (
	"backendGo/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(taskService *service.TaskService) *gin.Engine {
	r := gin.Default()

	return r
}
