package tests

import (
	entities "backendGo/internal/domain"
	"backendGo/internal/service"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository implements the UserRepository interface for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByID(id int) (entities.User, error) {
	args := m.Called(id)
	return args.Get(0).(entities.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(username string) (entities.User, error) {
	args := m.Called(username)
	return args.Get(0).(entities.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user entities.User, id int) (entities.User, error) {
	args := m.Called(user, id)
	return args.Get(0).(entities.User), args.Error(1)
}

func (m *MockUserRepository) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// TestNewUserService tests the constructor with validation
func TestNewUserService(t *testing.T) {
	t.Run("returns nil when repo is nil", func(t *testing.T) {
		svc := service.NewUserService(nil)
		assert.Nil(t, svc)
	})

	t.Run("returns valid service when repo is provided", func(t *testing.T) {
		mockRepo := &MockUserRepository{}
		svc := service.NewUserService(mockRepo)
		assert.NotNil(t, svc)
	})
}

// TestGetUserByID tests fetching user by ID
func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		setupMocks    func(m *MockUserRepository)
		expectSuccess bool
		expectedError error
	}{
		{
			name: "successfully retrieves user",
			id:   1,
			setupMocks: func(m *MockUserRepository) {
				m.On("GetUserByID", 1).Return(entities.User{
					ID:       1,
					Username: "founduser",
					Password: "hashed",
				}, nil)
			},
			expectSuccess: true,
			expectedError: nil,
		},
		{
			name: "fails when user not found",
			id:   999,
			setupMocks: func(m *MockUserRepository) {
				m.On("GetUserByID", 999).Return(entities.User{}, entities.ErrNotFound)
			},
			expectSuccess: false,
			expectedError: entities.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMocks(mockRepo)

			svc := service.NewUserService(mockRepo)
			result, err := svc.GetUserByID(tt.id)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.Equal(t, tt.id, result.ID)
			} else {
				assert.Error(t, err)
				assert.Empty(t, result.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetUserByUsername tests fetching user by username
func TestGetUserByUsername(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		setupMocks    func(m *MockUserRepository)
		expectSuccess bool
		expectedError error
	}{
		{
			name:     "successfully retrieves user by username",
			username: "testuser",
			setupMocks: func(m *MockUserRepository) {
				m.On("GetUserByUsername", "testuser").Return(entities.User{
					ID:       1,
					Username: "testuser",
					Password: "hashed",
				}, nil)
			},
			expectSuccess: true,
			expectedError: nil,
		},
		{
			name:     "fails when user not found by username",
			username: "nonexistent",
			setupMocks: func(m *MockUserRepository) {
				m.On("GetUserByUsername", "nonexistent").Return(entities.User{}, entities.ErrNotFound)
			},
			expectSuccess: false,
			expectedError: entities.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMocks(mockRepo)

			svc := service.NewUserService(mockRepo)
			result, err := svc.GetUserByUsername(tt.username)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.Equal(t, tt.username, result.Username)
			} else {
				assert.Error(t, err)
				assert.Empty(t, result.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestDeleteUser tests user deletion
func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		setupMocks    func(m *MockUserRepository)
		expectSuccess bool
		expectedError error
	}{
		{
			name: "successfully deletes user",
			id:   1,
			setupMocks: func(m *MockUserRepository) {
				m.On("DeleteUser", 1).Return(nil)
			},
			expectSuccess: true,
			expectedError: nil,
		},
		{
			name: "fails when repo delete errors",
			id:   1,
			setupMocks: func(m *MockUserRepository) {
				m.On("DeleteUser", 1).Return(errors.New("constraint violation"))
			},
			expectSuccess: false,
			expectedError: errors.New("constraint violation"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMocks(mockRepo)

			svc := service.NewUserService(mockRepo)
			err := svc.DeleteUser(tt.id)

			if tt.expectSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name          string
		inputUser     entities.User
		id            int
		setupMocks    func(m *MockUserRepository)
		expectedError error
	}{
		{
			name: "fails when repo update errors",
			inputUser: entities.User{
				Password: "newpassword",
			},
			id: 1,
			setupMocks: func(m *MockUserRepository) {
				m.On("UpdateUser", mock.Anything, 1).Return(entities.User{}, errors.New("update failed"))
			},
			expectedError: errors.New("update failed"),
		},
		{
			name: "fails when repo returns error for invalid id",
			inputUser: entities.User{
				Password: "anotherpass",
			},
			id: 999,
			setupMocks: func(m *MockUserRepository) {
				m.On("UpdateUser", mock.Anything, 999).Return(entities.User{}, entities.ErrNotFound)
			},
			expectedError: entities.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMocks(mockRepo)

			svc := service.NewUserService(mockRepo)
			result, err := svc.UpdateUser(tt.inputUser, tt.id)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
			assert.Empty(t, result.ID)

			mockRepo.AssertExpectations(t)
		})
	}
}
