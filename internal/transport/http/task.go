package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
	entities "github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"

	"github.com/gin-gonic/gin"
)

type TaskServicePort interface {
	CreateTask(ctx context.Context, task entities.Task) (entities.Task, error)
	GetTasks(ctx context.Context, terms entities.TaskSearch) ([]entities.Task, error)
	GetTaskByID(ctx context.Context, id int, userID int) (entities.Task, error)
	UpdateTask(ctx context.Context, task entities.Task, taskID int, userID int) (entities.Task, error)
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
	authID := GetUserIDFromContext(c)

	var in entities.Task
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	err := SanitizeInput(in)

	if err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	in.UserID = int64(authID)

	task, err := h.svc.CreateTask(c, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.ErrInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, task.ToTaskDTO())
}

func (h *TaskHandler) HandleGetTasks(c *gin.Context) {
	authID := GetUserIDFromContext(c)
	terms := entities.TaskSearch{
		Search:     c.Request.URL.Query().Get("search"),
		UserID:     int(authID),
		Priority:   c.Request.URL.Query().Get("priority"),
		Completed:  c.Request.URL.Query().Get("completed"),
		DueDateMin: c.Request.URL.Query().Get("dueDateMin"),
		DueDateMax: c.Request.URL.Query().Get("dueDateMax"),
		OrderBy:    c.Request.URL.Query().Get("orderBy"),
		Limit:      100,
	}

	tasks, err := h.svc.GetTasks(c, terms)

	if err != nil {
		if errors.Is(err, entities.ErrBadData) {
			c.JSON(http.StatusBadRequest, entities.ErrBadData)
			return
		}

		c.JSON(http.StatusInternalServerError, entities.ErrInternalServerError)
		return
	}

	if len(tasks) == 0 {
		tasks = []entities.Task{}
		c.JSON(http.StatusOK, tasks)
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) HandleGetTaskById(c *gin.Context) {
	id := c.Param("id")

	authID := GetUserIDFromContext(c)

	if checkId(id) {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	i, _ := strconv.Atoi(id)

	task, err := h.svc.GetTaskByID(c, i, int(authID))

	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			c.JSON(http.StatusNotFound, entities.ErrNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, entities.ErrInternalServerError)
		return
	}

	c.JSON(http.StatusOK, task.ToTaskDTO())
}

func (h *TaskHandler) HandleUpdateTask(c *gin.Context) {
	authID := GetUserIDFromContext(c)

	var in entities.Task
	id := c.Param("id")

	if checkId(id) {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	taskID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	task, err := h.svc.UpdateTask(c, in, taskID, int(authID))

	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			c.JSON(http.StatusNotFound, entities.ErrNotFound)
		}
		c.JSON(http.StatusInternalServerError, entities.ErrDatabaseFailed)
		return
	}

	c.JSON(http.StatusOK, task.ToTaskDTO())
}

func (h *TaskHandler) HandleDeleteTask(c *gin.Context) {
	authID := GetUserIDFromContext(c)

	id := c.Param("id")

	if checkId(id) {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	i, _ := strconv.Atoi(id)

	err := h.svc.DeleteTask(c, i, int(authID))

	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			c.JSON(http.StatusNotFound, entities.ErrNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, entities.ErrInternalServerError)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h *TaskHandler) HandleCompleteTask(c *gin.Context) {
	authID := GetUserIDFromContext(c)

	id := c.Param("id")

	if checkId(id) {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	i, _ := strconv.Atoi(id)

	rows, err := h.svc.MarkTaskAsDone(c, i, int(authID))

	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			c.JSON(http.StatusNotFound, entities.ErrNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, entities.ErrInternalServerError)
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

func SanitizeInput(task domain.Task) error {
	if len(task.Title) <= 0 {
		return domain.ErrBadData
	}

	return nil
}
