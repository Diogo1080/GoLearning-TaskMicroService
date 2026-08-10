package http

import (
	"context"
	"net/http"
	"time"

	authClient "backendGo/internal/auth"
	entities "backendGo/internal/domain"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authClient *authClient.Client
}

func NewAuthHandler(client *authClient.Client) *AuthHandler {
	return &AuthHandler{authClient: client}
}

func (h *AuthHandler) HandleRegister(c *gin.Context) {
	type RegisterInput struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Password string `json:"password" binding:"required,min=6"`
	}

	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.Register(ctx, input.Username, input.Password)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, entities.ErrServiceUnavailable)
		return
	}

	if !resp.Success {
		c.JSON(http.StatusConflict, entities.ErrConflict)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user_id": resp.UserId,
	})
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	type LoginInput struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.Login(ctx, input.Username, input.Password)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, entities.ErrServiceUnavailable)
		return
	}

	if resp == nil {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	// Set cookies for browser clients
	h.setAuthCookies(c, resp.AccessToken, resp.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user_id":       resp.UserId,
		"message":       resp.Message,
	})
}

func (h *AuthHandler) HandleLogout(c *gin.Context) {
	token, _ := c.Cookie("access_token")
	if token == "" {
		token = bearerFromHeader(c)
	}

	if token == "" {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err := h.authClient.Logout(ctx, token)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, entities.ErrServiceUnavailable)
		return
	}

	// Clear cookies
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *AuthHandler) HandleRefreshToken(c *gin.Context) {
	type RefreshInput struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var input RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrBadData)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.RefreshToken(ctx, input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, entities.ErrServiceUnavailable)
		return
	}

	if resp == nil {
		c.JSON(http.StatusUnauthorized, entities.ErrUnauthorized)
		return
	}

	h.setAuthCookies(c, resp.AccessToken, resp.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

// Helper functions
func (h *AuthHandler) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	// Calculate expiry times (matching auth service JWT expiry)
	expAcc := 15 * 60          // 15 minutes
	expRef := 7 * 24 * 60 * 60 // 7 days

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", accessToken, expAcc, "/", "", true, true)
	c.SetCookie("refresh_token", refreshToken, expRef, "/", "", true, true)
}

func bearerFromHeader(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if len(h) > 7 && h[:7] == "Bearer " {
		return h[7:]
	}
	return ""
}

func getUserID(c *gin.Context) (int, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	if userID, ok := val.(int32); ok {
		return int(userID), true
	}
	return 0, false
}
