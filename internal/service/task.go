package service

import (
	"backendGo/internal/entities"
	"backendGo/internal/store"
)

const Format = "2006-01-02"

type TaskService struct {
	Repo *store.SQLiteTaskRepository
}

func NewTaskService(repo *store.SQLiteTaskRepository) *TaskService {
	return &TaskService{Repo: repo}
}

func (s *TaskService) CreateTask(taskDTO entities.TaskDTO) (entities.Task, error) {

	task, err := taskDTO.ToTask()

	if err != nil {
		return entities.Task{}, err
	}

	return s.Repo.CreateTask(task)
}
