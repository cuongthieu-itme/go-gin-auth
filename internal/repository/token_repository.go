package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/me/go-gin-auth/internal/domain"
	"gorm.io/gorm"
)

// tokenRepository - implement TokenRepository interface
type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository - tạo token repository mới
func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

// CreateRefreshToken - lưu refresh token vào database
func (r *tokenRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}
	return nil
}

// GetRefreshToken - lấy refresh token từ database
func (r *tokenRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken

	err := r.db.WithContext(ctx).Where("token = ?", token).First(&refreshToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &refreshToken, nil
}

// RevokeRefreshToken - vô hiệu hóa 1 refresh token
func (r *tokenRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	// Update field revoked = true cho token này
	err := r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked", true).Error

	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllForUser - vô hiệu hóa tất cả token của 1 user
func (r *tokenRepository) RevokeAllForUser(ctx context.Context, userID uint) error {
	err := r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error

	if err != nil {
		return fmt.Errorf("failed to revoke all tokens for user: %w", err)
	}
	return nil
}

// CleanupExpired - xóa các token đã hết hạn hoặc bị revoke
func (r *tokenRepository) CleanupExpired(ctx context.Context) error {
	// Xóa token hết hạn hoặc bị revoke
	err := r.db.WithContext(ctx).
		Where("expires_at < ? OR revoked = ?", time.Now(), true).
		Delete(&domain.RefreshToken{}).Error

	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return nil
}
