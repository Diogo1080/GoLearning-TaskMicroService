package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httptransport "github.com/Diogo1080/GoLearning-TaskMicroService/internal/transport/http"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutesExposesHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := httptransport.RegisterRoutes(gin.New(), nil, rejectUnauthenticatedRequests)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestRegisterRoutesExposesLivenessAndReadinessEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := httptransport.RegisterRoutes(gin.New(), nil, rejectUnauthenticatedRequests, func(context.Context) error {
		return nil
	})

	for _, path := range []string{"/livez", "/readyz", "/metrics"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code, path)
	}
}

func TestRegisterRoutesReturnsUnavailableWhenDependencyIsNotReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := httptransport.RegisterRoutes(gin.New(), nil, rejectUnauthenticatedRequests, func(context.Context) error {
		return errors.New("dependency unavailable")
	})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestRegisterRoutesProtectsTaskEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := httptransport.RegisterRoutes(gin.New(), httptransport.NewTaskHandler(nil), rejectUnauthenticatedRequests)

	req := httptest.NewRequest(http.MethodGet, "/api/task", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func rejectUnauthenticatedRequests(c *gin.Context) {
	c.AbortWithStatus(http.StatusUnauthorized)
}
