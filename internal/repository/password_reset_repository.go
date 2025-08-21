package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/me/go-gin-auth/internal/domain"
	"gorm.io/gorm"
)

// passwordResetRepository - implement PasswordResetRepository interface
type passwordResetRepository struct {
	db *gorm.DB
}

// NewPasswordResetRepository - tạo password reset repository mới
func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

// Create - tạo password reset request
func (r *passwordResetRepository) Create(ctx context.Context, reset *domain.PasswordReset) error {
	if err := r.db.WithContext(ctx).Create(reset).Error; err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}
	return nil
}

// GetByToken - lấy password reset theo token
func (r *passwordResetRepository) GetByToken(ctx context.Context, token string) (*domain.PasswordReset, error) {
	var reset domain.PasswordReset

	err := r.db.WithContext(ctx).Where("token = ?", token).First(&reset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get password reset by token: %w", err)
	}

	return &reset, nil
}

// MarkAsUsed - đánh dấu token đã được sử dụng
func (r *passwordResetRepository) MarkAsUsed(ctx context.Context, token string) error {
	err := r.db.WithContext(ctx).Model(&domain.PasswordReset{}).
		Where("token = ?", token).
		Update("used", true).Error

	if err != nil {
		return fmt.Errorf("failed to mark password reset as used: %w", err)
	}
	return nil
}

// CleanupExpired - xóa các request hết hạn hoặc đã dùng
func (r *passwordResetRepository) CleanupExpired(ctx context.Context) error {
	err := r.db.WithContext(ctx).
		Where("expires_at < ? OR used = ?", time.Now(), true).
		Delete(&domain.PasswordReset{}).Error

	if err != nil {
		return fmt.Errorf("failed to cleanup expired password resets: %w", err)
	}
	return nil
}
