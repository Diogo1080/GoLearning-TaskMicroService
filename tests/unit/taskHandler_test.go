package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	entities "github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
	server "github.com/Diogo1080/GoLearning-TaskMicroService/internal/transport/http"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Reuse the same mock pattern
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, task entities.Task) (entities.Task, error) {
	args := m.Called(task)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskService) GetTasks(ctx context.Context, terms entities.TaskSearch) ([]entities.Task, error) {
	args := m.Called(terms)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Task), args.Error(1)
}

func (m *MockTaskService) GetTaskByID(ctx context.Context, id int, userID int) (entities.Task, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskService) UpdateTask(ctx context.Context, task entities.Task, id int, userID int) (entities.Task, error) {
	args := m.Called(task, id)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskService) DeleteTask(ctx context.Context, id int, userID int) error {
	args := m.Called(id)
	return args.Error(1)
}

func (m *MockTaskService) MarkTaskAsDone(ctx context.Context, id int, userID int) (int, error) {
	args := m.Called(id)
	return args.Get(0).(int), args.Error(1)
}

// --- Handler Tests ---

func TestTaskHandler_CreateTask(t *testing.T) {
	type testCase struct {
		name        string
		requestBody map[string]interface{}
		setupMock   func(*MockTaskService)
		wantStatus  int
		expectID    int
		expectError bool
	}

	tests := []testCase{
		{
			name: "creates task successfully",
			requestBody: map[string]interface{}{
				"title":       "New Task",
				"userID":      1,
				"description": "Task description",
				"completed":   false,
			},
			setupMock: func(svc *MockTaskService) {
				svc.On("CreateTask", mock.AnythingOfType("domain.Task")).
					Return(entities.Task{
						ID:          1,
						UserID:      1,
						Title:       "New Task",
						Description: "Task description",
						Completed:   false,
						DueDate:     time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusCreated,
			expectID:    1,
			expectError: false,
		},
		{
			name:        "rejects invalid JSON",
			requestBody: nil,
			setupMock: func(svc *MockTaskService) {
				svc.AssertNotCalled(t, "CreateTask", mock.AnythingOfType("domain.Task"))
			},
			wantStatus:  http.StatusBadRequest,
			expectID:    0,
			expectError: true,
		},
		{
			name: "handles service error",
			requestBody: map[string]interface{}{
				"title":       "Fail Task",
				"description": "Will fail",
			},
			setupMock: func(svc *MockTaskService) {
				svc.On("CreateTask", mock.AnythingOfType("domain.Task")).
					Return(entities.Task{}, errors.New("databaseerror"))
			},
			wantStatus:  http.StatusInternalServerError,
			expectID:    0,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup Gin in test mode
			gin.SetMode(gin.TestMode)
			router := gin.Default()

			// Arrange
			mockSvc := new(MockTaskService)
			tc.setupMock(mockSvc)
			h := server.NewTaskHandler(mockSvc)

			router.POST("/api/tasks", h.HandleAddTask)

			var body bytes.Buffer
			if tc.requestBody != nil {
				json.NewEncoder(&body).Encode(tc.requestBody)
			}

			req, _ := http.NewRequest("POST", "/api/tasks", &body)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			// Act
			router.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tc.wantStatus, rr.Code)

			if tc.expectError && tc.wantStatus != http.StatusBadRequest {
				var resp map[string]string
				json.Unmarshal(rr.Body.Bytes(), &resp)
			} else if !tc.expectError && tc.wantStatus == http.StatusCreated {
				var resp map[string]interface{}
				json.Unmarshal(rr.Body.Bytes(), &resp)
				assert.Equal(t, tc.expectID, int(resp["id"].(float64)))
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_GetTasks(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		setupMock   func(*MockTaskService)
		wantStatus  int
		taskCount   int
	}{
		{
			name:        "returns all tasks with default limit",
			queryParams: "",
			setupMock: func(svc *MockTaskService) {
				svc.On("GetTasks", entities.TaskSearch{Limit: 20}).Return([]entities.Task{
					{ID: 1, Title: "Task 1", DueDate: time.Now()},
					{ID: 2, Title: "Task 2", DueDate: time.Now()},
				}, nil)
			},
			wantStatus: http.StatusOK,
			taskCount:  2,
		},
		{
			name:        "returns no tasks with default limit",
			queryParams: "",
			setupMock: func(svc *MockTaskService) {
				svc.On("GetTasks", entities.TaskSearch{Limit: 20}).Return([]entities.Task{}, nil)
			},
			wantStatus: http.StatusOK,
			taskCount:  0,
		},
		{
			name:        "filters by title and description query param and limit",
			queryParams: "search=test&limit=10",
			setupMock: func(svc *MockTaskService) {
				svc.On("GetTasks", entities.TaskSearch{Search: "test", Limit: 10}).Return([]entities.Task{
					{ID: 1, Title: "Testing", DueDate: time.Now()},
					{ID: 2, Title: "Nothing", DueDate: time.Now()},
				}, nil)
			},
			wantStatus: http.StatusOK,
			taskCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()

			mockSvc := new(MockTaskService)
			tt.setupMock(mockSvc)
			h := server.NewTaskHandler(mockSvc)

			router.GET("/api/tasks", h.HandleGetTasks)

			req, _ := http.NewRequest("GET", "/api/tasks?"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			var tasks []entities.Task
			json.Unmarshal(rr.Body.Bytes(), &tasks)
			assert.Len(t, tasks, tt.taskCount)

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		setupMock  func(*MockTaskService)
		wantStatus int
	}{
		{
			name:   "deletes task successfully",
			taskID: "42",
			setupMock: func(svc *MockTaskService) {
				svc.On("DeleteTask", 42).Return(int64(1), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "returns 404 when task not found",
			taskID: "999",
			setupMock: func(svc *MockTaskService) {
				svc.On("DeleteTask", 999).Return(int64(0), entities.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "returns 400 for invalid ID",
			taskID: "abc",
			setupMock: func(svc *MockTaskService) {
				svc.AssertNotCalled(t, "DeleteTask", mock.Anything)
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()

			mockSvc := new(MockTaskService)
			tt.setupMock(mockSvc)
			h := server.NewTaskHandler(mockSvc)

			router.DELETE("/api/tasks/:id", h.HandleDeleteTask)

			req, _ := http.NewRequest("DELETE", "/api/tasks/"+tt.taskID, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}
