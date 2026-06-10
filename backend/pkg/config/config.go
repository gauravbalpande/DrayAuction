package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port         string
	Environment  string
	AllowOrigins []string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			AllowOrigins: []string{getEnv("CORS_ORIGIN", "http://localhost:3000")},
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://auctionxi:auctionxi@localhost:5432/auctionxi?sslmode=disable"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "dev-access-secret-change-in-production"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret-change-in-production"),
			AccessExpiry:  parseDuration(getEnv("JWT_ACCESS_EXPIRY", "24h")),
			RefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func parseInt(s string, fallback int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
