package middleware

import (
	"context"
	"net/http"
	"strings"

	authClient "backendGo/internal/auth"

	"github.com/gin-gonic/gin"
)

type AuthMiddlewareBuilder struct {
	authClient *authClient.Client
}

func NewAuthMiddlewareBuilder(client *authClient.Client) *AuthMiddlewareBuilder {
	return &AuthMiddlewareBuilder{authClient: client}
}

func (b *AuthMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := bearerFromHeader(c)

		if tokenStr == "" {
			tokenStr, _ = c.Cookie("access_token")
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		ctx := context.Background()
		resp, err := b.authClient.ValidateToken(ctx, tokenStr)
		if err != nil || !resp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or revoked token"})
			return
		}

		c.Set("userID", int32(resp.UserId))
		c.Next()
	}
}

func bearerFromHeader(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
