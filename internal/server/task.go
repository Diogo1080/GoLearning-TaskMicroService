package server

import (
	"backendGo/internal/entities"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) HandleAddTask() gin.HandlerFunc {
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

		h.TaskService.CreateTask(in)

	}
}
