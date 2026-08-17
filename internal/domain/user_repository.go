package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"

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

// UserRepository is the domain contract for storing and finding users.
// Infrastructure packages, such as repository, implement this interface.
type UserRepository interface {
	CreateUserWithRole(ctx context.Context, user *User, role string) error
	FindUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context, opts ListUsersOptions) ([]User, int64, error)
}
