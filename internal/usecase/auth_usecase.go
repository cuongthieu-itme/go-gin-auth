package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/me/go-gin-auth/internal/domain"
	"github.com/me/go-gin-auth/internal/repository"
	"github.com/me/go-gin-auth/pkg/jwt"
	"github.com/me/go-gin-auth/pkg/password"
)

// authUsecase - implement AuthUsecase interface
type authUsecase struct {
	userRepo          repository.UserRepository
	tokenRepo         repository.TokenRepository
	passwordResetRepo repository.PasswordResetRepository
	jwtService        jwt.Service
	passwordService   password.Service
	accessTokenTTL    time.Duration
	refreshTokenTTL   time.Duration
}

// NewAuthUsecase - tạo auth usecase mới
func NewAuthUsecase(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	passwordResetRepo repository.PasswordResetRepository,
	jwtService jwt.Service,
	passwordService password.Service,
	accessTokenTTL, refreshTokenTTL time.Duration,
) AuthUsecase {
	return &authUsecase{
		userRepo:          userRepo,
		tokenRepo:         tokenRepo,
		passwordResetRepo: passwordResetRepo,
		jwtService:        jwtService,
		passwordService:   passwordService,
		accessTokenTTL:    accessTokenTTL,
		refreshTokenTTL:   refreshTokenTTL,
	}
}

// Register - đăng ký user mới
func (u *authUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error) {
	// 1. Kiểm tra email đã tồn tại chưa
	existingUser, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// 2. Hash password
	hashedPassword, err := u.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. Tạo user mới
	user := &domain.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         "user",   // Mặc định là user
		Status:       "active", // Mặc định active
	}

	// 4. Lưu vào database
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 5. Trả về user response (không có password)
	return user.ToResponse(), nil
}

// Login - đăng nhập
func (u *authUsecase) Login(ctx context.Context, req *domain.LoginRequest) (string, string, *domain.UserResponse, error) {
	// 1. Tìm user theo email
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", "", nil, errors.New("invalid credentials") // Không nói cụ thể để tránh enumerate attack
	}

	// 2. Kiểm tra password
	if !u.passwordService.CheckPassword(req.Password, user.PasswordHash) {
		return "", "", nil, errors.New("invalid credentials")
	}

	// 3. Kiểm tra user status
	if user.Status != "active" {
		return "", "", nil, errors.New("user account is not active")
	}

	// 4. Tạo access token
	accessToken, err := u.jwtService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 5. Tạo refresh token
	refreshToken, err := u.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 6. Lưu refresh token vào database
	refreshTokenEntity := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(u.refreshTokenTTL),
	}

	if err := u.tokenRepo.CreateRefreshToken(ctx, refreshTokenEntity); err != nil {
		return "", "", nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return accessToken, refreshToken, user.ToResponse(), nil
}

// Logout - đăng xuất (vô hiệu hóa refresh token)
func (u *authUsecase) Logout(ctx context.Context, userID uint, refreshToken string) error {
	// Vô hiệu hóa refresh token
	if err := u.tokenRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// RefreshToken - làm mới token (token rotation)
func (u *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// 1. Validate refresh token format
	claims, err := u.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// 2. Kiểm tra token có trong database không
	tokenEntity, err := u.tokenRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to get refresh token: %w", err)
	}
	if tokenEntity == nil {
		return "", "", errors.New("refresh token not found")
	}

	// 3. Kiểm tra token đã bị revoke chưa
	if tokenEntity.Revoked {
		return "", "", errors.New("refresh token is revoked")
	}

	// 4. Kiểm tra token đã hết hạn chưa
	if tokenEntity.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("refresh token is expired")
	}

	// 5. Lấy user info
	user, err := u.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", "", errors.New("user not found")
	}

	// 6. Vô hiệu hóa token cũ (Token Rotation Pattern)
	if err := u.tokenRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// 7. Tạo token mới
	newAccessToken, err := u.jwtService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshToken, err := u.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// 8. Lưu refresh token mới
	newTokenEntity := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(u.refreshTokenTTL),
	}

	if err := u.tokenRepo.CreateRefreshToken(ctx, newTokenEntity); err != nil {
		return "", "", fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

// ForgotPassword - quên password (gửi reset token)
func (u *authUsecase) ForgotPassword(ctx context.Context, email string) error {
	// 1. Kiểm tra user có tồn tại không
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		// Không nói email không tồn tại để tránh enumerate attack
		return nil
	}

	// 2. Tạo reset token
	resetToken := uuid.New().String()

	// 3. Tạo password reset record
	passwordReset := &domain.PasswordReset{
		Email:     email,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Hết hạn sau 1 giờ
	}

	if err := u.passwordResetRepo.Create(ctx, passwordReset); err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	// 4. TODO: Gửi email chứa reset token
	// Trong thực tế, bạn sẽ gửi email có link: https://yourapp.com/reset-password?token=resetToken
	fmt.Printf("Password reset token for %s: %s\n", email, resetToken)

	return nil
}

// ResetPassword - đặt lại password bằng reset token
func (u *authUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	// 1. Lấy password reset record
	passwordReset, err := u.passwordResetRepo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get password reset: %w", err)
	}
	if passwordReset == nil {
		return errors.New("invalid reset token")
	}

	// 2. Kiểm tra token đã hết hạn chưa
	if passwordReset.ExpiresAt.Before(time.Now()) {
		return errors.New("reset token is expired")
	}

	// 3. Kiểm tra token đã được dùng chưa
	if passwordReset.Used {
		return errors.New("reset token is already used")
	}

	// 4. Lấy user
	user, err := u.userRepo.GetByEmail(ctx, passwordReset.Email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 5. Hash password mới
	hashedPassword, err := u.passwordService.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 6. Update password
	user.PasswordHash = hashedPassword
	if err := u.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	// 7. Đánh dấu token đã được sử dụng
	if err := u.passwordResetRepo.MarkAsUsed(ctx, token); err != nil {
		return fmt.Errorf("failed to mark reset token as used: %w", err)
	}

	// 8. Vô hiệu hóa tất cả refresh token của user (force re-login)
	if err := u.tokenRepo.RevokeAllForUser(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	return nil
}
