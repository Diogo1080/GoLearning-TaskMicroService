package service

import (
	"context"
	"fmt"
	"log/slog"

	entities "github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/logger"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task entities.Task) (entities.Task, error)
	GetTasks(ctx context.Context, terms entities.TaskSearch) ([]entities.Task, error)
	GetTaskByID(ctx context.Context, id int, userID int) (entities.Task, error)
	UpdateTask(ctx context.Context, task entities.Task, id int, userID int) (entities.Task, error)
	DeleteTask(ctx context.Context, id int, userID int) (int64, error)
	MarkTaskAsDone(ctx context.Context, id int, userID int) (int64, error)
}

type TaskService struct {
	repo   TaskRepository
	logger *slog.Logger
}

func NewTaskService(repo TaskRepository) *TaskService {
	// Validation
	if repo == nil {
		return nil // or panic, depending on policy
	}

	// Setup
	svc := &TaskService{repo: repo, logger: logger.New().WithGroup("TaskService")}
	return svc
}

func (s *TaskService) CreateTask(ctx context.Context, task entities.Task) (entities.Task, error) {
	s.logger.Info("Attempting to create task", "task", task)
	created, err := s.repo.CreateTask(ctx, task)

	if err != nil {
		s.logger.Error("Failed creating task", "error", err)
		return entities.Task{}, err
	}
	s.logger.Info("Successful created task")
	return created, nil
}

func (s *TaskService) GetTasks(ctx context.Context, taskTerms entities.TaskSearch) ([]entities.Task, error) {
	s.logger.Info("Attempting to get tasks", "terms", taskTerms)
	empty := make([]entities.Task, 0)

	//Sanatise the terms
	err := taskTerms.Sanatise()
	if err != nil {
		return empty, entities.ErrBadData
	}

	tasks, err := s.repo.GetTasks(ctx, taskTerms)

	if err != nil {
		return empty, err
	}

	return tasks, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int, userID int) (entities.Task, error) {
	task, err := s.repo.GetTaskByID(ctx, id, userID)

	if err != nil {
		fmt.Print(err)
		return entities.Task{}, err
	}

	if task.ID == 0 {
		return entities.Task{}, entities.ErrNotFound
	}

	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, task entities.Task, taskID int, userID int) (entities.Task, error) {
	task, err := s.repo.UpdateTask(ctx, task, taskID, userID)

	if err != nil {
		return entities.Task{}, err
	}

	task.ID = int64(taskID)

	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int, userID int) error {
	_, err := s.repo.DeleteTask(ctx, id, userID)

	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) MarkTaskAsDone(ctx context.Context, id int, userID int) (int, error) {
	rows, err := s.repo.MarkTaskAsDone(ctx, id, userID)

	if err != nil {
		return 0, err
	}

	return int(rows), nil
}
