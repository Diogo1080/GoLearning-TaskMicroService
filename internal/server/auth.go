package server

import (
	"backendGo/internal/auth"
	"backendGo/internal/entities"
	"backendGo/utils"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) HandleLogin(c *gin.Context) {
	var in entities.CreateAndLoginUserDTO

	if err := c.ShouldBindJSON(&in); err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	//Authenticate user
	user, err := h.AuthService.AuthenticateUser(in)
	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	//Issue token
	toks, err := auth.IssueTokens(strconv.Itoa(user.ID))
	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	//persist token in redis
	if err := auth.Persist(c, h.rds, toks); err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	auth.SetAuthCookies(c, toks)
	c.JSON(http.StatusOK, gin.H{"message": toks})
}

func (h *Handlers) HandleCreateUser(c *gin.Context) {

	var in entities.CreateAndLoginUserDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.CreateNewUser(in)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, user)
}

func (h *Handlers) HandleGetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	if len(username) == 0 {
		utils.SendErrorResponse(c.Writer, "Username is not set", http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.Repo.GetUserByUsername(username)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, user.ToUserDTO())
}

func (h *Handlers) HandleGetUserByID(c *gin.Context) {
	id := c.Param("id")

	if len(id) == 0 {
		utils.SendErrorResponse(c.Writer, "Id is not set", http.StatusBadRequest)
		return
	}
	i, _ := strconv.Atoi(id)

	user, err := h.AuthService.Repo.GetUserByID(i)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, user.ToUserDTO())
}

func (h *Handlers) HandlePasswordChange(c *gin.Context) {
	var in entities.CreateAndLoginUserDTO
	id := c.Param("id")

	if err := c.ShouldBindJSON(&in); err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	idint, err := strconv.Atoi(id)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.UpdateUser(in, idint)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, user)
}

func (h *Handlers) HandleDeleteUser(id int) (int64, error) {
	return h.AuthService.Repo.DeleteUser(id)
}
