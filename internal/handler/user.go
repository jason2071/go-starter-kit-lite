package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

type UserHandler struct {
	userService *usecase.UserService
}

func NewUserHandler(userService *usecase.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	var isActive *bool
	if rawIsActive := c.Query("is_active"); rawIsActive != "" {
		value, err := strconv.ParseBool(rawIsActive)
		if err != nil {
			return usecase.NewError(
				usecase.ErrValidation,
				"INVALID_FILTER",
				"is_active must be boolean",
			)
		}
		isActive = &value
	}

	response, err := h.userService.ListUsers(
		c.UserContext(),
		usecase.ListUsersRequest{
			Page:     page,
			PageSize: pageSize,
			Search:   c.Query("search"),
			IsActive: isActive,
			Sort:     c.Query("sort", "created_at"),
			Order:    c.Query("order", "desc"),
		},
	)
	if err != nil {
		return err
	}

	return c.JSON(dataResponse{Success: true, Data: response})
}
