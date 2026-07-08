package server

import (
	"backendGo/internal/auth"
	"backendGo/internal/entities"
	"backendGo/internal/store"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) HandleLogin(rds *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in entities.UserDTO
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid json"})
			return
		}

		//Authenticate user
		user, err := h.AuthService.AuthenticateUser(in.Username, in.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		//Issue token
		toks, err := auth.IssueTokens(strconv.Itoa(user.ID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue tokens"})
			return
		}

		//persist token in redis
		if err := auth.Persist(c, rds, toks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not persist tokens"})
			return
		}

		auth.SetAuthCookies(c, toks)
		c.JSON(http.StatusOK, gin.H{"message": "login successful"})
	}
}
