package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/me/go-gin-auth/pkg/response"
	"gorm.io/gorm"
)

// HealthHandler - xử lý health check APIs
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler - tạo health handler mới
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheck - API health check cơ bản (không cần auth)
// GET /health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, "Server is healthy", gin.H{
		"status":  "ok",
		"service": "go-gin-auth",
	})
}

// DatabaseHealthCheck - API health check với database (cần auth)
// GET /api/v1/health
func (h *HealthHandler) DatabaseHealthCheck(c *gin.Context) {
	// Kiểm tra kết nối database
	sqlDB, err := h.db.DB()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Database connection error", err)
		return
	}

	// Ping database
	if err := sqlDB.Ping(); err != nil {
		response.Error(c, http.StatusInternalServerError, "Database ping failed", err)
		return
	}

	response.Success(c, http.StatusOK, "Database is healthy", gin.H{
		"status":   "ok",
		"database": "connected",
		"service":  "go-gin-auth",
	})
}
