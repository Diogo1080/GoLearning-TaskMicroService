package server

import (
	"backendGo/internal/store"
	"backendGo/utils"
	"encoding/json"
	"net/http"
)

type Handlers struct {
	AuthHandler *AuthHandler
	TaskHandler *TaskHandler
	UserHandler *UserHandler
	rds         *store.Redis
}

func NewHandlers(taskHandler *TaskHandler, userHandler *UserHandler, authHandler *AuthHandler, rds *store.Redis) *Handlers {
	return &Handlers{TaskHandler: taskHandler, UserHandler: userHandler, AuthHandler: authHandler, rds: rds}
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
