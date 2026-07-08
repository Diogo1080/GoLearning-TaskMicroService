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
