package service

import "backendGo/internal/store"

const Format = "2006-01-02"

type TaskService struct {
	Repo *store.SQLiteTaskRepository
}

func NewTaskService(repo *store.SQLiteTaskRepository) *TaskService {
	return &TaskService{Repo: repo}
}
