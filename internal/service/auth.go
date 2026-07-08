package service

import (
	"backendGo/internal/entities"
	"backendGo/internal/store"
	"fmt"
)

type AuthService struct {
	Repo *store.SQLiteUserRepository
}

func NewAuthService(repo *store.SQLiteUserRepository) *AuthService {
	return &AuthService{Repo: repo}
}

func (s *AuthService) CreateUser(userDTO entities.UserDTO) (entities.User, error) {
	user := userDTO.ToUser()
	return s.Repo.CreateUser(user)
}

func (s *AuthService) GetUserByID(username string) (entities.User, error) {
	return s.Repo.GetUserByUsername(username)
}

func (s *AuthService) UpdateUser(user entities.User) (entities.User, error) {
	return s.Repo.UpdateUser(user)
}

func (s *AuthService) DeleteUser(id int) error {
	return s.Repo.DeleteUser(id)
}

func (s *AuthService) AuthenticateUser(username, password string) (entities.User, error) {
	user, err := s.Repo.GetUserByUsername(username)
	if err != nil {
		return entities.User{}, err
	}
	if user.Password != password {
		return entities.User{}, fmt.Errorf("invalid credentials")
	}
	return user, nil
}
