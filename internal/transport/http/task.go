package http

import (
	entities "backendGo/internal/domain"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskServicePort interface {
	CreateTask(task entities.Task) (entities.Task, error)
	GetTasks(terms entities.TaskSearch) ([]entities.Task, error)
	GetTaskByID(id int, userID int) (entities.Task, error)
	UpdateTask(task entities.Task, taskID int, userID int) (entities.Task, error)
	DeleteTask(id int, userID int) error
	MarkTaskAsDone(id int, userID int) (int, error)
}

type TaskHandler struct {
	svc TaskServicePort
}

func NewTaskHandler(svc TaskServicePort) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) HandleAddTask(c *gin.Context) {
	authID, ok := getUserID(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	var in entities.Task
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	in.UserID = int64(authID)

	task, err := h.svc.CreateTask(in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, task.ToTaskDTO())
}

func (h *TaskHandler) HandleGetTasks(c *gin.Context) {
	authID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	terms := entities.TaskSearch{
		Search:     c.Request.URL.Query().Get("search"),
		UserID:     int(authID),
		Priority:   c.Request.URL.Query().Get("priority"),
		Completed:  c.Request.URL.Query().Get("completed"),
		DueDateMin: c.Request.URL.Query().Get("dueDataMin"),
		DueDateMax: c.Request.URL.Query().Get("dueDataMax"),
		OrderBy:    c.Request.URL.Query().Get("orderBy"),
		Limit:      100,
	}

	tasks, err := h.svc.GetTasks(terms)

	if err != nil {
		if errors.Is(err, entities.ErrBadData) {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if len(tasks) == 0 {
		tasks = []entities.Task{}
		c.JSON(http.StatusOK, tasks)
		return
	}

	c.JSON(http.StatusFound, tasks)
}

func (h *TaskHandler) HandleGetTaskById(c *gin.Context) {
	id := c.Param("id")

	authID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	if checkId(id) {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	i, _ := strconv.Atoi(id)

	task, err := h.svc.GetTaskByID(i, authID)

	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			c.JSON(http.StatusNotFound, err)
			return
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusFound, task.ToTaskDTO())
}

func (h *TaskHandler) HandleUpdateTask(c *gin.Context) {
	authID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

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

	task, err := h.svc.UpdateTask(in, taskID, authID)

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
	authID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	id := c.Param("id")

	if checkId(id) {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	i, _ := strconv.Atoi(id)

	err := h.svc.DeleteTask(i, authID)

	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			c.JSON(http.StatusNotFound, err)
			return
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h *TaskHandler) HandleCompleteTask(c *gin.Context) {
	authID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	id := c.Param("id")

	if checkId(id) {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	i, _ := strconv.Atoi(id)

	rows, err := h.svc.MarkTaskAsDone(i, authID)

	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			c.JSON(http.StatusNotFound, err)
			return
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, rows)
}
