package server

import (
	"backendGo/internal/service"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	TaskService *service.TaskService
}

func NewHandlers(taskService *service.TaskService) *Handlers {
	return &Handlers{TaskService: taskService}
}

func (h *Handlers) HandleAddTask(ctx *gin.Context) {

}
