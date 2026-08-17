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
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	cfg := platform.Config{
		DatabaseURL:  databaseURL,
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	}

	db, err := platform.NewDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}

	userRepository := repository.NewUserRepository(db)

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "integration-" + uuid.NewString() + "@example.com",
		PasswordHash: "x",
		Name:         "Integration",
		IsActive:     true,
	}

	if err := userRepository.CreateUserWithRole(
		context.Background(),
		user,
		"user",
	); err != nil {
		t.Fatal(err)
	}

	createdUser, err := userRepository.FindUserByID(
		context.Background(),
		user.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if createdUser.Email != user.Email {
		t.Fatalf("got %s", createdUser.Email)
	}
}
