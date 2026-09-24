package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

type TaskRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Completed   bool      `json:"completed"`
	DueDate     time.Time `json:"dueDate"`
}

func (r TaskRequest) ToDomain() domain.Task {
	return domain.Task{
		Title:       r.Title,
		Description: r.Description,
		Priority:    r.Priority,
		Completed:   r.Completed,
		DueDate:     r.DueDate,
	}
}

type TaskResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Priority    int       `json:"priority"`
	Completed   bool      `json:"completed"`
	DueDate     time.Time `json:"dueDate"`
}

func newTaskResponse(task domain.Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    task.Priority,
		Completed:   task.Completed,
		DueDate:     task.DueDate,
	}
}

type TaskServicePort interface {
	CreateTask(ctx context.Context, task domain.Task) (domain.Task, error)
	GetTasks(ctx context.Context, terms domain.TaskSearch) ([]domain.Task, error)
	GetTaskByID(ctx context.Context, id int, userID int) (domain.Task, error)
	UpdateTask(ctx context.Context, task domain.Task, taskID int, userID int) (domain.Task, error)
	DeleteTask(ctx context.Context, id int, userID int) error
	MarkTaskAsDone(ctx context.Context, id int, userID int) (int, error)
}

type TaskHandler struct {
	svc TaskServicePort
}

func NewTaskHandler(svc TaskServicePort) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) HandleAddTask(c *gin.Context) {
	logger := middleware.GetLoggerFromContext(c)

	authID := GetUserIDFromContext(c)

	var request TaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Info("Failed to bind User")
		c.JSON(http.StatusBadRequest, domain.ErrBadRequest)
		return
	}

	in := request.ToDomain()
	in.UserID = int64(authID)

	task, err := h.svc.CreateTask(c, in)

	if err != nil {
		c.JSON(mapDomainError(err))
		return
	}

	c.JSON(http.StatusCreated, newTaskResponse(task))
}

func (h *TaskHandler) HandleGetTasks(c *gin.Context) {
	//logger := middleware.GetLoggerFromContext(c)

	authID := GetUserIDFromContext(c)
	var limit, offset int

	limit, err := strconv.Atoi(c.Request.URL.Query().Get("limit"))

	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err = strconv.Atoi(c.Request.URL.Query().Get("offset"))
	if err != nil || offset < 1 {
		offset = 1
	}

	terms := domain.TaskSearch{
		Search:     c.Request.URL.Query().Get("search"),
		UserID:     int(authID),
		Priority:   c.Request.URL.Query().Get("priority"),
		Completed:  c.Request.URL.Query().Get("completed"),
		DueDateMin: c.Request.URL.Query().Get("dueDateMin"),
		DueDateMax: c.Request.URL.Query().Get("dueDateMax"),
		OrderBy:    c.Request.URL.Query().Get("orderBy"),
		Limit:      limit,
		Offset:     offset,
	}

	tasks, err := h.svc.GetTasks(c, terms)

	if err != nil {
		c.JSON(mapDomainError(err))
		return
	}

	if len(tasks) == 0 {
		tasks = []domain.Task{}
	}

	responses := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, newTaskResponse(task))
	}
	c.JSON(http.StatusOK, responses)
}

func (h *TaskHandler) HandleGetTaskById(c *gin.Context) {
	id := c.Param("id")

	authID := GetUserIDFromContext(c)

	if checkId(id) {
		c.JSON(http.StatusBadRequest, domain.ErrBadRequest)
		return
	}

	i, _ := strconv.Atoi(id)

	task, err := h.svc.GetTaskByID(c, i, int(authID))

	if err != nil {
		c.JSON(mapDomainError(err))
		return
	}

	c.JSON(http.StatusOK, newTaskResponse(task))
}

func (h *TaskHandler) HandleUpdateTask(c *gin.Context) {

	authID := GetUserIDFromContext(c)

	var request TaskRequest
	id := c.Param("id")

	if checkId(id) {
		c.JSON(http.StatusBadRequest, domain.ErrBadRequest)
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrBadRequest)
		return
	}

	taskID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrBadRequest)
		return
	}

	in := request.ToDomain()
	task, err := h.svc.UpdateTask(c, in, taskID, int(authID))

	if err != nil {
		c.JSON(mapDomainError(err))
		return
	}

	c.JSON(http.StatusOK, newTaskResponse(task))
}

func (h *TaskHandler) HandleDeleteTask(c *gin.Context) {
	authID := GetUserIDFromContext(c)

	id := c.Param("id")

	if checkId(id) {
		c.JSON(http.StatusBadRequest, domain.ErrBadRequest)
		return
	}

	i, _ := strconv.Atoi(id)

	err := h.svc.DeleteTask(c, i, int(authID))

	if err != nil {
		c.JSON(mapDomainError(err))
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h *TaskHandler) HandleCompleteTask(c *gin.Context) {
	authID := GetUserIDFromContext(c)

	id := c.Param("id")

	if checkId(id) {
		c.JSON(http.StatusBadRequest, domain.ErrBadRequest)
		return
	}

	i, _ := strconv.Atoi(id)

	rows, err := h.svc.MarkTaskAsDone(c, i, int(authID))

	if err != nil {
		c.JSON(mapDomainError(err))
		return
	}

	c.JSON(http.StatusOK, rows)
}

//
// Helpers
//

func GetUserIDFromContext(c *gin.Context) int32 {
	return c.GetInt32("userID")
}
