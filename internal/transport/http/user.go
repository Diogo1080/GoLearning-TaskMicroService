package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	authClient "backendGo/internal/auth"
	entities "backendGo/internal/domain"
	"errors"

	"github.com/gin-gonic/gin"
)

type UserServicePort interface {
	GetUserByID(id int) (entities.UserDTO, error)
	GetUserByUsername(username string) (entities.UserDTO, error)
	UpdateUser(user entities.User, id int) (entities.UserDTO, error)
	DeleteUser(id int) error
}

type UserHandler struct {
	svc        UserServicePort
	authClient *authClient.Client
}

func NewUserHandler(svc UserServicePort, authClient *authClient.Client) *UserHandler {
	return &UserHandler{svc: svc, authClient: authClient}
}

func (h *UserHandler) HandleGetUserByUsername(c *gin.Context) {
	authID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	username := c.Param("username")
	if len(username) == 0 {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	user, err := h.svc.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			c.JSON(http.StatusNotFound, entities.ErrNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, entities.ErrDatabaseFailed)
		return
	}

	if user.ID != authID {
		c.JSON(http.StatusForbidden, entities.ErrUnauthorized)
		return
	}

	c.JSON(http.StatusOK, user) // Fixed: Was StatusFound (302)
}

func (h *UserHandler) HandleGetUserByID(c *gin.Context) {
	authID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	id := c.Param("id")

	if checkId(id) {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	i, _ := strconv.Atoi(id)

	if i != authID {
		c.JSON(http.StatusForbidden, entities.ErrUnauthorized)
		return
	}

	user, err := h.svc.GetUserByID(i)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.ErrDatabaseFailed)
		return
	}

	c.JSON(http.StatusOK, user) // Fixed: Was StatusFound (302)
}
func (h *UserHandler) HandleChangePassword(c *gin.Context) {
	authID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	var input struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.ChangePassword(ctx, authID, input.CurrentPassword, input.NewPassword)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, entities.ErrServiceUnavailable)
		return
	}

	if !resp.Success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
func (h *UserHandler) HandleDeleteUser(c *gin.Context) {
	authID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	id := c.Param("id")

	if checkId(id) {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	idint, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	if idint != authID {
		c.JSON(http.StatusForbidden, entities.ErrUnauthorized)
		return
	}

	if err := h.svc.DeleteUser(idint); err != nil {
		c.JSON(http.StatusInternalServerError, entities.ErrDatabaseFailed)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// HandleUpdateProfile updates non-credential fields
func (h *UserHandler) HandleUpdateProfile(c *gin.Context) {
	authID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	var input struct {
		Username string `json:"username"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	user := entities.User{
		Username: input.Username,
	}

	updated, err := h.svc.UpdateUser(user, authID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.ErrDatabaseFailed)
		return
	}

	c.JSON(http.StatusOK, updated)
}
