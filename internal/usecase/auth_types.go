package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenUsed     = errors.New("refresh token already used or expired")
)

type RefreshTokenStore interface {
	CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	FindActiveRefreshTokenByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	RotateRefreshToken(ctx context.Context, oldHash string, next *domain.RefreshToken) error
	RevokeRefreshToken(ctx context.Context, hash string) error
}

type AccessClaims struct {
	Subject string
	Roles   []string
}

type TokenManager interface {
	GenerateAccessToken(userID string, roles []string) (string, error)
	ParseAccessToken(raw string) (*AccessClaims, error)
	AccessTokenTTL() time.Duration
	GenerateRefreshToken() (raw string, hash string, err error)
	HashRefreshToken(raw string) string
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"name" validate:"required,min=2,max=255"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,max=72"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type AuthResponse struct {
	User   UserResponse `json:"user"`
	Tokens TokenPair    `json:"tokens"`
}
