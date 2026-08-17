package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
	"github.com/jason2071/go-starter-kit-lite/internal/shared"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

type ListUsersOptions struct {
	Page     shared.PageParams
	Search   string
	IsActive *bool
	Sort     string
	Order    string
}

type UserRepository interface {
	CreateUserWithRole(ctx context.Context, user *domain.User, role string) error
	FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindUserByEmail(ctx context.Context, email string) (*domain.User, error)
	ListUsers(ctx context.Context, opts ListUsersOptions) ([]domain.User, int64, error)
}
