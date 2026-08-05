package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"

	authv1 "backendGo/api/auth/v1"
	auth "backendGo/internal/auth"
	entities "backendGo/internal/domain"
	store "backendGo/internal/store"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	authv1.UnimplementedAuthServiceServer
	repo store.AuthRepository
	rds  *store.Redis
}

func NewAuthService(repo store.AuthRepository, rds *store.Redis) *AuthService {
	return &AuthService{repo: repo, rds: rds}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	// Check if username already exists
	_, err := s.repo.GetUserByUsername(req.Username)
	if err == nil {
		return &authv1.RegisterResponse{
			Success: false,
			Error:   "registration failed",
		}, nil
	}

	if err := auth.ValidateRegisterRequest(req.Username, req.Password); err != nil {
		return &authv1.RegisterResponse{
			Success: false,
			Error:   "invalid request",
		}, nil
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return &authv1.RegisterResponse{
			Success: false,
			Error:   "internal server error",
		}, nil
	}

	// Create user
	user := entities.User{
		Username: req.Username,
		Password: hashedPassword,
	}

	createdUser, err := s.repo.CreateUser(user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return &authv1.RegisterResponse{
			Success: false,
			Error:   "failed to create user",
		}, nil
	}

	return &authv1.RegisterResponse{
		Success: true,
		UserId:  int32(createdUser.ID),
	}, nil
}

// Login authenticates user and issues JWT tokens
func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	user, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	if !auth.VerifyPassword(req.Password, user.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token, err := auth.IssueTokens(strconv.Itoa(int(user.ID)))
	if err != nil {
		log.Printf("Error issuing tokens: %v", err)
		return nil, status.Error(codes.Internal, "failed to issue tokens")
	}

	if err := auth.Persist(ctx, s.rds, token); err != nil {
		log.Printf("Error persisting tokens: %v", err)
		return nil, status.Error(codes.Internal, "failed to persist tokens")
	}

	return &authv1.LoginResponse{
		AccessToken:  token.Access,
		RefreshToken: token.Refresh,
		UserId:       int32(user.ID),
		Message:      "login successful",
	}, nil
}

// ValidateToken checks if access token is valid and not revoked
func (s *AuthService) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, err := auth.ParseAccess(req.Token)
	if err != nil {
		return &authv1.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	_, err = s.rds.GetUserByJTI(ctx, "access:"+claims.ID)
	if err != nil {
		return &authv1.ValidateTokenResponse{
			Valid: false,
			Error: "token revoked",
		}, nil
	}

	userID, _ := strconv.Atoi(claims.Subject)
	return &authv1.ValidateTokenResponse{
		Valid:  true,
		UserId: int32(userID),
	}, nil
}

// RefreshToken issues new tokens using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	claims, err := auth.ParseRefresh(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	_, err = s.rds.GetUserByJTI(ctx, "refresh:"+claims.ID)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "refresh token revoked")
	}

	newTokens, err := auth.IssueTokens(claims.Subject)
	if err != nil {
		log.Printf("Error refreshing tokens: %v", err)
		return nil, status.Error(codes.Internal, "failed to refresh tokens")
	}

	if err := auth.Persist(ctx, s.rds, newTokens); err != nil {
		log.Printf("Error persisting refreshed tokens: %v", err)
		return nil, status.Error(codes.Internal, "failed to persist tokens")
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:  newTokens.Access,
		RefreshToken: newTokens.Refresh,
	}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest) (*authv1.ChangePasswordResponse, error) {
	// Get current user to verify password
	user, err := s.repo.GetUserByID(int(req.UserId)) // Need this method or getByID
	if err != nil {
		return &authv1.ChangePasswordResponse{
			Success: false,
			Error:   "user not found",
		}, nil
	}

	// Verify current password
	if !auth.VerifyPassword(req.CurrentPassword, user.Password) {
		return &authv1.ChangePasswordResponse{
			Success: false,
			Error:   "current password is incorrect",
		}, nil
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing new password: %v", err)
		return &authv1.ChangePasswordResponse{
			Success: false,
			Error:   "internal server error",
		}, nil
	}

	// Update password in database
	err = s.repo.UpdatePassword(int(req.UserId), hashedPassword)
	if err != nil {
		log.Printf("Error updating password: %v", err)
		return &authv1.ChangePasswordResponse{
			Success: false,
			Error:   "failed to update password",
		}, nil
	}

	// Revoke all active sessions (force re-login)
	s.revokeAllUserTokens(ctx, req.UserId)

	return &authv1.ChangePasswordResponse{
		Success: true,
		Error:   "",
	}, nil
}

// revokeAllUserTokens removes all tokens for a user (force logout on password change)
func (s *AuthService) revokeAllUserTokens(ctx context.Context, userID int32) {
	// This would ideally scan Redis for all "access:user:<userID>" keys and delete them
	// For now, we just log it happens
	log.Printf("Revoking all tokens for user %d", userID)
}

// Logout revokes access token
func (s *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	claims, err := auth.ParseAccess(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	s.rds.DelJTI(ctx, "access:"+claims.ID)

	return &authv1.LogoutResponse{
		Message: "logged out successfully",
	}, nil
}

type Config struct {
	DBDSN    string
	RedisDSN string
	Port     string
}

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println(err.Error())
	}

	db, err := store.Connect(store.GetConnectionURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	rds := store.NewRedis()
	defer rds.Client.Close()

	repo := store.NewSQLiteAuthRepository(db)

	lis, err := net.Listen("tcp", ":"+os.Getenv("AUTH_PORT"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	authv1.RegisterAuthServiceServer(grpcServer, NewAuthService(repo, rds))

	log.Printf("Auth gRPC server listening on port %s", os.Getenv("AUTH_PORT"))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
