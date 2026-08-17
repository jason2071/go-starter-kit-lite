package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"name" validate:"required,min=2,max=255"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,max=72"`
}

type RefreshRequest struct {
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

type AuthService struct {
	users      UserRepository
	refresh    RefreshTokenRepository
	tokens     TokenManager
	passwords  PasswordHasher
	refreshTTL time.Duration
}

func NewAuthService(users UserRepository, refresh RefreshTokenRepository, tokens TokenManager, passwords PasswordHasher, refreshTTL time.Duration) *AuthService {
	return &AuthService{users: users, refresh: refresh, tokens: tokens, passwords: passwords, refreshTTL: refreshTTL}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, NewError(ErrConflict, "EMAIL_ALREADY_EXISTS", "email already exists")
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, WrapError(ErrInternal, "USER_LOOKUP_FAILED", "failed to check user", err)
	}

	hash, err := s.passwords.Hash(req.Password)
	if err != nil {
		return nil, WrapError(ErrInternal, "PASSWORD_HASH_FAILED", "failed to secure password", err)
	}
	user := &domain.User{ID: uuid.New(), Email: email, PasswordHash: hash, Name: strings.TrimSpace(req.Name), IsActive: true}
	if err := s.users.CreateWithRole(ctx, user, "user"); err != nil {
		if errors.Is(err, ErrEmailExists) {
			return nil, NewError(ErrConflict, "EMAIL_ALREADY_EXISTS", "email already exists")
		}
		return nil, WrapError(ErrInternal, "USER_CREATE_FAILED", "failed to create user", err)
	}
	created, err := s.users.FindByID(ctx, user.ID)
	if err != nil {
		return nil, WrapError(ErrInternal, "USER_READ_FAILED", "failed to read created user", err)
	}
	pair, err := s.issueTokenPair(ctx, created)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{User: toUserResponse(created), Tokens: *pair}, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, NewError(ErrUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		}
		return nil, WrapError(ErrInternal, "LOGIN_FAILED", "failed to login", err)
	}
	if !user.IsActive {
		return nil, NewError(ErrForbidden, "USER_DISABLED", "user account is disabled")
	}
	if err := s.passwords.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, NewError(ErrUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
	}
	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{User: toUserResponse(user), Tokens: *pair}, nil
}

func (s *AuthService) Refresh(ctx context.Context, raw string) (*TokenPair, error) {
	oldHash := s.tokens.HashRefreshToken(raw)
	stored, err := s.refresh.FindActiveByHash(ctx, oldHash)
	if err != nil {
		if errors.Is(err, ErrRefreshNotFound) {
			return nil, NewError(ErrUnauthorized, "INVALID_REFRESH_TOKEN", "invalid or expired refresh token")
		}
		return nil, WrapError(ErrInternal, "REFRESH_FAILED", "failed to refresh token", err)
	}
	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil || user == nil || !user.IsActive {
		return nil, NewError(ErrUnauthorized, "INVALID_REFRESH_TOKEN", "invalid or expired refresh token")
	}
	access, err := s.tokens.GenerateAccessToken(user.ID.String(), user.RoleNames())
	if err != nil {
		return nil, WrapError(ErrInternal, "TOKEN_CREATE_FAILED", "failed to create access token", err)
	}
	rawNext, hashNext, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, WrapError(ErrInternal, "TOKEN_CREATE_FAILED", "failed to create refresh token", err)
	}
	next := &domain.RefreshToken{ID: uuid.New(), UserID: user.ID, TokenHash: hashNext, ExpiresAt: time.Now().UTC().Add(s.refreshTTL)}
	if err := s.refresh.Rotate(ctx, oldHash, next); err != nil {
		if errors.Is(err, ErrRefreshTokenUsed) {
			return nil, NewError(ErrUnauthorized, "INVALID_REFRESH_TOKEN", "refresh token was already used or expired")
		}
		return nil, WrapError(ErrInternal, "REFRESH_FAILED", "failed to rotate refresh token", err)
	}
	return &TokenPair{AccessToken: access, RefreshToken: rawNext, TokenType: "Bearer", ExpiresIn: int64(s.tokens.AccessTokenTTL().Seconds())}, nil
}

func (s *AuthService) Logout(ctx context.Context, raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	if err := s.refresh.Revoke(ctx, s.tokens.HashRefreshToken(raw)); err != nil {
		return WrapError(ErrInternal, "LOGOUT_FAILED", "failed to logout", err)
	}
	return nil
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, NewError(ErrNotFound, "USER_NOT_FOUND", "user not found")
		}
		return nil, WrapError(ErrInternal, "USER_READ_FAILED", "failed to read user", err)
	}
	response := toUserResponse(user)
	return &response, nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, user *domain.User) (*TokenPair, error) {
	access, err := s.tokens.GenerateAccessToken(user.ID.String(), user.RoleNames())
	if err != nil {
		return nil, WrapError(ErrInternal, "TOKEN_CREATE_FAILED", "failed to create access token", err)
	}
	raw, hash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, WrapError(ErrInternal, "TOKEN_CREATE_FAILED", "failed to create refresh token", err)
	}
	refresh := &domain.RefreshToken{ID: uuid.New(), UserID: user.ID, TokenHash: hash, ExpiresAt: time.Now().UTC().Add(s.refreshTTL)}
	if err := s.refresh.Create(ctx, refresh); err != nil {
		return nil, WrapError(ErrInternal, "TOKEN_STORE_FAILED", "failed to store refresh token", err)
	}
	return &TokenPair{AccessToken: access, RefreshToken: raw, TokenType: "Bearer", ExpiresIn: int64(s.tokens.AccessTokenTTL().Seconds())}, nil
}
