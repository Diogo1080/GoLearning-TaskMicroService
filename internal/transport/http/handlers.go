package http

import (
	"regexp"
)

type Handlers struct {
	TaskHandler *TaskHandler
}

func NewHandlers(taskHandler *TaskHandler) *Handlers {
	return &Handlers{TaskHandler: taskHandler}
}

func checkId(id string) bool {
	if len(id) == 0 || !regexp.MustCompile(`^[0-9]+$`).MatchString(id) {
		return true
	}

	return false
}
