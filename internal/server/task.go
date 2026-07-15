package server

import (
	"backendGo/internal/entities"
	"backendGo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) HandleAddTask(c *gin.Context) {
	var in entities.TaskDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.SendErrorResponse(c.Writer, "Bad Json data", http.StatusBadRequest)
		return
	}

	task, err := h.TaskService.CreateTask(in)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, task)
}

func (h *Handlers) HandleGetTasks(c *gin.Context) {

	terms := entities.TaskSearch{
		Search:     c.Request.URL.Query().Get("search"),
		Priority:   c.Request.URL.Query().Get("priority"),
		Completed:  c.Request.URL.Query().Get("completed"),
		DueDateMin: c.Request.URL.Query().Get("minDate"),
		DueDateMax: c.Request.URL.Query().Get("maxDate"),
	}
	limit := 100

	tasks, err := h.TaskService.Repo.GetTasks(terms, limit)
	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(tasks) == 0 {
		tasks = []entities.Task{}
	}

	sendJSONResponse(c.Writer, http.StatusOK, map[string][]entities.Task{"tasks": tasks})

}

func (h *Handlers) HandleGetTaskById(c *gin.Context) {
	id := c.Param("id")

	if len(id) == 0 {
		utils.SendErrorResponse(c.Writer, "Id is not set", http.StatusBadRequest)
		return
	}

	i, _ := strconv.Atoi(id)

	userDTO, err := h.TaskService.Repo.GetTaskByID(i)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusNotFound)
		return
	}

	if !(len(userDTO.Title) > 0) {
		utils.SendErrorResponse(c.Writer, "No id match", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, userDTO)
}

func (h *Handlers) HandleUpdateTask(c *gin.Context) {
	var in entities.TaskDTO
	id := c.Param("id")

	if len(id) == 0 {
		utils.SendErrorResponse(c.Writer, "Id is not set", http.StatusBadRequest)
		return
	}

	if err := c.ShouldBindJSON(&in); err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	idint, err := strconv.Atoi(id)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	taskDTO, err := h.TaskService.UpdateTask(in, idint)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, taskDTO)

}

func (h *Handlers) HandleDeleteTask(c *gin.Context) {
	id := c.Param("id")

	if len(id) == 0 {
		utils.SendErrorResponse(c.Writer, "Id is not set", http.StatusBadRequest)
		return
	}

	i, _ := strconv.Atoi(id)

	numberRows, err := h.TaskService.Repo.DeleteTask(i)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, numberRows)
}

func (h *Handlers) HandleCompleteTask(c *gin.Context) {
	id := c.Param("id")

	if len(id) == 0 {
		utils.SendErrorResponse(c.Writer, "Id is not set", http.StatusBadRequest)
		return
	}

	i, _ := strconv.Atoi(id)

	numberRows, err := h.TaskService.Repo.MarkTaskAsDone(i)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, numberRows)
}
