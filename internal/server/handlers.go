package server

import "backendGo/internal/service"

type Handlers struct {
	AuthService *service.AuthService
	TaskService *service.TaskService
}

func NewHandlers(taskService *service.TaskService, authService *service.AuthService) *Handlers {
	return &Handlers{TaskService: taskService, AuthService: authService}
}
