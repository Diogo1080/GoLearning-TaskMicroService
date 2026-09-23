package service

import (
	"context"
	"time"

	entities "github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/transport/http/middleware"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/validation"
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
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	// Validation
	if repo == nil {
		return nil // or panic, depending on policy
	}

	// Setup
	svc := &TaskService{repo: repo}
	return svc
}

func (s *TaskService) CreateTask(ctx context.Context, task entities.Task) (entities.Task, error) {
	logger := middleware.GetLoggerFromContext(ctx)
	result := validation.SanitizeAndValidateTask(task, time.Time{})
	if !result.Valid {
		return entities.Task{}, entities.ErrBadRequest
	}
	task = result.Sanitized

	logger.Info("Attempting to create task", "task", task)
	created, err := s.repo.CreateTask(ctx, task)

	if err != nil {
		logger.Error("Failed creating task", "error", err)
		return entities.Task{}, err
	}
	logger.Info("Successful created task")
	return created, nil
}

func (s *TaskService) GetTasks(ctx context.Context, taskTerms entities.TaskSearch) ([]entities.Task, error) {
	logger := middleware.GetLoggerFromContext(ctx)

	logger.Info("Attempting to get tasks", "terms", taskTerms)
	empty := make([]entities.Task, 0)

	//Sanatise the terms
	err := taskTerms.Sanatise()
	if err != nil {
		logger.Error("Failed get tasks", "error", err)
		return empty, entities.ErrBadRequest
	}

	tasks, err := s.repo.GetTasks(ctx, taskTerms)

	logger.Info("Successful get tasks")
	if err != nil {
		return empty, err
	}

	return tasks, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int, userID int) (entities.Task, error) {
	logger := middleware.GetLoggerFromContext(ctx)

	logger.Info("Attempting to get Task By ID")
	task, err := s.repo.GetTaskByID(ctx, id, userID)

	if err != nil {
		logger.Error("Failed get task", "error", err)
		return entities.Task{}, err
	}

	logger.Info("Successful got task")
	if task.ID == 0 {
		return entities.Task{}, entities.ErrNotFound
	}

	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, task entities.Task, taskID int, userID int) (entities.Task, error) {
	logger := middleware.GetLoggerFromContext(ctx)
	task.ID = int64(taskID)
	task.UserID = int64(userID)
	result := validation.SanitizeAndValidateTask(task, time.Time{})
	if !result.Valid {
		return entities.Task{}, entities.ErrBadRequest
	}
	task = result.Sanitized

	logger.Info("Attempting to update Task")
	task, err := s.repo.UpdateTask(ctx, task, taskID, userID)

	if err != nil {
		logger.Error("Failed update task", "error", err)
		return entities.Task{}, err
	}

	task.ID = int64(taskID)

	logger.Info("Success update")
	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int, userID int) error {
	logger := middleware.GetLoggerFromContext(ctx)

	logger.Info("Attempting to delete Task")
	_, err := s.repo.DeleteTask(ctx, id, userID)

	if err != nil {
		logger.Error("Failed deleting task", "error", err)
		return err
	}

	logger.Info("Successful task delete")
	return nil
}

func (s *TaskService) MarkTaskAsDone(ctx context.Context, id int, userID int) (int, error) {
	logger := middleware.GetLoggerFromContext(ctx)

	logger.Info("Attempting to mark task as done")
	rows, err := s.repo.MarkTaskAsDone(ctx, id, userID)

	if err != nil {
		logger.Error("Failed mark task as done ")
		return 0, err
	}

	logger.Info("Successful mark task as done")
	return int(rows), nil
}
