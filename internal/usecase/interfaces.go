package usecase

import (
	"context"

	"github.com/me/go-gin-auth/internal/domain"
)

// AuthUsecase - interface cho authentication logic
type AuthUsecase interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (string, string, *domain.UserResponse, error) // accessToken, refreshToken, user, error
	Logout(ctx context.Context, userID uint, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error) // newAccessToken, newRefreshToken, error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// UserUsecase - interface cho user management logic
type UserUsecase interface {
	GetProfile(ctx context.Context, userID uint) (*domain.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uint, req *domain.UpdateProfileRequest) (*domain.UserResponse, error)
	ChangePassword(ctx context.Context, userID uint, req *domain.ChangePasswordRequest) error
	ListUsers(ctx context.Context, req *domain.ListUsersRequest) (*domain.PaginatedUsersResponse, error)
}
