package storage

import (
	"fmt"
	"time"

	"github.com/me/go-gin-auth/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase - tạo kết nối database mới
func NewDatabase(cfg *config.DatabaseConfig, logLevel logger.LogLevel) (*gorm.DB, error) {
	// Tạo DSN (Data Source Name) cho MySQL
	// Format: user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,     // root
		cfg.Password, // root
		cfg.Host,     // localhost
		cfg.Port,     // 3306
		cfg.Name,     // authdb
	)

	// Mở kết nối với GORM
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel), // Set log level cho GORM
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Lấy underlying sql.DB để config connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Cài đặt connection pool để tối ưu hiệu suất
	sqlDB.SetMaxOpenConns(25)                 // Tối đa 25 kết nối đồng thời
	sqlDB.SetMaxIdleConns(5)                  // Giữ 5 kết nối nhàn rỗi
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Mỗi kết nối sống tối đa 5 phút

	// Test kết nối
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
