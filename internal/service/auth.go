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

func (a *AuthService) AuthenticateUser(userDto entities.CreateAndLoginUserDTO) (entities.UserDTO, error) {
	user, err := a.Repo.GetUserByUsername(userDto.Username)
	if err != nil {
		return entities.UserDTO{}, err
	}

	userDto.Password, err = HashPassword(user.Password)

	if err != nil {
		return entities.UserDTO{}, err
	}

	if VerifyPassword(userDto.Password, user.Password) {
		return entities.UserDTO{}, fmt.Errorf("invalid credentials")
	}

	return user.ToUserDTO(), nil
}

func (a *AuthService) CreateNewUser(userDto entities.CreateAndLoginUserDTO) (entities.UserDTO, error) {
	var user entities.User

	//Check if already exists
	user, err := a.Repo.GetUserByUsername(userDto.Username)

	if err == nil {
		return entities.UserDTO{}, fmt.Errorf("Already exists")
	}

	user = userDto.ToUser()

	user.Password, err = HashPassword(user.Password)

	if err != nil {
		return entities.UserDTO{}, err
	}

	user, err = a.Repo.CreateUser(user)
	if err != nil {
		return entities.UserDTO{}, err
	}

	return user.ToUserDTO(), nil
}

func (a *AuthService) UpdateUser(userDTO entities.CreateAndLoginUserDTO, id int) (entities.UserDTO, error) {
	newUser := userDTO.ToUser()

	user, err := a.Repo.UpdateUser(newUser, id)

	if err != nil {
		return entities.UserDTO{}, err
	}

	return user, nil
}
