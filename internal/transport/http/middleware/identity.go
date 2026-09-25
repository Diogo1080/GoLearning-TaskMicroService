package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	Identity "github.com/Diogo1080/GoLearning-IdentityMicroService/api/v1"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/logger"

	"github.com/gin-gonic/gin"
)

type IdentityMiddlewareBuilder struct {
	IdentityClient TokenValidator
	logger         *slog.Logger
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*Identity.ValidateTokenResponse, error)
}

func NewIdentityMiddlewareBuilder(client TokenValidator, appEnv string) *IdentityMiddlewareBuilder {
	return &IdentityMiddlewareBuilder{IdentityClient: client, logger: logger.New(appEnv).WithGroup("IdentityMiddleware")}
}

func (b *IdentityMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := bearerFromHeader(c)

		if tokenStr == "" {
			tokenStr, _ = c.Cookie("access_token")
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrUnauthorized)
			return
		}

		ctx := c.Request.Context()
		resp, err := b.IdentityClient.ValidateToken(ctx, tokenStr)

		if err != nil || resp.UserId == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrUnauthorized)
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
