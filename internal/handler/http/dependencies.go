package http

import (
	"log/slog"

	"github.com/go-playground/validator/v10"

	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

type Dependencies struct {
	Ready          func() error
	AuthService    *usecase.AuthService
	UserService    *usecase.UserService
	TokenManager   usecase.TokenManager
	Logger         *slog.Logger
	AllowedOrigins string
}

type Handler struct {
	ready       func() error
	authService *usecase.AuthService
	userService *usecase.UserService
	validator   *validator.Validate
}

func NewHandler(dependencies Dependencies) *Handler {
	return &Handler{
		ready:       dependencies.Ready,
		authService: dependencies.AuthService,
		userService: dependencies.UserService,
		validator:   validator.New(),
	}
}
