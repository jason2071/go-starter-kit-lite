//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
	"github.com/jason2071/go-starter-kit-lite/internal/platform"
	"github.com/jason2071/go-starter-kit-lite/internal/repository"
)

func TestUserRepository(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	cfg := platform.Config{DatabaseURL: url, MaxOpenConns: 5, MaxIdleConns: 2}
	db, err := platform.NewDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewPostgres(db)
	u := &domain.User{ID: uuid.New(), Email: "integration-" + uuid.NewString() + "@example.com", PasswordHash: "x", Name: "Integration", IsActive: true}
	if err := repo.CreateWithRole(context.Background(), u, "user"); err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindByID(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != u.Email {
		t.Fatalf("got %s", got.Email)
	}
}
