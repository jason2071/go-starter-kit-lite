package usecase

import (
	"context"
	"errors"
	"strings"
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

type UserListRequest struct {
	Page     int
	PageSize int
	Search   string
	IsActive *bool
	Sort     string
	Order    string
}

type UserListResponse struct {
	Items []UserResponse  `json:"items"`
	Meta  shared.PageMeta `json:"meta"`
}

type UserService struct{ users UserRepository }

func NewUserService(users UserRepository) *UserService { return &UserService{users: users} }

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, NewError(ErrNotFound, "USER_NOT_FOUND", "user not found")
		}
		return nil, WrapError(ErrInternal, "USER_READ_FAILED", "failed to read user", err)
	}
	response := toUserResponse(user)
	return &response, nil
}

func (s *UserService) List(ctx context.Context, req UserListRequest) (*UserListResponse, error) {
	page := shared.NormalizePage(req.Page, req.PageSize)
	order := strings.ToLower(req.Order)
	if order != "asc" {
		order = "desc"
	}
	users, total, err := s.users.List(ctx, UserListOptions{
		Page: page, Search: strings.TrimSpace(req.Search), IsActive: req.IsActive,
		Sort: req.Sort, Order: order,
	})
	if err != nil {
		return nil, WrapError(ErrInternal, "USER_LIST_FAILED", "failed to list users", err)
	}
	items := make([]UserResponse, 0, len(users))
	for i := range users {
		items = append(items, toUserResponse(&users[i]))
	}
	return &UserListResponse{Items: items, Meta: shared.NewPageMeta(page, total)}, nil
}

func toUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID: user.ID, Email: user.Email, Name: user.Name, IsActive: user.IsActive,
		Roles: user.RoleNames(), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
}
