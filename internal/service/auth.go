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

func (a *AuthService) AuthenticateUser(userDto entities.UserDTO) (entities.User, error) {
	user, err := a.Repo.GetUserByUsername(userDto.Username)
	if err != nil {
		return entities.User{}, err
	}

	//TODO Encrypt password in the future
	if user.Password != userDto.Password {
		return entities.User{}, fmt.Errorf("invalid credentials")
	}
	return user, nil
}

func (a *AuthService) CreateNewUser(userDto entities.UserDTO) (entities.User, error) {
	var user entities.User

	//Check if already exists
	user, err := a.Repo.GetUserByUsername(userDto.Username)

	if err == nil {
		return entities.User{}, fmt.Errorf("Already exists")
	}

	user = userDto.ToUser()

	//TODO Encrypt password in the future

	user, err = a.Repo.CreateUser(user)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}
