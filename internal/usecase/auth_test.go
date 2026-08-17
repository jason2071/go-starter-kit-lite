package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
)

type fakeRepo struct {
	users   map[string]*domain.User
	refresh map[string]*domain.RefreshToken
}

func (f *fakeRepo) CreateWithRole(_ context.Context, u *domain.User, _ string) error {
	u.Roles = []domain.Role{{ID: 1, Name: "user"}}
	f.users[u.Email] = u
	return nil
}
func (f *fakeRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}
func (f *fakeRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	u := f.users[email]
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}
func (f *fakeRepo) List(context.Context, UserListOptions) ([]domain.User, int64, error) {
	return nil, 0, nil
}
func (f *fakeRepo) Create(_ context.Context, t *domain.RefreshToken) error {
	f.refresh[t.TokenHash] = t
	return nil
}
func (f *fakeRepo) FindActiveByHash(_ context.Context, h string) (*domain.RefreshToken, error) {
	t := f.refresh[h]
	if t == nil {
		return nil, ErrRefreshNotFound
	}
	return t, nil
}
func (f *fakeRepo) Rotate(context.Context, string, *domain.RefreshToken) error { return nil }
func (f *fakeRepo) Revoke(context.Context, string) error                       { return nil }

type fakeSecurity struct{}

func (fakeSecurity) GenerateAccessToken(string, []string) (string, error) { return "access", nil }
func (fakeSecurity) ParseAccessToken(string) (*AccessClaims, error)       { return nil, nil }
func (fakeSecurity) AccessTokenTTL() time.Duration                        { return 15 * time.Minute }
func (fakeSecurity) GenerateRefreshToken() (string, string, error)        { return "refresh", "hash", nil }
func (fakeSecurity) HashRefreshToken(string) string                       { return "hash" }
func (fakeSecurity) Hash(p string) (string, error) {
	b, e := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
	return string(b), e
}
func (fakeSecurity) Compare(h, p string) error {
	return bcrypt.CompareHashAndPassword([]byte(h), []byte(p))
}

func TestRegister(t *testing.T) {
	repo := &fakeRepo{users: map[string]*domain.User{}, refresh: map[string]*domain.RefreshToken{}}
	sec := fakeSecurity{}
	svc := NewAuthService(repo, repo, sec, sec, 24*time.Hour)
	out, err := svc.Register(context.Background(), RegisterRequest{Email: "TEST@example.com", Password: "password123", Name: "Test"})
	if err != nil {
		t.Fatal(err)
	}
	if out.User.Email != "test@example.com" {
		t.Fatalf("unexpected email: %s", out.User.Email)
	}
	if out.Tokens.AccessToken != "access" {
		t.Fatal("missing access token")
	}
}
