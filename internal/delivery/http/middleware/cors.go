package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware - middleware xử lý CORS
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	config := cors.DefaultConfig()

	// Cài đặt allowed origins
	config.AllowOrigins = allowedOrigins

	// Cài đặt allowed methods
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	// Cài đặt allowed headers
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}

	// Cho phép cookies
	config.AllowCredentials = true

	return cors.New(config)
}
