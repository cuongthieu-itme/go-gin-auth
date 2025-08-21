package password

import (
	"golang.org/x/crypto/bcrypt"
)

// Service - interface cho password operations
type Service interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hashedPassword string) bool
}

// passwordService - implementation
type passwordService struct {
	cost int // Độ mạnh mã hóa (12 = mạnh, 4 = yếu)
}

// NewPasswordService - tạo password service mới
func NewPasswordService(cost int) Service {
	return &passwordService{cost: cost}
}

// HashPassword - mã hóa password
func (s *passwordService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword - kiểm tra password có đúng không
func (s *passwordService) CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil // true nếu đúng, false nếu sai
}
