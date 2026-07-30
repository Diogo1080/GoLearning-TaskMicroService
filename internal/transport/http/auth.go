package http

import (
	"backendGo/internal/auth"
	entities "backendGo/internal/domain"
	"backendGo/internal/store"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthServicePort interface {
	AuthenticateUser(userDto entities.User) (entities.UserDTO, error)
}

type AuthHandler struct {
	svc AuthServicePort // ← interface, not concrete type
	rds *store.Redis
}

func NewAuthHandler(svc AuthServicePort, rds *store.Redis) *AuthHandler {
	return &AuthHandler{svc: svc, rds: rds}
}

func getAuthenticatedUserID(c *gin.Context) (int, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return 0, false
	}

	uid, ok := val.(int)
	if !ok {
		return 0, false
	}

	return uid, true
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var in entities.User

	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	//Authenticate user
	user, err := h.svc.AuthenticateUser(in)
	if err != nil {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	fmt.Print(user.ID)
	//Issue token
	toks, err := auth.IssueTokens(strconv.Itoa(user.ID))
	if err != nil {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	//persist token in redis
	if err := auth.Persist(c, h.rds, toks); err != nil {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	auth.SetAuthCookies(c, toks)
	c.JSON(http.StatusOK, gin.H{"message": toks})
}
