package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"backendGo/internal/auth"
	"backendGo/internal/store"

	"github.com/gin-gonic/gin"
)

func bearerFromHeader(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

func AuthMiddleware(r *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, _ := c.Cookie("access_token")
		if tokenStr == "" {
			tokenStr = bearerFromHeader(c)
		}
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		claims, err := auth.ParseAccess(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		ctx := context.Background()
		if _, err := r.GetUserByJTI(ctx, "access:"+claims.ID); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}

		c.Set("userID", claims.Subject)
		c.Next()
	}
}

func MustCookie(c *gin.Context, name string) (string, error) {
	val, err := c.Cookie(name)
	if err != nil || val == "" {
		return "", errors.New("missing cookie: " + name)
	}
	return val, nil
}
