package usecase

import "time"

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
