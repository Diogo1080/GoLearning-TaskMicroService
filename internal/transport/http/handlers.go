package http

import (
	"strconv"

	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
)

type Handlers struct {
	TaskHandler *TaskHandler
}

func NewHandlers(taskHandler *TaskHandler) *Handlers {
	return &Handlers{TaskHandler: taskHandler}
}

func checkId(id string) bool {
	_, err := parseID(id)
	return err != nil
}

func parseID(id string) (int, error) {
	parsed, err := strconv.ParseUint(id, 10, strconv.IntSize)
	if err != nil || parsed == 0 {
		return 0, domain.ErrBadRequest
	}
	return int(parsed), nil
}
