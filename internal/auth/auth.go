package auth

import (
	"context"
	"errors"
	"time"

	authv1 "backendGo/api/auth/v1"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client authv1.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	return &Client{
		client: authv1.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Register(ctx context.Context, username, password string) (*authv1.RegisterResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Register(ctx, &authv1.RegisterRequest{
		Username: username,
		Password: password,
	})
}

func (c *Client) ChangePassword(ctx context.Context, userID int, currentPassword, newPassword string) (*authv1.ChangePasswordResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.ChangePassword(ctx, &authv1.ChangePasswordRequest{
		UserId:          int32(userID),
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
	})
}

func (c *Client) Login(ctx context.Context, username, password string) (*authv1.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Login(ctx, &authv1.LoginRequest{
		Username: username,
		Password: password,
	})
}

func (c *Client) ValidateToken(ctx context.Context, token string) (*authv1.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.ValidateToken(ctx, &authv1.ValidateTokenRequest{
		Token: token,
	})
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*authv1.RefreshTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.RefreshToken(ctx, &authv1.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
}

func (c *Client) Logout(ctx context.Context, token string) (*authv1.LogoutResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Logout(ctx, &authv1.LogoutRequest{
		Token: token,
	})
}

// Helper functions for password handling
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateRegisterRequest(username, password string) error {
	if len(username) < 3 || len(username) > 50 {
		return errors.New("invalid username")
	}
	if len(password) < 6 {
		return errors.New("password too short")
	}
	return nil
}
