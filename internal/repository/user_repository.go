package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/me/go-gin-auth/internal/domain"
	"gorm.io/gorm"
)

// userRepository - struct implement UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository - tạo user repository mới
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create - tạo user mới
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// GORM sẽ tự động set ID, CreatedAt, UpdatedAt
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByEmail - lấy user theo email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Không tìm thấy -> trả về nil, không phải error
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetByID - lấy user theo ID
func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User

	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// Update - cập nhật user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	// GORM sẽ tự động update UpdatedAt
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// List - lấy danh sách user với filter và pagination
func (r *userRepository) List(ctx context.Context, limit, offset int, search, role, status string) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	// Tạo query builder
	query := r.db.WithContext(ctx).Model(&domain.User{})

	// Apply filters (nếu có)
	if search != "" {
		// Tìm kiếm trong full_name hoặc email
		query = query.Where("full_name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Đếm tổng số record (trước khi áp dụng limit/offset)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Lấy danh sách với pagination
	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}
