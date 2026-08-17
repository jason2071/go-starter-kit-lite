package usecase

import (
	"time"

	"github.com/google/uuid"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
	"github.com/jason2071/go-starter-kit-lite/internal/shared"
)

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
