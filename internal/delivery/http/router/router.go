package router

import (
	"github.com/gin-gonic/gin"
	"github.com/me/go-gin-auth/internal/config"
	"github.com/me/go-gin-auth/internal/delivery/http/handler"
	"github.com/me/go-gin-auth/internal/delivery/http/middleware"
	"github.com/me/go-gin-auth/pkg/jwt"
	"go.uber.org/zap"
)

// RouterConfig - config để setup router
type RouterConfig struct {
	AuthHandler   *handler.AuthHandler
	UserHandler   *handler.UserHandler
	HealthHandler *handler.HealthHandler
	JWTService    jwt.Service
	Logger        *zap.Logger
	Config        *config.Config
}

// NewRouter - tạo Gin router với tất cả routes
func NewRouter(cfg *RouterConfig) *gin.Engine {
	// 1. Set Gin mode dựa vào environment
	if cfg.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. Tạo Gin engine
	r := gin.New()

	// 3. Global middleware (áp dụng cho tất cả routes)
	r.Use(gin.Recovery())                                            // Recover từ panic
	r.Use(middleware.LoggerMiddleware(cfg.Logger))                   // Log requests
	r.Use(middleware.CORSMiddleware(cfg.Config.CORS.AllowedOrigins)) // CORS

	// 4. Health check endpoint (không cần auth)
	r.GET("/health", cfg.HealthHandler.HealthCheck)

	// 5. API routes group
	api := r.Group("/api/v1")
	{
		// Auth routes (không cần authentication)
		auth := api.Group("/auth")
		{
			auth.POST("/register", cfg.AuthHandler.Register)
			auth.POST("/login", cfg.AuthHandler.Login)
			auth.POST("/refresh", cfg.AuthHandler.RefreshToken)
			auth.POST("/forgot-password", cfg.AuthHandler.ForgotPassword)
			auth.POST("/reset-password", cfg.AuthHandler.ResetPassword)

			// Logout cần auth để lấy user_id
			auth.POST("/logout", middleware.AuthMiddleware(cfg.JWTService), cfg.AuthHandler.Logout)
		}

		// Protected routes (cần authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTService))
		{
			// Health check với auth
			protected.GET("/health", cfg.HealthHandler.DatabaseHealthCheck)

			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", cfg.UserHandler.GetProfile)
				users.PUT("/me", cfg.UserHandler.UpdateProfile)
				users.POST("/change-password", cfg.UserHandler.ChangePassword)

				// Admin only routes
				users.GET("", middleware.RequireRoles("admin"), cfg.UserHandler.ListUsers)
			}
		}
	}

	return r
}
