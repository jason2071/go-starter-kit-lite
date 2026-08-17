package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenUsed     = errors.New("refresh token already used or expired")
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// RefreshTokenRepository is the domain contract for refresh-token storage.
type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	FindActiveRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error)
	RotateRefreshToken(ctx context.Context, oldHash string, next *RefreshToken) error
	RevokeRefreshToken(ctx context.Context, hash string) error
}
