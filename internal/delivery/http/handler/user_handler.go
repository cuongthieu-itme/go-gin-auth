package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/me/go-gin-auth/internal/domain"
	"github.com/me/go-gin-auth/internal/usecase"
	"github.com/me/go-gin-auth/pkg/response"
	"github.com/me/go-gin-auth/pkg/validator"
)

// UserHandler - xử lý các API về user management
type UserHandler struct {
	userUsecase usecase.UserUsecase
	validator   *validator.Validator
}

// NewUserHandler - tạo user handler mới
func NewUserHandler(userUsecase usecase.UserUsecase, validator *validator.Validator) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
		validator:   validator,
	}
}

// GetProfile - API lấy thông tin profile của user hiện tại
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 1. Lấy user ID từ JWT token (đã được set bởi AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// 2. Call usecase
	user, err := h.userUsecase.GetProfile(c.Request.Context(), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user profile", err)
		return
	}

	response.Success(c, http.StatusOK, "User profile retrieved successfully", user)
}

// UpdateProfile - API cập nhật profile
// PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// 1. Get user ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// 2. Parse request
	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 3. Validate
	if errs := h.validator.Validate(&req); len(errs) > 0 {
		response.ValidationError(c, "Validation failed", errs)
		return
	}

	// 4. Call usecase
	user, err := h.userUsecase.UpdateProfile(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to update profile", err)
		return
	}

	response.Success(c, http.StatusOK, "Profile updated successfully", user)
}

// ChangePassword - API đổi password
// POST /api/v1/users/change-password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// 1. Get user ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// 2. Parse request
	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 3. Validate
	if errs := h.validator.Validate(&req); len(errs) > 0 {
		response.ValidationError(c, "Validation failed", errs)
		return
	}

	// 4. Call usecase
	err := h.userUsecase.ChangePassword(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to change password", err)
		return
	}

	response.Success(c, http.StatusOK, "Password changed successfully", nil)
}

// ListUsers - API lấy danh sách user (chỉ admin)
// GET /api/v1/users?page=1&limit=10&search=john
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 1. Parse query parameters
	var req domain.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// 2. Set defaults nếu không có
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// 3. Validate
	if errs := h.validator.Validate(&req); len(errs) > 0 {
		response.ValidationError(c, "Validation failed", errs)
		return
	}

	// 4. Call usecase
	users, err := h.userUsecase.ListUsers(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list users", err)
		return
	}

	response.Success(c, http.StatusOK, "Users retrieved successfully", users)
}
