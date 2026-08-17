package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
	"github.com/jason2071/go-starter-kit-lite/internal/shared"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already exists")
	ErrRefreshNotFound  = errors.New("refresh token not found")
	ErrRefreshTokenUsed = errors.New("refresh token already used or expired")
)

type UserListOptions struct {
	Page     shared.PageParams
	Search   string
	IsActive *bool
	Sort     string
	Order    string
}

type UserRepository interface {
	CreateWithRole(ctx context.Context, user *domain.User, role string) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, opts UserListOptions) ([]domain.User, int64, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	FindActiveByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	Rotate(ctx context.Context, oldHash string, next *domain.RefreshToken) error
	Revoke(ctx context.Context, hash string) error
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
