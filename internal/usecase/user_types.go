package usecase

import (
	"context"
	"errors"
	"time"

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

type UserStore interface {
	CreateUserWithRole(ctx context.Context, user *domain.User, role string) error
	FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindUserByEmail(ctx context.Context, email string) (*domain.User, error)
	ListUsers(ctx context.Context, opts ListUsersOptions) ([]domain.User, int64, error)
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListUsersRequest struct {
	Page     int
	PageSize int
	Search   string
	IsActive *bool
	Sort     string
	Order    string
}

type ListUsersResponse struct {
	Items []UserResponse  `json:"items"`
	Meta  shared.PageMeta `json:"meta"`
}

func toUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsActive:  user.IsActive,
		Roles:     user.RoleNames(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
