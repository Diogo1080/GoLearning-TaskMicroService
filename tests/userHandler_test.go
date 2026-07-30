package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	entities "backendGo/internal/domain"
	server "backendGo/internal/transport/http"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService implements UserServicePort for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(user entities.User) (entities.UserDTO, error) {
	args := m.Called(user)
	return args.Get(0).(entities.UserDTO), args.Error(1)
}

func (m *MockUserService) GetUserByID(id int) (entities.UserDTO, error) {
	args := m.Called(id)
	return args.Get(0).(entities.UserDTO), args.Error(1)
}

func (m *MockUserService) GetUserByUsername(username string) (entities.UserDTO, error) {
	args := m.Called(username)
	return args.Get(0).(entities.UserDTO), args.Error(1)
}

func (m *MockUserService) UpdateUser(user entities.User, id int) (entities.UserDTO, error) {
	args := m.Called(user, id)
	return args.Get(0).(entities.UserDTO), args.Error(1)
}

func (m *MockUserService) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleGetUserByUsername(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		authUserID     int
		setAuth        bool
		setupMocks     func(m *MockUserService)
		expectedStatus int
	}{
		{
			name:           "returns 401 when not authenticated",
			username:       "testuser",
			setAuth:        false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "returns 400 when username is empty",
			username:       "",
			authUserID:     1,
			setAuth:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 403 when looking up another user",
			username:   "otheruser",
			authUserID: 1,
			setAuth:    true,
			setupMocks: func(m *MockUserService) {
				m.On("GetUserByUsername", "otheruser").Return(entities.UserDTO{
					ID:       2,
					Username: "otheruser",
				}, nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "returns 200 when looking up own username",
			username:   "myuser",
			authUserID: 1,
			setAuth:    true,
			setupMocks: func(m *MockUserService) {
				m.On("GetUserByUsername", "myuser").Return(entities.UserDTO{
					ID:       1,
					Username: "myuser",
				}, nil)
			},
			expectedStatus: http.StatusFound,
		},
		{
			name:       "returns 404 when user not found",
			username:   "nonexistent",
			authUserID: 1,
			setAuth:    true,
			setupMocks: func(m *MockUserService) {
				m.On("GetUserByUsername", "nonexistent").Return(entities.UserDTO{}, entities.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "returns 500 for other service errors",
			username:   "erruser",
			authUserID: 1,
			setAuth:    true,
			setupMocks: func(m *MockUserService) {
				m.On("GetUserByUsername", "erruser").Return(entities.UserDTO{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockUserService{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockSvc)
			}

			handler := server.NewUserHandler(mockSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.username != "" {
				c.Params = gin.Params{{Key: "username", Value: tt.username}}
			}
			c.Request, _ = http.NewRequest("GET", "/users/"+tt.username, nil)

			if tt.setAuth {
				c.Set("userID", tt.authUserID)
			}

			handler.HandleGetUserByUsername(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.setupMocks != nil {
				mockSvc.AssertExpectations(t)
			}
		})
	}
}

func TestHandleGetUserByID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		authUserID     int
		setAuth        bool
		setupMocks     func(m *MockUserService)
		expectedStatus int
	}{
		{
			name:           "returns 401 when not authenticated",
			id:             "1",
			setAuth:        false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "returns 400 when ID is invalid",
			id:             "abc",
			authUserID:     1,
			setAuth:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "returns 400 when ID is negative",
			id:             "-1",
			authUserID:     1,
			setAuth:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "returns 403 when requesting another user's ID",
			id:             "2",
			authUserID:     1,
			setAuth:        true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "returns 200 when requesting own ID",
			id:         "1",
			authUserID: 1,
			setAuth:    true,
			setupMocks: func(m *MockUserService) {
				m.On("GetUserByID", 1).Return(entities.UserDTO{
					ID:       1,
					Username: "testuser",
				}, nil)
			},
			expectedStatus: http.StatusFound,
		},
		{
			name:       "returns 500 when service returns error",
			id:         "1",
			authUserID: 1,
			setAuth:    true,
			setupMocks: func(m *MockUserService) {
				m.On("GetUserByID", 1).Return(entities.UserDTO{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockUserService{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockSvc)
			}

			handler := server.NewUserHandler(mockSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.id}}
			c.Request, _ = http.NewRequest("GET", "/users/"+tt.id, nil)

			if tt.setAuth {
				c.Set("userID", tt.authUserID)
			}

			handler.HandleGetUserByID(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.setupMocks != nil {
				mockSvc.AssertExpectations(t)
			}
		})
	}
}

func TestHandlePasswordChange(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		authUserID     int
		setAuth        bool
		jsonPayload    string
		setupMocks     func(m *MockUserService)
		expectedStatus int
	}{
		{
			name:           "returns 401 when not authenticated",
			id:             "1",
			setAuth:        false,
			jsonPayload:    `{"password":"newpass"}`,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "returns 400 when ID is invalid",
			id:             "abc",
			authUserID:     1,
			setAuth:        true,
			jsonPayload:    `{"password":"newpass"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "returns 400 when JSON is invalid",
			id:             "1",
			authUserID:     1,
			setAuth:        true,
			jsonPayload:    `{invalid}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "returns 403 when changing another user's password",
			id:             "2",
			authUserID:     1,
			setAuth:        true,
			jsonPayload:    `{"password":"newpass"}`,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "returns 200 when changing own password",
			id:          "1",
			authUserID:  1,
			setAuth:     true,
			jsonPayload: `{"password":"newpass"}`,
			setupMocks: func(m *MockUserService) {
				m.On("UpdateUser", mock.MatchedBy(func(u entities.User) bool {
					return u.Password == "newpass"
				}), 1).Return(entities.UserDTO{
					ID:       1,
					Username: "testuser",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "returns 500 when service returns error",
			id:          "1",
			authUserID:  1,
			setAuth:     true,
			jsonPayload: `{"password":"newpass"}`,
			setupMocks: func(m *MockUserService) {
				m.On("UpdateUser", mock.Anything, 1).Return(entities.UserDTO{}, errors.New("update failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockUserService{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockSvc)
			}

			handler := server.NewUserHandler(mockSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.id}}
			c.Request, _ = http.NewRequest("PUT", "/users/"+tt.id+"/password", strings.NewReader(tt.jsonPayload))
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.setAuth {
				c.Set("userID", tt.authUserID)
			}

			handler.HandlePasswordChange(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.setupMocks != nil {
				mockSvc.AssertExpectations(t)
			}
		})
	}
}

func TestHandleDeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		authUserID     int
		setAuth        bool
		setupMocks     func(m *MockUserService)
		expectedStatus int
	}{
		{
			name:           "returns 401 when not authenticated",
			id:             "1",
			setAuth:        false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "returns 400 when ID is invalid",
			id:             "abc",
			authUserID:     1,
			setAuth:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "returns 400 when ID is negative",
			id:             "-1",
			authUserID:     1,
			setAuth:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "returns 403 when deleting another user",
			id:             "2",
			authUserID:     1,
			setAuth:        true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "returns 200 when deleting own account",
			id:         "1",
			authUserID: 1,
			setAuth:    true,
			setupMocks: func(m *MockUserService) {
				m.On("DeleteUser", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "returns 500 when service errors on deleting own account",
			id:         "1",
			authUserID: 1,
			setAuth:    true,
			setupMocks: func(m *MockUserService) {
				m.On("DeleteUser", 1).Return(errors.New("constraint violation"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockUserService{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockSvc)
			}

			handler := server.NewUserHandler(mockSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.id}}
			c.Request, _ = http.NewRequest("DELETE", "/users/"+tt.id, nil)

			if tt.setAuth {
				c.Set("userID", tt.authUserID)
			}

			handler.HandleDeleteUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.setupMocks != nil {
				mockSvc.AssertExpectations(t)
			}
		})
	}
}
