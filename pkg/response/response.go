package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response - format chuẩn cho API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// Success - trả về response thành công
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error - trả về response lỗi
func Error(c *gin.Context, statusCode int, message string, err error) {
	response := Response{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	c.JSON(statusCode, response)
}

// ValidationError - trả về lỗi validation
func ValidationError(c *gin.Context, message string, errors interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: message,
		Data:    errors,
		Code:    "VALIDATION_ERROR",
	})
}

// Unauthorized - trả về lỗi không có quyền
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Message: message,
		Code:    "UNAUTHORIZED",
	})
}

// Forbidden - trả về lỗi cấm truy cập
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Message: message,
		Code:    "FORBIDDEN",
	})
}
