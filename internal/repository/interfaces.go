package repository

import (
	"context"

	"github.com/me/go-gin-auth/internal/domain"
)

// UserRepository - interface định nghĩa các thao tác với user
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error                                                    // Tạo user mới
	GetByEmail(ctx context.Context, email string) (*domain.User, error)                                     // Lấy user theo email
	GetByID(ctx context.Context, id uint) (*domain.User, error)                                             // Lấy user theo ID
	Update(ctx context.Context, user *domain.User) error                                                    // Cập nhật user
	List(ctx context.Context, limit, offset int, search, role, status string) ([]domain.User, int64, error) // Lấy danh sách user
}

// TokenRepository - interface cho refresh token
type TokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error        // Tạo refresh token
	GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) // Lấy refresh token
	RevokeRefreshToken(ctx context.Context, token string) error                      // Vô hiệu hóa token
	RevokeAllForUser(ctx context.Context, userID uint) error                         // Vô hiệu hóa tất cả token của user
	CleanupExpired(ctx context.Context) error                                        // Xóa token hết hạn
}

// PasswordResetRepository - interface cho reset password
type PasswordResetRepository interface {
	Create(ctx context.Context, reset *domain.PasswordReset) error               // Tạo reset request
	GetByToken(ctx context.Context, token string) (*domain.PasswordReset, error) // Lấy theo token
	MarkAsUsed(ctx context.Context, token string) error                          // Đánh dấu đã dùng
	CleanupExpired(ctx context.Context) error                                    // Xóa request hết hạn
}
