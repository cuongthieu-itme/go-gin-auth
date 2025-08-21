package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/me/go-gin-auth/internal/domain"
	"github.com/me/go-gin-auth/internal/usecase"
	"github.com/me/go-gin-auth/pkg/response"
	"github.com/me/go-gin-auth/pkg/validator"
)

// AuthHandler - xử lý các API về authentication
type AuthHandler struct {
	authUsecase usecase.AuthUsecase
	validator   *validator.Validator
}

// NewAuthHandler - tạo auth handler mới
func NewAuthHandler(authUsecase usecase.AuthUsecase, validator *validator.Validator) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		validator:   validator,
	}
}

// Register - API đăng ký user mới
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	// 1. Parse JSON request
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 2. Validate input
	if errs := h.validator.Validate(&req); len(errs) > 0 {
		response.ValidationError(c, "Validation failed", errs)
		return
	}

	// 3. Call usecase
	user, err := h.authUsecase.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Registration failed", err)
		return
	}

	// 4. Return success response
	response.Success(c, http.StatusCreated, "User registered successfully", user)
}

// Login - API đăng nhập
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	// 1. Parse request
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 2. Validate
	if errs := h.validator.Validate(&req); len(errs) > 0 {
		response.ValidationError(c, "Validation failed", errs)
		return
	}

	// 3. Call usecase
	accessToken, refreshToken, user, err := h.authUsecase.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Login failed", err)
		return
	}

	// 4. Return tokens and user info
	data := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	}

	response.Success(c, http.StatusOK, "Login successful", data)
}

// Logout - API đăng xuất
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// 1. Parse request
	var req domain.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 2. Get user ID từ middleware (đã được set bởi AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// 3. Call usecase
	err := h.authUsecase.Logout(c.Request.Context(), userID.(uint), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Logout failed", err)
		return
	}

	response.Success(c, http.StatusOK, "Logout successful", nil)
}

// RefreshToken - API làm mới token
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 1. Parse request
	var req domain.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 2. Validate
	if errs := h.validator.Validate(&req); len(errs) > 0 {
		response.ValidationError(c, "Validation failed", errs)
		return
	}

	// 3. Call usecase
	accessToken, refreshToken, err := h.authUsecase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Token refresh failed", err)
		return
	}

	// 4. Return new tokens
	data := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	response.Success(c, http.StatusOK, "Token refreshed successfully", data)
}

// ForgotPassword - API quên password
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	// 1. Parse request
	var req domain.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 2. Validate
	if errs := h.validator.Validate(&req); len(errs) > 0 {
		response.ValidationError(c, "Validation failed", errs)
		return
	}

	// 3. Call usecase
	err := h.authUsecase.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to process forgot password request", err)
		return
	}

	response.Success(c, http.StatusOK, "Password reset instructions sent to your email", nil)
}

// ResetPassword - API reset password
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	// 1. Parse request
	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 2. Validate
	if errs := h.validator.Validate(&req); len(errs) > 0 {
		response.ValidationError(c, "Validation failed", errs)
		return
	}

	// 3. Call usecase
	err := h.authUsecase.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Password reset failed", err)
		return
	}

	response.Success(c, http.StatusOK, "Password reset successfully", nil)
}
