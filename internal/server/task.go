package server

import (
	"backendGo/internal/entities"
	"backendGo/models"
	"backendGo/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) HandleAddTask(res http.ResponseWriter, req *http.Request) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in entities.TaskDTO
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid json"})
			return
		}

		_, exists := c.Get("userID")
		if exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "userID found in context"})
			return
		}

		task, err := h.TaskService.CreateTask(in)

		if err != nil {
			utils.SendErrorResponse(res, "Error:", http.StatusInternalServerError)
		}

		sendJSONResponse(res, http.StatusOK, models.IDResponse{ID: task.ID})
	}
}

func (h *Handlers) HandleGetTasks(res http.ResponseWriter, req *http.Request) gin.HandlerFunc {
	return func(c *gin.Context) {

		searchTerm := req.URL.Query().Get("search")
		limit := 100

		tasks, err := h.TaskService.Repo.GetTasks(searchTerm, limit)
		if err != nil {
			utils.SendErrorResponse(res, "Error:", http.StatusInternalServerError)
		}

		if len(tasks) == 0 {
			tasks = []entities.Task{}
		}

		sendJSONResponse(res, http.StatusOK, map[string][]entities.Task{"tasks": tasks})
	}
}
