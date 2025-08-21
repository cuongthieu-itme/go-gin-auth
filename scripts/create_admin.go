package main

import (
	"context"
	"log"
	"os"

	"github.com/me/go-gin-auth/internal/config"
	"github.com/me/go-gin-auth/internal/domain"
	"github.com/me/go-gin-auth/internal/repository"
	"github.com/me/go-gin-auth/internal/storage"
	"github.com/me/go-gin-auth/pkg/password"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("Creating admin user...")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config load failed:", err)
	}

	// Connect database
	db, err := storage.NewDatabase(&cfg.Database, logger.Error)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Create services
	passwordService := password.NewPasswordService(cfg.Security.BcryptCost)
	userRepo := repository.NewUserRepository(db)

	// Get admin credentials from env or use defaults
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@example.com"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
	}

	adminName := os.Getenv("ADMIN_NAME")
	if adminName == "" {
		adminName = "Admin User"
	}

	// Check if admin already exists
	ctx := context.Background()
	existingAdmin, err := userRepo.GetByEmail(ctx, adminEmail)
	if err != nil {
		log.Fatal("Check existing admin failed:", err)
	}

	if existingAdmin != nil {
		log.Printf("Admin user already exists: %s", adminEmail)
		return
	}

	// Hash admin password
	hashedPassword, err := passwordService.HashPassword(adminPassword)
	if err != nil {
		log.Fatal("Password hash failed:", err)
	}

	// Create admin user
	admin := &domain.User{
		Email:        adminEmail,
		PasswordHash: hashedPassword,
		FullName:     adminName,
		Role:         "admin",
		Status:       "active",
	}

	if err := userRepo.Create(ctx, admin); err != nil {
		log.Fatal("Create admin failed:", err)
	}

	log.Printf("✅ Admin user created successfully!")
	log.Printf("Email: %s", adminEmail)
	log.Printf("Password: %s", adminPassword)
	log.Printf("⚠️  Please change the password after first login!")
}
