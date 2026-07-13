package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment string
	Port        int
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	JWTTTL      time.Duration
	LogLevel    string
	CORSOrigins []string
}

func Load() (Config, error) {
	port, err := strconv.Atoi(value("API_PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("API_PORT: %w", err)
	}
	ttl, err := time.ParseDuration(value("JWT_TTL", "24h"))
	if err != nil {
		return Config{}, fmt.Errorf("JWT_TTL: %w", err)
	}
	cfg := Config{
		Environment: value("APP_ENV", "development"),
		Port:        port,
		DatabaseURL: value("DATABASE_URL", "postgres://safescope:change-me@localhost:5432/safescope?sslmode=disable"),
		RedisURL:    value("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:   value("JWT_SECRET", "development-only-secret-change-me-now"),
		JWTTTL:      ttl,
		LogLevel:    value("LOG_LEVEL", "info"),
		CORSOrigins: split(value("CORS_ORIGINS", "http://localhost:3000")),
	}
	if cfg.Environment == "production" && len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 characters in production")
	}
	return cfg, nil
}

func value(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func split(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
