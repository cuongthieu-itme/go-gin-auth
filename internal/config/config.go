package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config chứa tất cả cài đặt của app
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Security  SecurityConfig  `mapstructure:"security"`
}

// AppConfig - cài đặt app chung
type AppConfig struct {
	Port string `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

// DatabaseConfig - cài đặt database
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// JWTConfig - cài đặt JWT
type JWTConfig struct {
	AccessSecret  string        `mapstructure:"access_secret"`
	RefreshSecret string        `mapstructure:"refresh_secret"`
	AccessTTL     time.Duration `mapstructure:"access_ttl"`
	RefreshTTL    time.Duration `mapstructure:"refresh_ttl"`
}

// LogConfig - cài đặt logging
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// CORSConfig - cài đặt CORS
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// RateLimitConfig - cài đặt rate limiting
type RateLimitConfig struct {
	Requests int           `mapstructure:"requests"`
	Window   time.Duration `mapstructure:"window"`
}

// SecurityConfig - cài đặt bảo mật
type SecurityConfig struct {
	BcryptCost int `mapstructure:"bcrypt_cost"`
}

// Load - đọc config từ file .env
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv() // Đọc từ environment variables

	// Giá trị mặc định (nếu không có trong .env)
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_WINDOW", "1m")
	viper.SetDefault("BCRYPT_COST", 12)

	// Đọc file .env (optional - nếu không có hoặc không đọc được thì skip)
	if err := viper.ReadInConfig(); err != nil {
		// Chỉ log warning, không fail app nếu không đọc được .env
		// Vì trong Docker, ta sử dụng environment variables
		log.Printf("Could not read .env file (this is fine in Docker): %v", err)
	}

	// Tạo struct config từ các giá trị đã đọc
	config := &Config{
		App: AppConfig{
			Port: viper.GetString("APP_PORT"),
			Env:  viper.GetString("APP_ENV"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASS"),
			Name:     viper.GetString("DB_NAME"),
		},
		JWT: JWTConfig{
			AccessSecret:  viper.GetString("JWT_ACCESS_SECRET"),
			RefreshSecret: viper.GetString("JWT_REFRESH_SECRET"),
			AccessTTL:     viper.GetDuration("JWT_ACCESS_TTL"),
			RefreshTTL:    viper.GetDuration("JWT_REFRESH_TTL"),
		},
		Log: LogConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
		CORS: CORSConfig{
			AllowedOrigins: viper.GetStringSlice("CORS_ALLOWED_ORIGINS"),
		},
		RateLimit: RateLimitConfig{
			Requests: viper.GetInt("RATE_LIMIT_REQUESTS"),
			Window:   viper.GetDuration("RATE_LIMIT_WINDOW"),
		},
		Security: SecurityConfig{
			BcryptCost: viper.GetInt("BCRYPT_COST"),
		},
	}

	return config, nil
}
