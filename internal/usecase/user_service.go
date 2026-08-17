package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
	"github.com/jason2071/go-starter-kit-lite/internal/shared"
)

type UserService struct {
	userRepository domain.UserRepository
}

func NewUserService(userRepository domain.UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepository.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, NewError(ErrNotFound, "USER_NOT_FOUND", "user not found")
		}
		return nil, WrapError(ErrInternal, "USER_READ_FAILED", "failed to read user", err)
	}

	response := toUserResponse(user)
	return &response, nil
}

func (s *UserService) ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	page := shared.NormalizePage(req.Page, req.PageSize)

	order := strings.ToLower(req.Order)
	if order != "asc" {
		order = "desc"
	}

	users, total, err := s.userRepository.ListUsers(ctx, domain.ListUsersOptions{
		Page:     page,
		Search:   strings.TrimSpace(req.Search),
		IsActive: req.IsActive,
		Sort:     req.Sort,
		Order:    order,
	})
	if err != nil {
		return nil, WrapError(ErrInternal, "USER_LIST_FAILED", "failed to list users", err)
	}

	items := make([]UserResponse, 0, len(users))
	for i := range users {
		items = append(items, toUserResponse(&users[i]))
	}

	return &ListUsersResponse{
		Items: items,
		Meta:  shared.NewPageMeta(page, total),
	}, nil
}
