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

func (s *TaskService) CreateTask(taskDTO entities.TaskDTO) (entities.TaskDTO, error) {

	task, err := taskDTO.ToTask()

	if err != nil {
		return entities.TaskDTO{}, err
	}

	return s.Repo.CreateTask(task)
}

func (s *TaskService) UpdateTask(taskDto entities.TaskDTO, id int) (entities.TaskDTO, error) {
	newTask, err := taskDto.ToTask()

	if err != nil {
		return entities.TaskDTO{}, err
	}

	taskDTO, err := s.Repo.UpdateTask(newTask, id)

	if err != nil {
		return entities.TaskDTO{}, err
	}

	return taskDTO, nil
}
