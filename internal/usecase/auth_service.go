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

type AuthService struct {
	userRepository         UserRepository
	refreshTokenRepository RefreshTokenRepository
	tokenManager           TokenManager
	passwordHasher         PasswordHasher
	refreshTokenTTL        time.Duration
}

func NewAuthService(
	userRepository UserRepository,
	refreshTokenRepository RefreshTokenRepository,
	tokenManager TokenManager,
	passwordHasher PasswordHasher,
	refreshTokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		tokenManager:           tokenManager,
		passwordHasher:         passwordHasher,
		refreshTokenTTL:        refreshTokenTTL,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if _, err := s.userRepository.FindUserByEmail(ctx, email); err == nil {
		return nil, NewError(ErrConflict, "EMAIL_ALREADY_EXISTS", "email already exists")
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, WrapError(ErrInternal, "USER_LOOKUP_FAILED", "failed to check user", err)
	}

	passwordHash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, WrapError(ErrInternal, "PASSWORD_HASH_FAILED", "failed to secure password", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Name:         strings.TrimSpace(req.Name),
		IsActive:     true,
	}

	if err := s.userRepository.CreateUserWithRole(ctx, user, "user"); err != nil {
		if errors.Is(err, ErrEmailExists) {
			return nil, NewError(ErrConflict, "EMAIL_ALREADY_EXISTS", "email already exists")
		}
		return nil, WrapError(ErrInternal, "USER_CREATE_FAILED", "failed to create user", err)
	}

	createdUser, err := s.userRepository.FindUserByID(ctx, user.ID)
	if err != nil {
		return nil, WrapError(ErrInternal, "USER_READ_FAILED", "failed to read created user", err)
	}

	tokenPair, err := s.createTokenPair(ctx, createdUser)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:   toUserResponse(createdUser),
		Tokens: *tokenPair,
	}, nil
}

func (s *AuthService) LoginUser(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepository.FindUserByEmail(
		ctx,
		strings.ToLower(strings.TrimSpace(req.Email)),
	)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, NewError(ErrUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		}
		return nil, WrapError(ErrInternal, "LOGIN_FAILED", "failed to login", err)
	}

	if !user.IsActive {
		return nil, NewError(ErrForbidden, "USER_DISABLED", "user account is disabled")
	}

	if err := s.passwordHasher.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, NewError(ErrUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
	}

	tokenPair, err := s.createTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:   toUserResponse(user),
		Tokens: *tokenPair,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	oldHash := s.tokenManager.HashRefreshToken(rawRefreshToken)

	storedToken, err := s.refreshTokenRepository.FindActiveRefreshTokenByHash(ctx, oldHash)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenNotFound) {
			return nil, NewError(ErrUnauthorized, "INVALID_REFRESH_TOKEN", "invalid or expired refresh token")
		}
		return nil, WrapError(ErrInternal, "REFRESH_FAILED", "failed to refresh token", err)
	}

	user, err := s.userRepository.FindUserByID(ctx, storedToken.UserID)
	if err != nil || user == nil || !user.IsActive {
		return nil, NewError(ErrUnauthorized, "INVALID_REFRESH_TOKEN", "invalid or expired refresh token")
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID.String(), user.RoleNames())
	if err != nil {
		return nil, WrapError(ErrInternal, "TOKEN_CREATE_FAILED", "failed to create access token", err)
	}

	nextRawToken, nextHash, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, WrapError(ErrInternal, "TOKEN_CREATE_FAILED", "failed to create refresh token", err)
	}

	nextToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: nextHash,
		ExpiresAt: time.Now().UTC().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepository.RotateRefreshToken(ctx, oldHash, nextToken); err != nil {
		if errors.Is(err, ErrRefreshTokenUsed) {
			return nil, NewError(ErrUnauthorized, "INVALID_REFRESH_TOKEN", "refresh token was already used or expired")
		}
		return nil, WrapError(ErrInternal, "REFRESH_FAILED", "failed to rotate refresh token", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: nextRawToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenManager.AccessTokenTTL().Seconds()),
	}, nil
}

func (s *AuthService) LogoutUser(ctx context.Context, rawRefreshToken string) error {
	if strings.TrimSpace(rawRefreshToken) == "" {
		return nil
	}

	hash := s.tokenManager.HashRefreshToken(rawRefreshToken)
	if err := s.refreshTokenRepository.RevokeRefreshToken(ctx, hash); err != nil {
		return WrapError(ErrInternal, "LOGOUT_FAILED", "failed to logout", err)
	}
	return nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepository.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, NewError(ErrNotFound, "USER_NOT_FOUND", "user not found")
		}
		return nil, WrapError(ErrInternal, "USER_READ_FAILED", "failed to read user", err)
	}

	response := toUserResponse(user)
	return &response, nil
}

func (s *AuthService) createTokenPair(ctx context.Context, user *domain.User) (*TokenPair, error) {
	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID.String(), user.RoleNames())
	if err != nil {
		return nil, WrapError(ErrInternal, "TOKEN_CREATE_FAILED", "failed to create access token", err)
	}

	rawRefreshToken, refreshTokenHash, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, WrapError(ErrInternal, "TOKEN_CREATE_FAILED", "failed to create refresh token", err)
	}

	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().UTC().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepository.CreateRefreshToken(ctx, refreshToken); err != nil {
		return nil, WrapError(ErrInternal, "TOKEN_STORE_FAILED", "failed to store refresh token", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenManager.AccessTokenTTL().Seconds()),
	}, nil
}
