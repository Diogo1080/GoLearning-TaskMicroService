package server

import (
	"backendGo/internal/service"
	"backendGo/internal/store"
	"backendGo/utils"
	"encoding/json"
	"net/http"
)

type Handlers struct {
	AuthService *service.AuthService
	TaskService *service.TaskService
	rds         *store.Redis
}

func NewHandlers(taskService *service.TaskService, authService *service.AuthService, rds *store.Redis) *Handlers {
	return &Handlers{TaskService: taskService, AuthService: authService, rds: rds}
}

func sendJSONResponse(res http.ResponseWriter, statusCode int, data interface{}) {
	respBytes, err := json.Marshal(data)
	if err != nil {
		utils.SendErrorResponse(res, "Error:", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	_, err = res.Write(respBytes)

	if err != nil {
		utils.SendErrorResponse(res, "Error writing", http.StatusInternalServerError)
		return
	}
}
