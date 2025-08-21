package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/me/go-gin-auth/internal/config"
	"github.com/me/go-gin-auth/internal/delivery/http/handler"
	"github.com/me/go-gin-auth/internal/delivery/http/router"
	"github.com/me/go-gin-auth/internal/repository"
	"github.com/me/go-gin-auth/internal/storage"
	"github.com/me/go-gin-auth/internal/usecase"
	"github.com/me/go-gin-auth/pkg/jwt"
	"github.com/me/go-gin-auth/pkg/logger"
	"github.com/me/go-gin-auth/pkg/password"
	"github.com/me/go-gin-auth/pkg/validator"
	"go.uber.org/zap"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 2. Initialize logger
	appLogger, err := logger.NewLogger(cfg.Log.Level, cfg.App.Env)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer appLogger.Sync()

	appLogger.Info("Starting Go Gin Auth API...")

	// 3. Set GORM log level
	var gormLogLevel gormLogger.LogLevel
	switch cfg.Log.Level {
	case "debug":
		gormLogLevel = gormLogger.Info
	case "info":
		gormLogLevel = gormLogger.Warn
	default:
		gormLogLevel = gormLogger.Error
	}

	// 4. Initialize database
	db, err := storage.NewDatabase(&cfg.Database, gormLogLevel)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", zap.Error(err))
	}

	appLogger.Info("Database connected successfully")

	// 5. Initialize services
	jwtService := jwt.NewJWTService(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
	)
	passwordService := password.NewPasswordService(cfg.Security.BcryptCost)
	validatorService := validator.New()

	// 6. Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)

	// 7. Initialize usecases
	authUsecase := usecase.NewAuthUsecase(
		userRepo,
		tokenRepo,
		passwordResetRepo,
		jwtService,
		passwordService,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
	)
	userUsecase := usecase.NewUserUsecase(userRepo, passwordService)

	// 8. Initialize handlers
	authHandler := handler.NewAuthHandler(authUsecase, validatorService)
	userHandler := handler.NewUserHandler(userUsecase, validatorService)
	healthHandler := handler.NewHealthHandler(db)

	// 9. Initialize router
	r := router.NewRouter(&router.RouterConfig{
		AuthHandler:   authHandler,
		UserHandler:   userHandler,
		HealthHandler: healthHandler,
		JWTService:    jwtService,
		Logger:        appLogger,
		Config:        cfg,
	})

	// 10. Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}

	// 11. Start server trong goroutine
	go func() {
		appLogger.Info("Starting server", zap.String("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 12. Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// 13. Graceful shutdown với timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 14. Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	// 15. Close database connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	appLogger.Info("Server exited")
}
