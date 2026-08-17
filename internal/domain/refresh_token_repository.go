package domain

import (
	"context"
	"errors"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenUsed     = errors.New("refresh token already used or expired")
)

// RefreshTokenRepository is the domain contract for refresh-token storage.
type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	FindActiveRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error)
	RotateRefreshToken(ctx context.Context, oldHash string, next *RefreshToken) error
	RevokeRefreshToken(ctx context.Context, hash string) error
}
