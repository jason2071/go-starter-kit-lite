package platform

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName       string
	AppEnv        string
	AppPort       string
	LogLevel      string
	DatabaseURL   string
	MaxOpenConns  int
	MaxIdleConns  int
	JWTSecret     string
	JWTIssuer     string
	JWTAudience   string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	AllowedOrigin string
}

func LoadConfig() (Config, error) {
	accessTTL, err := durationEnv("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}

	refreshTTL, err := durationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	secret := env("JWT_SECRET", "")
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}

	return Config{
		AppName:       env("APP_NAME", "go-starter-kit-lite"),
		AppEnv:        env("APP_ENV", "development"),
		AppPort:       env("APP_PORT", "8080"),
		LogLevel:      env("LOG_LEVEL", "info"),
		DatabaseURL:   env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"),
		MaxOpenConns:  intEnv("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:  intEnv("DB_MAX_IDLE_CONNS", 10),
		JWTSecret:     secret,
		JWTIssuer:     env("JWT_ISSUER", "go-starter-kit-lite"),
		JWTAudience:   env("JWT_AUDIENCE", "go-starter-kit-lite-api"),
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
		AllowedOrigin: env("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
	}, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func intEnv(key string, fallback int) int {
	value := env(key, "")
	if value == "" {
		return fallback
	}

	number, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return number
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := env(key, "")
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return duration, nil
}
