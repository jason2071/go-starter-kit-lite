package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/jason2071/go-starter-kit-lite/internal/middleware"
	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

type AuthHandler struct {
	authService *usecase.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService *usecase.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

func (h *AuthHandler) RegisterUser(c *fiber.Ctx) error {
	var request usecase.RegisterRequest
	if err := parseAndValidate(c, h.validator, &request); err != nil {
		return err
	}

	response, err := h.authService.RegisterUser(c.UserContext(), request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(
		dataResponse{Success: true, Data: response},
	)
}

func (h *AuthHandler) LoginUser(c *fiber.Ctx) error {
	var request usecase.LoginRequest
	if err := parseAndValidate(c, h.validator, &request); err != nil {
		return err
	}

	response, err := h.authService.LoginUser(c.UserContext(), request)
	if err != nil {
		return err
	}

	return c.JSON(dataResponse{Success: true, Data: response})
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var request usecase.RefreshTokenRequest
	if err := parseAndValidate(c, h.validator, &request); err != nil {
		return err
	}

	response, err := h.authService.RefreshToken(
		c.UserContext(),
		request.RefreshToken,
	)
	if err != nil {
		return err
	}

	return c.JSON(dataResponse{Success: true, Data: response})
}

func (h *AuthHandler) LogoutUser(c *fiber.Ctx) error {
	var request usecase.RefreshTokenRequest
	if err := parseAndValidate(c, h.validator, &request); err != nil {
		return err
	}

	if err := h.authService.LogoutUser(
		c.UserContext(),
		request.RefreshToken,
	); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) GetCurrentUser(c *fiber.Ctx) error {
	response, err := h.authService.GetCurrentUser(
		c.UserContext(),
		middleware.CurrentUser(c).ID,
	)
	if err != nil {
		return err
	}

	return c.JSON(dataResponse{Success: true, Data: response})
}
