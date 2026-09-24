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
	args := m.Called(id)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskRepo) UpdateTask(ctx context.Context, task entities.Task, id int, userID int) (entities.Task, error) {
	args := m.Called(task, id)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskRepo) DeleteTask(ctx context.Context, id int, userID int) (int64, error) {
	args := m.Called(id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTaskRepo) MarkTaskAsDone(ctx context.Context, id int, userID int) (int64, error) {
	args := m.Called(id)
	return args.Get(0).(int64), args.Error(1)
}

func TestTaskService_CreateTask(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*MockTaskRepo)
		input      entities.Task
		wantErr    bool
		expectedID int64
	}{
		{
			name: "creates task successfully",
			setupMock: func(repo *MockTaskRepo) {
				repo.On("CreateTask", mock.AnythingOfType("domain.Task")).Return(entities.Task{
					ID:          1,
					UserID:      1,
					Title:       "Test",
					Description: "Desc",
					Completed:   false,
					DueDate:     time.Now(),
				}, nil)
			},
			input: entities.Task{
				Title:       "Test",
				Description: "Desc",
				Completed:   false,
			},
			wantErr:    false,
			expectedID: 1,
		},
		{
			name: "handles repository error",
			setupMock: func(repo *MockTaskRepo) {
				repo.On("CreateTask", mock.AnythingOfType("domain.Task")).
					Return(entities.Task{}, errors.New("database error"))
			},
			input:      entities.Task{Title: "Test"},
			wantErr:    true,
			expectedID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockTaskRepo)
			tt.setupMock(mockRepo)

			svc := service.NewTaskService(mockRepo)

			// Act
			got, err := svc.CreateTask(t.Context(), tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedID, got.ID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, got.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_GetTaskByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		setupMock  func(*MockTaskRepo)
		input      int
		userID     int
		wantErr    bool
		wantTaskID int64
	}{
		{
			name: "returns task when found",
			setupMock: func(repo *MockTaskRepo) {
				repo.On("GetTaskByID", 42).Return(entities.Task{
					ID:          42,
					Title:       "Found Task",
					Description: "It was found",
					Completed:   false,
					DueDate:     now,
				}, nil)
			},
			input:      42,
			userID:     1,
			wantErr:    false,
			wantTaskID: 42,
		},
		{
			name: "returns error when not found",
			setupMock: func(repo *MockTaskRepo) {
				repo.On("GetTaskByID", 999).Return(entities.Task{}, errors.New("task not found"))
			},
			input:      999,
			userID:     1,
			wantErr:    true,
			wantTaskID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepo)
			tt.setupMock(mockRepo)

			svc := service.NewTaskService(mockRepo)
			got, err := svc.GetTaskByID(t.Context(), tt.input, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTaskID, got.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
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
	tests := []struct {
		name      string
		setupMock func(*MockTaskRepo)
		input     int
		userID    int
		wantRows  int
		wantErr   bool
	}{
		{
			name: "marks task as done successfully",
			setupMock: func(repo *MockTaskRepo) {
				repo.On("MarkTaskAsDone", 10).Return(int64(1), nil)
			},
			input:    10,
			userID:   1,
			wantRows: 1,
			wantErr:  false,
		},
		{
			name: "handles task not found",
			setupMock: func(repo *MockTaskRepo) {
				repo.On("MarkTaskAsDone", 999).Return(int64(0), nil)
			},
			input:    999,
			userID:   1,
			wantRows: 0,
			wantErr:  false, // rowsAffected = 0 tells us it wasn't found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepo)
			tt.setupMock(mockRepo)

			svc := service.NewTaskService(mockRepo)
			rows, err := svc.MarkTaskAsDone(t.Context(), tt.input, tt.userID)

			assert.Equal(t, tt.wantRows, rows)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
