package usecase

import (
	"context"
	"errors"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenUsed     = errors.New("refresh token already used or expired")
)

type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	FindActiveRefreshTokenByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	RotateRefreshToken(ctx context.Context, oldHash string, next *domain.RefreshToken) error
	RevokeRefreshToken(ctx context.Context, hash string) error
}
