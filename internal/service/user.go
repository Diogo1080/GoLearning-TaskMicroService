package service

import (
	entities "backendGo/internal/domain"
)

type UserRepository interface {
	CreateUser(User entities.User) (entities.User, error)
	GetUserByID(id int) (entities.User, error)
	GetUserByUsername(id string) (entities.User, error)
	UpdateUser(User entities.User, id int) (entities.User, error)
	DeleteUser(id int) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	// Validation
	if repo == nil {
		return nil // or panic, depending on policy
	}

	// Setup
	svc := &UserService{repo: repo}

	return svc
}

func (a *UserService) CreateUser(user entities.User) (entities.UserDTO, error) {

	//Check if already exists
	_, err := a.repo.GetUserByUsername(user.Username)

	if err == nil {
		return entities.UserDTO{}, entities.ErrAlreadyExists
	}

	user.Password, err = HashPassword(user.Password)

	if err != nil {
		return entities.UserDTO{}, err
	}

	user, err = a.repo.CreateUser(user)
	if err != nil {
		return entities.UserDTO{}, err
	}

	return user.ToUserDTO(), nil
}

func (a *UserService) GetUserByID(id int) (entities.UserDTO, error) {
	user, err := a.repo.GetUserByID(id)

	if err != nil {
		return entities.UserDTO{}, err
	}

	return user.ToUserDTO(), nil
}
func (a *UserService) GetUserByUsername(username string) (entities.UserDTO, error) {

	user, err := a.repo.GetUserByUsername(username)

	if err != nil {
		return entities.UserDTO{}, err
	}

	return user.ToUserDTO(), nil
}
func (a *UserService) DeleteUser(id int) error {
	err := a.repo.DeleteUser(id)

	if err != nil {
		return err
	}

	return nil
}

func (a *UserService) UpdateUser(user entities.User, id int) (entities.UserDTO, error) {
	user, err := a.repo.UpdateUser(user, id)

	if err != nil {
		return entities.UserDTO{}, err
	}

	return user.ToUserDTO(), nil
}
