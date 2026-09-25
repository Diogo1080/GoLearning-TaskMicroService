package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	identity "github.com/Diogo1080/GoLearning-IdentityMicroService/api/v1"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeTokenValidator struct {
	validate func(context.Context, string) (*identity.ValidateTokenResponse, error)
}

func (f fakeTokenValidator) ValidateToken(ctx context.Context, token string) (*identity.ValidateTokenResponse, error) {
	return f.validate(ctx, token)
}

func TestIdentityMiddlewareAcceptsBearerToken(t *testing.T) {
	validator := fakeTokenValidator{validate: func(ctx context.Context, token string) (*identity.ValidateTokenResponse, error) {
		assert.Equal(t, "bearer-token", token)
		return &identity.ValidateTokenResponse{UserId: 42}, nil
	}}

	recorder := runIdentityMiddlewareRequest(t, validator, "Bearer bearer-token", "", context.Background())

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "42", recorder.Body.String())
}

func TestIdentityMiddlewareAcceptsCookieToken(t *testing.T) {
	validator := fakeTokenValidator{validate: func(ctx context.Context, token string) (*identity.ValidateTokenResponse, error) {
		assert.Equal(t, "cookie-token", token)
		return &identity.ValidateTokenResponse{UserId: 7}, nil
	}}

	recorder := runIdentityMiddlewareRequest(t, validator, "", "cookie-token", context.Background())

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "7", recorder.Body.String())
}

func TestIdentityMiddlewareRejectsMissingToken(t *testing.T) {
	called := false
	validator := fakeTokenValidator{validate: func(ctx context.Context, token string) (*identity.ValidateTokenResponse, error) {
		called = true
		return nil, nil
	}}

	recorder := runIdentityMiddlewareRequest(t, validator, "", "", context.Background())

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.False(t, called)
}

func TestIdentityMiddlewareRejectsInvalidTokenResponse(t *testing.T) {
	validator := fakeTokenValidator{validate: func(ctx context.Context, token string) (*identity.ValidateTokenResponse, error) {
		return &identity.ValidateTokenResponse{}, nil
	}}

	recorder := runIdentityMiddlewareRequest(t, validator, "Bearer invalid", "", context.Background())

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestIdentityMiddlewareRejectsValidatorError(t *testing.T) {
	validator := fakeTokenValidator{validate: func(ctx context.Context, token string) (*identity.ValidateTokenResponse, error) {
		return nil, errors.New("identity unavailable")
	}}

	recorder := runIdentityMiddlewareRequest(t, validator, "Bearer unavailable", "", context.Background())

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestIdentityMiddlewarePropagatesRequestCancellation(t *testing.T) {
	requestContext, cancel := context.WithCancel(context.Background())
	cancel()

	validator := fakeTokenValidator{validate: func(ctx context.Context, token string) (*identity.ValidateTokenResponse, error) {
		assert.ErrorIs(t, ctx.Err(), context.Canceled)
		return nil, ctx.Err()
	}}

	recorder := runIdentityMiddlewareRequest(t, validator, "Bearer cancelled", "", requestContext)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func runIdentityMiddlewareRequest(t *testing.T, validator middleware.TokenValidator, authorization, cookie string, requestContext context.Context) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.NewIdentityMiddlewareBuilder(validator, "test").Build())
	router.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "%d", c.GetInt32("userID"))
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil).WithContext(requestContext)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	if cookie != "" {
		request.AddCookie(&http.Cookie{Name: "access_token", Value: cookie})
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
