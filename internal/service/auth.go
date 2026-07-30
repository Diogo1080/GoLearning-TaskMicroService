package service

import (
	entities "backendGo/internal/domain"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthReader interface {
	GetUserByUsername(id string) (entities.User, error)
}

type AuthService struct {
	repo AuthReader
}

func NewAuthService(repo AuthReader) *AuthService {
	// Validation
	if repo == nil {
		return nil // or panic, depending on policy
	}

	// Setup
	svc := &AuthService{repo: repo}

	return svc
}

func (a *AuthService) AuthenticateUser(user entities.User) (entities.UserDTO, error) {
	existing, err := a.repo.GetUserByUsername(user.Username)

	if err != nil {
		return entities.UserDTO{}, err
	}

	if !VerifyPassword(user.Password, existing.Password) {
		return entities.UserDTO{}, entities.ErrUnauthorized
	}

	return existing.ToUserDTO(), nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	fmt.Print(len(bytes))
	return string(bytes), err
}

// VerifyPassword verifies if the given password matches the stored hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
