package server

import (
	"backendGo/internal/entities"
	"backendGo/models"
	"backendGo/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) HandleAddTask(c *gin.Context) {
	var in entities.TaskDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.SendErrorResponse(c.Writer, "Error: Bad Json data", http.StatusBadRequest)
		return
	}

	task, err := h.TaskService.CreateTask(in)

	if err != nil {
		utils.SendErrorResponse(c.Writer, "Error:"+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, models.IDResponse{ID: task.ID})
}

func (h *Handlers) HandleGetTasks(c *gin.Context) {

	searchTerm := c.Request.URL.Query().Get("search")
	limit := 100

	tasks, err := h.TaskService.Repo.GetTasks(searchTerm, limit)
	if err != nil {
		utils.SendErrorResponse(c.Writer, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(tasks) == 0 {
		tasks = []entities.Task{}
	}

	sendJSONResponse(c.Writer, http.StatusOK, map[string][]entities.Task{"tasks": tasks})

}
