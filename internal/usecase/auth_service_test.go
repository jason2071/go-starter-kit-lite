package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
)

type fakeUserStore struct {
	users map[string]*domain.User
}

func (f *fakeUserStore) CreateUserWithRole(
	_ context.Context,
	user *domain.User,
	_ string,
) error {
	user.Roles = []domain.Role{{ID: 1, Name: "user"}}
	f.users[user.Email] = user
	return nil
}

func (f *fakeUserStore) FindUserByID(
	_ context.Context,
	id uuid.UUID,
) (*domain.User, error) {
	for _, user := range f.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (f *fakeUserStore) FindUserByEmail(
	_ context.Context,
	email string,
) (*domain.User, error) {
	user := f.users[email]
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (f *fakeUserStore) ListUsers(
	context.Context,
	ListUsersOptions,
) ([]domain.User, int64, error) {
	return nil, 0, nil
}

type fakeRefreshTokenStore struct {
	tokens map[string]*domain.RefreshToken
}

func (f *fakeRefreshTokenStore) CreateRefreshToken(
	_ context.Context,
	token *domain.RefreshToken,
) error {
	f.tokens[token.TokenHash] = token
	return nil
}

func (f *fakeRefreshTokenStore) FindActiveRefreshTokenByHash(
	_ context.Context,
	hash string,
) (*domain.RefreshToken, error) {
	token := f.tokens[hash]
	if token == nil {
		return nil, ErrRefreshTokenNotFound
	}
	return token, nil
}

func (f *fakeRefreshTokenStore) RotateRefreshToken(
	context.Context,
	string,
	*domain.RefreshToken,
) error {
	return nil
}

func (f *fakeRefreshTokenStore) RevokeRefreshToken(
	context.Context,
	string,
) error {
	return nil
}

type fakeSecurity struct{}

func (fakeSecurity) GenerateAccessToken(string, []string) (string, error) {
	return "access", nil
}

func (fakeSecurity) ParseAccessToken(string) (*AccessClaims, error) {
	return nil, nil
}

func (fakeSecurity) AccessTokenTTL() time.Duration {
	return 15 * time.Minute
}

func (fakeSecurity) GenerateRefreshToken() (string, string, error) {
	return "refresh", "hash", nil
}

func (fakeSecurity) HashRefreshToken(string) string {
	return "hash"
}

func (fakeSecurity) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.MinCost,
	)
	return string(hash), err
}

func (fakeSecurity) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}

func TestRegisterUser(t *testing.T) {
	userStore := &fakeUserStore{
		users: map[string]*domain.User{},
	}
	refreshTokenStore := &fakeRefreshTokenStore{
		tokens: map[string]*domain.RefreshToken{},
	}
	security := fakeSecurity{}

	service := NewAuthService(
		userStore,
		refreshTokenStore,
		security,
		security,
		24*time.Hour,
	)

	response, err := service.RegisterUser(
		context.Background(),
		RegisterRequest{
			Email:    "TEST@example.com",
			Password: "password123",
			Name:     "Test",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if response.User.Email != "test@example.com" {
		t.Fatalf(
			"unexpected email: %s",
			response.User.Email,
		)
	}

	if response.Tokens.AccessToken != "access" {
		t.Fatal("missing access token")
	}
}
