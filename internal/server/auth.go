package server

import (
	"backendGo/internal/auth"
	"backendGo/internal/entities"
	"backendGo/internal/store"
	"backendGo/models"
	"backendGo/utils"
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
		toks, err := auth.IssueTokens(strconv.FormatInt(user.ID, 10))
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

func (h *Handlers) HandleCreateUser(res http.ResponseWriter, req *http.Request) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in entities.UserDTO
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid json"})
			return
		}

		user, err := h.AuthService.Repo.CreateUser(in.ToUser())

		if err != nil {
			utils.SendErrorResponse()
		}

		sendJSONResponse(res, http.StatusOK, models.IDResponse{ID: user.ID})
	}

}

func (h *Handlers) HandleGetUserByID(username string) (entities.User, error) {
	return h.AuthService.Repo.GetUserByUsername(username)
}

func (h *Handlers) HandleUpdateUser(user map[string]entities.User) (int64, error) {
	return h.AuthService.Repo.UpdateUser(user)
}

func (h *Handlers) HandleDeleteUser(id int) (int64, error) {
	return h.AuthService.Repo.DeleteUser(id)
}
