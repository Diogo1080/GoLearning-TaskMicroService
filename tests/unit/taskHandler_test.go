package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskService) UpdateTask(ctx context.Context, task entities.Task, id int, userID int) (entities.Task, error) {
	args := m.Called(task, id, userID)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskService) DeleteTask(ctx context.Context, id int, userID int) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockTaskService) MarkTaskAsDone(ctx context.Context, id int, userID int) (int, error) {
	args := m.Called(id, userID)
	return args.Get(0).(int), args.Error(1)
}

// --- Handler Tests ---

func TestTaskHandler_CreateTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockSvc := new(MockTaskService)
	mockSvc.On("CreateTask", mock.MatchedBy(func(task entities.Task) bool {
		return task.UserID == 0 && task.Title == "New Task" && task.Description == "Task description"
	})).Return(entities.Task{ID: 1, UserID: 1, Title: "New Task", Description: "Task description"}, nil)
	router.POST("/api/tasks", server.NewTaskHandler(mockSvc).HandleAddTask)

	body := bytes.NewBufferString(`{"title":"New Task","description":"Task description","completed":false}`)
	req, _ := http.NewRequest("POST", "/api/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestTaskHandler_CreateTaskRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockSvc := new(MockTaskService)
	mockSvc.AssertNotCalled(t, "CreateTask", mock.Anything)
	router.POST("/api/tasks", server.NewTaskHandler(mockSvc).HandleAddTask)

	req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestTaskHandler_CreateTaskReturnsServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockSvc := new(MockTaskService)
	mockSvc.On("CreateTask", mock.AnythingOfType("domain.Task")).Return(entities.Task{}, errors.New("databaseerror"))
	router.POST("/api/tasks", server.NewTaskHandler(mockSvc).HandleAddTask)

	body := bytes.NewBufferString(`{"title":"Fail Task","description":"Will fail"}`)
	req, _ := http.NewRequest("POST", "/api/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestTaskHandler_GetTasksUsesDefaultPagination(t *testing.T) {
	runGetTasksHandlerTest(t, "", entities.TaskSearch{Limit: 20, Offset: 1}, []entities.Task{{ID: 1}, {ID: 2}})
}

func TestTaskHandler_GetTasksReturnsEmptyArray(t *testing.T) {
	runGetTasksHandlerTest(t, "", entities.TaskSearch{Limit: 20, Offset: 1}, []entities.Task{})
}

func TestTaskHandler_GetTasksUsesRequestedPage(t *testing.T) {
	runGetTasksHandlerTest(t, "limit=5&offset=3", entities.TaskSearch{Limit: 5, Offset: 3}, []entities.Task{{ID: 11}})
}

func TestTaskHandler_GetTasksDefaultsInvalidPage(t *testing.T) {
	runGetTasksHandlerTest(t, "limit=5&offset=invalid", entities.TaskSearch{Limit: 5, Offset: 1}, []entities.Task{})
}

func TestTaskHandler_GetTasksAppliesSearchAndPagination(t *testing.T) {
	runGetTasksHandlerTest(t, "search=test&limit=10&offset=2", entities.TaskSearch{Search: "test", Limit: 10, Offset: 2}, []entities.Task{{ID: 1}, {ID: 2}})
}

func runGetTasksHandlerTest(t *testing.T, query string, expectedTerms entities.TaskSearch, returnedTasks []entities.Task) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockSvc := new(MockTaskService)
	mockSvc.On("GetTasks", expectedTerms).Return(returnedTasks, nil)
	router.GET("/api/tasks", server.NewTaskHandler(mockSvc).HandleGetTasks)

	req, _ := http.NewRequest("GET", "/api/tasks?"+query, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var tasks []entities.Task
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &tasks))
	assert.Len(t, tasks, len(returnedTasks))
	mockSvc.AssertExpectations(t)
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockSvc := new(MockTaskService)
	mockSvc.On("DeleteTask", 42, 7).Return(nil)
	registerAuthenticatedDeleteRoute(router, mockSvc)

	req, _ := http.NewRequest("DELETE", "/api/tasks/42", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestTaskHandler_DeleteTaskReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockSvc := new(MockTaskService)
	mockSvc.On("DeleteTask", 999, 7).Return(entities.ErrNotFound)
	registerAuthenticatedDeleteRoute(router, mockSvc)

	req, _ := http.NewRequest("DELETE", "/api/tasks/999", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestTaskHandler_DeleteTaskRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockSvc := new(MockTaskService)
	registerAuthenticatedDeleteRoute(router, mockSvc)

	req, _ := http.NewRequest("DELETE", "/api/tasks/abc", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "DeleteTask", mock.Anything, mock.Anything)
}

func registerAuthenticatedDeleteRoute(router *gin.Engine, mockSvc *MockTaskService) {
	router.Use(func(c *gin.Context) {
		c.Set("userID", int32(7))
		c.Next()
	})
	router.DELETE("/api/tasks/:id", server.NewTaskHandler(mockSvc).HandleDeleteTask)
}
