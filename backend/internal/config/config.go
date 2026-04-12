package config

import (
	"fmt"
	"strconv"

	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/utils"
)

type Config struct {
	AppEnv         string
	AppPort        string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	JWTSecret      string
	JWTExpiryHours int
}

func LoadConfig() (*Config, error) {
	jwtExpiryHours, err := strconv.Atoi(utils.GetEnv("JWT_EXPIRATION_HOURS", utils.GetEnv("JWT_EXPIRY_HOURS", "24")))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION_HOURS: %w", err)
	}

	cfg := &Config{
		AppEnv:         utils.GetEnv("APP_ENV", "development"),
		AppPort:        utils.GetEnv("APP_PORT", "8080"),
		DBHost:         utils.GetEnv("DB_HOST", "localhost"),
		DBPort:         utils.GetEnv("DB_PORT", "5432"),
		DBUser:         utils.GetEnv("DB_USER", "postgres"),
		DBPassword:     utils.GetEnv("DB_PASSWORD", "password"),
		DBName:         utils.GetEnv("DB_NAME", "myapp"),
		DBSSLMode:      utils.GetEnv("DB_SSLMODE", utils.GetEnv("DB_SSL_MODE", "disable")),
		JWTSecret:      utils.GetEnv("JWT_SECRET", ""),
		JWTExpiryHours: jwtExpiryHours,
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}


