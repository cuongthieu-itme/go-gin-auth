package domain

import (
	"time"
)

// User - struct đại diện cho user trong database
type User struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	Email           string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash    string     `json:"-" gorm:"not null"` // Dấu "-" nghĩa là không trả về trong JSON
	FullName        string     `json:"full_name" gorm:"not null"`
	Role            string     `json:"role" gorm:"type:enum('admin','user');default:'user'"`
	Status          string     `json:"status" gorm:"type:enum('active','inactive','suspended');default:'active'"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// DTO (Data Transfer Object) - dùng cho API request/response

// RegisterRequest - dữ liệu khi user đăng ký
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
}

// LoginRequest - dữ liệu khi user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserResponse - dữ liệu trả về (không có password)
type UserResponse struct {
	ID              uint       `json:"id"`
	Email           string     `json:"email"`
	FullName        string     `json:"full_name"`
	Role            string     `json:"role"`
	Status          string     `json:"status"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// UpdateProfileRequest - dữ liệu update profile
type UpdateProfileRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
}

// ChangePasswordRequest - dữ liệu đổi password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// ListUsersRequest - dữ liệu lọc danh sách user
type ListUsersRequest struct {
	Page   int    `form:"page" validate:"min=1"`
	Limit  int    `form:"limit" validate:"min=1,max=100"`
	Search string `form:"search"`
	Role   string `form:"role"`
	Status string `form:"status"`
}

// PaginatedUsersResponse - kết quả danh sách user có phân trang
type PaginatedUsersResponse struct {
	Users      []UserResponse `json:"users"`
	Pagination Pagination     `json:"pagination"`
}

// Pagination - thông tin phân trang
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ToResponse - chuyển từ User entity sang UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		FullName:        u.FullName,
		Role:            u.Role,
		Status:          u.Status,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
	}
}
