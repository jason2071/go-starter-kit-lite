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
