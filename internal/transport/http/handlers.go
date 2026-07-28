package http

import (
	"backendGo/internal/store"
	"regexp"
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

func checkId(id string) bool {
	if len(id) == 0 || !regexp.MustCompile(`^[0-9]+$`).MatchString(id) {
		return true
	}

	return false

}
