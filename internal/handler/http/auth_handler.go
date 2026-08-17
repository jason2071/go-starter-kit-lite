package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jason2071/go-starter-kit-lite/internal/middleware"
	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

func (h *Handler) RegisterUser(c *fiber.Ctx) error {
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

func (h *Handler) LoginUser(c *fiber.Ctx) error {
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

func (h *Handler) RefreshToken(c *fiber.Ctx) error {
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

func (h *Handler) LogoutUser(c *fiber.Ctx) error {
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

func (h *Handler) GetCurrentUser(c *fiber.Ctx) error {
	response, err := h.authService.GetCurrentUser(
		c.UserContext(),
		middleware.CurrentUser(c).ID,
	)
	if err != nil {
		return err
	}

	return c.JSON(dataResponse{Success: true, Data: response})
}
