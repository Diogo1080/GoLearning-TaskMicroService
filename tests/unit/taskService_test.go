package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	entities "github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
	service "github.com/Diogo1080/GoLearning-TaskMicroService/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Repository ---

type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) CreateTask(ctx context.Context, task entities.Task) (entities.Task, error) {
	args := m.Called(task)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskRepo) GetTasks(ctx context.Context, terms entities.TaskSearch) ([]entities.Task, error) {
	args := m.Called(terms)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Task), args.Error(1)
}

func (m *MockTaskRepo) GetTaskByID(ctx context.Context, id int, userID int) (entities.Task, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskRepo) UpdateTask(ctx context.Context, task entities.Task, id int, userID int) (entities.Task, error) {
	args := m.Called(task, id, userID)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskRepo) DeleteTask(ctx context.Context, id int, userID int) (int64, error) {
	args := m.Called(id, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTaskRepo) MarkTaskAsDone(ctx context.Context, id int, userID int) (int64, error) {
	args := m.Called(id, userID)
	return args.Get(0).(int64), args.Error(1)
}

func TestTaskService_CreateTask(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	mockRepo.On("CreateTask", mock.MatchedBy(func(task entities.Task) bool {
		return task.Title == "Test" && task.Description == "Desc"
	})).Return(entities.Task{ID: 1, UserID: 1, Title: "Test", Description: "Desc"}, nil)

	svc := service.NewTaskService(mockRepo)
	got, err := svc.CreateTask(t.Context(), entities.Task{Title: "Test", Description: "Desc"})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), got.ID)
	mockRepo.AssertExpectations(t)
}

func TestTaskService_CreateTaskReturnsRepositoryError(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	mockRepo.On("CreateTask", mock.AnythingOfType("domain.Task")).Return(entities.Task{}, errors.New("database error"))

	svc := service.NewTaskService(mockRepo)
	got, err := svc.CreateTask(t.Context(), entities.Task{Title: "Test"})

	assert.Error(t, err)
	assert.Zero(t, got.ID)
	mockRepo.AssertExpectations(t)
}

func TestTaskService_GetTaskByID(t *testing.T) {
	now := time.Now()
	mockRepo := new(MockTaskRepo)
	mockRepo.On("GetTaskByID", 42, 1).Return(entities.Task{ID: 42, Title: "Found Task", Description: "It was found", DueDate: now}, nil)

	svc := service.NewTaskService(mockRepo)
	got, err := svc.GetTaskByID(t.Context(), 42, 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(42), got.ID)
	mockRepo.AssertExpectations(t)
}

func TestTaskService_GetTaskByIDReturnsRepositoryError(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	mockRepo.On("GetTaskByID", 999, 1).Return(entities.Task{}, errors.New("task not found"))

	svc := service.NewTaskService(mockRepo)
	got, err := svc.GetTaskByID(t.Context(), 999, 1)

	assert.Error(t, err)
	assert.Zero(t, got.ID)
	mockRepo.AssertExpectations(t)
}

func TestTaskService_GetTasks_PreservesPage(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	mockRepo.On("GetTasks", entities.TaskSearch{
		UserID: 1,
		Limit:  10,
		Offset: 3,
	}).Return([]entities.Task{{ID: 21}}, nil)

	svc := service.NewTaskService(mockRepo)
	tasks, err := svc.GetTasks(t.Context(), entities.TaskSearch{
		UserID: 1,
		Limit:  10,
		Offset: 3,
	})

	assert.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, int64(21), tasks[0].ID)
	mockRepo.AssertExpectations(t)
}

func TestTaskService_GetTasks_DefaultsToFirstPage(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	mockRepo.On("GetTasks", entities.TaskSearch{
		UserID: 1,
		Limit:  10,
		Offset: 1,
	}).Return([]entities.Task{}, nil)

	svc := service.NewTaskService(mockRepo)
	tasks, err := svc.GetTasks(t.Context(), entities.TaskSearch{
		UserID: 1,
		Limit:  10,
	})

	assert.NoError(t, err)
	assert.Empty(t, tasks)
	mockRepo.AssertExpectations(t)
}

func TestTaskService_MarkTaskAsDone(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	mockRepo.On("MarkTaskAsDone", 10, 1).Return(int64(1), nil)

	svc := service.NewTaskService(mockRepo)
	rows, err := svc.MarkTaskAsDone(t.Context(), 10, 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, rows)
	mockRepo.AssertExpectations(t)
}

func TestTaskService_MarkTaskAsDoneReturnsZeroWhenTaskNotFound(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	mockRepo.On("MarkTaskAsDone", 999, 1).Return(int64(0), nil)

	svc := service.NewTaskService(mockRepo)
	rows, err := svc.MarkTaskAsDone(t.Context(), 999, 1)

	assert.NoError(t, err)
	assert.Zero(t, rows)
	mockRepo.AssertExpectations(t)
}
