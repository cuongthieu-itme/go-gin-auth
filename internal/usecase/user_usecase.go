package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/me/go-gin-auth/internal/domain"
	"github.com/me/go-gin-auth/internal/repository"
	"github.com/me/go-gin-auth/pkg/password"
)

// userUsecase - implement UserUsecase interface
type userUsecase struct {
	userRepo        repository.UserRepository
	passwordService password.Service
}

// NewUserUsecase - tạo user usecase mới
func NewUserUsecase(
	userRepo repository.UserRepository,
	passwordService password.Service,
) UserUsecase {
	return &userUsecase{
		userRepo:        userRepo,
		passwordService: passwordService,
	}
}

// GetProfile - lấy profile user
func (u *userUsecase) GetProfile(ctx context.Context, userID uint) (*domain.UserResponse, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user.ToResponse(), nil
}

// UpdateProfile - cập nhật profile
func (u *userUsecase) UpdateProfile(ctx context.Context, userID uint, req *domain.UpdateProfileRequest) (*domain.UserResponse, error) {
	// 1. Lấy user hiện tại
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 2. Update thông tin
	user.FullName = req.FullName

	// 3. Lưu vào database
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user.ToResponse(), nil
}

// ChangePassword - đổi password
func (u *userUsecase) ChangePassword(ctx context.Context, userID uint, req *domain.ChangePasswordRequest) error {
	// 1. Lấy user
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 2. Kiểm tra old password
	if !u.passwordService.CheckPassword(req.OldPassword, user.PasswordHash) {
		return errors.New("invalid old password")
	}

	// 3. Hash password mới
	hashedPassword, err := u.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Update password
	user.PasswordHash = hashedPassword

	if err := u.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// ListUsers - lấy danh sách user (admin only)
func (u *userUsecase) ListUsers(ctx context.Context, req *domain.ListUsersRequest) (*domain.PaginatedUsersResponse, error) {
	// 1. Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// 2. Tính offset cho pagination
	offset := (req.Page - 1) * req.Limit

	// 3. Lấy danh sách từ repository
	users, total, err := u.userRepo.List(ctx, req.Limit, offset, req.Search, req.Role, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 4. Convert sang response format
	userResponses := make([]domain.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *user.ToResponse()
	}

	// 5. Tính total pages
	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	// 6. Tạo response
	return &domain.PaginatedUsersResponse{
		Users: userResponses,
		Pagination: domain.Pagination{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
