package service

import (
	entities "backendGo/internal/domain"
	"fmt"
)

/*
	type TaskCreator interface {
		CreateTask(task entities.Task) (entities.Task, error)
	}

	type TaskReader interface {
		GetTasks(terms entities.TaskSearch, limit int) ([]entities.Task, error)
		GetTaskByID(id int) (entities.Task, error)
	}

	type TaskWriter interface {
		UpdateTask(task entities.Task, id int) (entities.Task, error)
		DeleteTask(id int) (int64, error)
		MarkTaskAsDone(id int) (int64, error)
	}
*/

type TaskRepository interface {
	CreateTask(task entities.Task) (entities.Task, error)
	GetTasks(terms entities.TaskSearch) ([]entities.Task, error)
	GetTaskByID(id int, userID int) (entities.Task, error)
	UpdateTask(task entities.Task, id int, userID int) (entities.Task, error)
	DeleteTask(id int, userID int) (int64, error)
	MarkTaskAsDone(id int, userID int) (int64, error)
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

func (s *TaskService) CreateTask(task entities.Task) (entities.Task, error) {
	created, err := s.repo.CreateTask(task)

	if err != nil {
		return entities.Task{}, err
	}

	return created, nil
}

func (s *TaskService) GetTasks(taskTerms entities.TaskSearch) ([]entities.Task, error) {
	empty := make([]entities.Task, 0)

	//Sanatise the
	err := taskTerms.Sanatise()
	if err != nil {
		return empty, entities.ErrBadData
	}

	tasks, err := s.repo.GetTasks(taskTerms)

	if err != nil {
		return empty, err
	}

	return tasks, nil
}

func (s *TaskService) GetTaskByID(id int, userID int) (entities.Task, error) {
	task, err := s.repo.GetTaskByID(id, userID)

	if err != nil {
		fmt.Print(err)
		return entities.Task{}, err
	}

	if task.ID == 0 {
		return entities.Task{}, entities.ErrNotFound
	}

	return task, nil
}

func (s *TaskService) UpdateTask(task entities.Task, taskID int, userID int) (entities.Task, error) {
	task, err := s.repo.UpdateTask(task, taskID, userID)

	if err != nil {
		return entities.Task{}, err
	}

	task.ID = int64(taskID)

	return task, nil
}

func (s *TaskService) DeleteTask(id int, userID int) error {
	_, err := s.repo.DeleteTask(id, userID)

	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) MarkTaskAsDone(id int, userID int) (int, error) {
	rows, err := s.repo.MarkTaskAsDone(id, userID)

	if err != nil {
		return 0, err
	}

	return int(rows), nil
}
