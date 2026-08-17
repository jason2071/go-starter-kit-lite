package platform

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
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
		AppName: env("APP_NAME", "go-starter-kit-lite"), AppEnv: env("APP_ENV", "development"), AppPort: env("APP_PORT", "8080"), LogLevel: env("LOG_LEVEL", "info"),
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"), MaxOpenConns: intEnv("DB_MAX_OPEN_CONNS", 25), MaxIdleConns: intEnv("DB_MAX_IDLE_CONNS", 10),
		JWTSecret: secret, JWTIssuer: env("JWT_ISSUER", "go-starter-kit-lite"), JWTAudience: env("JWT_AUDIENCE", "go-starter-kit-lite-api"), AccessTTL: accessTTL, RefreshTTL: refreshTTL,
		AllowedOrigin: env("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
	}, nil
}

func NewLogger(level string) *slog.Logger {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}

func NewDatabase(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{TranslateError: true, Logger: gormlogger.Default.LogMode(gormlogger.Warn)})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

type Security struct {
	secret           []byte
	issuer, audience string
	accessTTL        time.Duration
}

type claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func NewSecurity(secret, issuer, audience string, accessTTL time.Duration) *Security {
	return &Security{[]byte(secret), issuer, audience, accessTTL}
}

func (s *Security) GenerateAccessToken(userID string, roles []string) (string, error) {
	now := time.Now().UTC()
	c := claims{Roles: roles, RegisteredClaims: jwt.RegisteredClaims{Subject: userID, Issuer: s.issuer, Audience: jwt.ClaimStrings{s.audience}, IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL))}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.secret)
}
func (s *Security) ParseAccessToken(raw string) (*usecase.AccessClaims, error) {
	c := &claims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(s.issuer), jwt.WithAudience(s.audience))
	token, err := parser.ParseWithClaims(raw, c, func(*jwt.Token) (any, error) { return s.secret, nil })
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid access token")
	}
	return &usecase.AccessClaims{Subject: c.Subject, Roles: c.Roles}, nil
}
func (s *Security) AccessTokenTTL() time.Duration { return s.accessTTL }
func (s *Security) GenerateRefreshToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)
	return raw, s.HashRefreshToken(raw), nil
}
func (s *Security) HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
func (s *Security) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
func (s *Security) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
func intEnv(key string, fallback int) int {
	v := env(key, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	v := env(key, "")
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return d, nil
}
