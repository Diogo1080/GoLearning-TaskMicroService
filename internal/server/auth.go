package server

import (
	"backendGo/internal/auth"
	"backendGo/internal/entities"
	"backendGo/models"
	"backendGo/utils"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) HandleLogin(c *gin.Context) {
	var in entities.UserDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid json"})
		return
	}

	//Authenticate user
	user, err := h.AuthService.AuthenticateUser(in)
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
	if err := auth.Persist(c, h.rds, toks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not persist tokens"})
		return
	}

	auth.SetAuthCookies(c, toks)
	c.JSON(http.StatusOK, gin.H{"message": toks})
}

func (h *Handlers) HandleCreateUser(c *gin.Context) {

	var in entities.UserDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid json"})
		utils.SendErrorResponse(c.Writer, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := h.AuthService.CreateNewUser(in)

	if err != nil {
		utils.SendErrorResponse(c.Writer, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, models.IDResponse{ID: user.ID})
}

func (h *Handlers) HandleGetUserByUsername(username string) (entities.User, error) {
	return h.AuthService.Repo.GetUserByUsername(username)
}

func (h *Handlers) HandleGetUserByID(ID int) (entities.User, error) {
	return h.AuthService.Repo.GetUserByID(ID)
}

func (h *Handlers) HandleUpdateUser(user map[string]entities.User) (int64, error) {
	return h.AuthService.Repo.UpdateUser(user)
}

func (h *Handlers) HandleDeleteUser(id int) (int64, error) {
	return h.AuthService.Repo.DeleteUser(id)
}
