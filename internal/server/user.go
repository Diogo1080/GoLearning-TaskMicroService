package server

import (
	entities "backendGo/internal/domain"
	"backendGo/utils"
	"errors"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserServicePort interface {
	CreateUser(user entities.User) (entities.UserDTO, error)
	GetUserByID(id int) (entities.UserDTO, error)
	GetUserByUsername(username string) (entities.UserDTO, error)
	UpdateUser(user entities.User, id int) (entities.UserDTO, error)
	DeleteUser(id int) error
}

type UserHandler struct {
	svc UserServicePort // ← interface, not concrete type
}

func NewUserHandler(svc UserServicePort) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) HandleCreateUser(c *gin.Context) {
	var in entities.User
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	user, err := h.svc.CreateUser(in)

	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) HandleGetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	if len(username) == 0 {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	user, err := h.svc.GetUserByUsername(username)

	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			c.JSON(http.StatusNotFound, err)
		}

		c.JSON(http.StatusFailedDependency, err)
		return
	}

	c.JSON(http.StatusFound, user)
}

func (h *UserHandler) HandleGetUserByID(c *gin.Context) {
	id := c.Param("id")

	if len(id) == 0 {
		utils.SendErrorResponse(c.Writer, "Id is not set", http.StatusBadRequest)
		return
	}
	i, _ := strconv.Atoi(id)

	user, err := h.svc.GetUserByID(i)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, user)
}

func (h *UserHandler) HandlePasswordChange(c *gin.Context) {
	var in entities.User
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

	user, err := h.svc.UpdateUser(in, idint)

	if err != nil {
		utils.SendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(c.Writer, http.StatusOK, user)
}

func (h *UserHandler) HandleDeleteUser(id int) error {
	err := h.svc.DeleteUser(id)

	if err != nil {
		return err
	}

	return nil
}
