package server

import (
	entities "backendGo/internal/domain"
	"backendGo/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskServicePort interface {
	CreateTask(task entities.Task) (entities.Task, error)
	GetTasks(terms entities.TaskSearch, limit int) ([]entities.Task, error)
	GetTaskByID(id int) (entities.Task, error)
	UpdateTask(task entities.Task, id int) (entities.Task, error)
	DeleteTask(id int) error
	MarkTaskAsDone(id int) error
}

type TaskHandler struct {
	svc TaskServicePort // ← interface, not concrete type
}

func NewTaskHandler(svc TaskServicePort) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) HandleAddTask(c *gin.Context) {
	var in entities.Task
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	task, err := h.svc.CreateTask(in)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err.Error())
		return
	}

	c.JSON(http.StatusCreated, task.ToTaskDTO())
}

func (h *TaskHandler) HandleGetTasks(c *gin.Context) {
	terms := entities.TaskSearch{
		Search:     c.Request.URL.Query().Get("search"),
		Priority:   c.Request.URL.Query().Get("priority"),
		Completed:  c.Request.URL.Query().Get("completed"),
		DueDateMin: c.Request.URL.Query().Get("dueDataMin"),
		DueDateMax: c.Request.URL.Query().Get("dueDataMax"),
	}

	limit := 100

	tasks, err := h.svc.GetTasks(terms, limit)

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

	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	i, _ := strconv.Atoi(id)

	task, err := h.svc.GetTaskByID(i)
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
	var in entities.Task
	id := c.Param("id")

	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	fmt.Print(in)
	idint, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	task, err := h.svc.UpdateTask(in, idint)

	if err != nil {
		return
	}

	c.JSON(http.StatusOK, task.ToTaskDTO())
}

func (h *TaskHandler) HandleDeleteTask(c *gin.Context) {
	id := c.Param("id")

	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	i, _ := strconv.Atoi(id)

	err := h.svc.DeleteTask(i)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h *TaskHandler) HandleCompleteTask(c *gin.Context) {
	id := c.Param("id")

	if len(id) == 0 {
		utils.SendErrorResponse(c.Writer, "Id is not set", http.StatusBadRequest)
		return
	}

	i, _ := strconv.Atoi(id)

	err := h.svc.MarkTaskAsDone(i)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}
