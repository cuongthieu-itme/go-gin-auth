package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/me/go-gin-auth/pkg/jwt"
	"github.com/me/go-gin-auth/pkg/response"
	"github.com/me/go-gin-auth/pkg/utils"
)

// AuthMiddleware - middleware xác thực JWT token
func AuthMiddleware(jwtService jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// 2. Kiểm tra format "Bearer <token>"
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		// 3. Extract token
		token := authHeader[len(bearerPrefix):]
		if token == "" {
			response.Unauthorized(c, "Token required")
			c.Abort()
			return
		}

		// 4. Validate token
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 5. Set user info vào context để handler khác dùng
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		// 6. Continue to next handler
		c.Next()
	}
}

// RequireRoles - middleware kiểm tra role user
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy role từ context (đã set bởi AuthMiddleware)
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Unauthorized(c, "User role not found")
			c.Abort()
			return
		}

		// 2. Kiểm tra role có trong danh sách cho phép không
		if !utils.Contains(roles, userRole.(string)) {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
