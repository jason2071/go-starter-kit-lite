package handler

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

type errorResponse struct {
	Success bool     `json:"success"`
	Error   apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type dataResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
}

func parseAndValidate(
	c *fiber.Ctx,
	requestValidator *validator.Validate,
	target any,
) error {
	if err := c.BodyParser(target); err != nil {
		return usecase.NewError(
			usecase.ErrValidation,
			"INVALID_REQUEST",
			"invalid request body",
		)
	}

	if err := requestValidator.Struct(target); err != nil {
		return &usecase.AppError{
			Kind:    usecase.ErrValidation,
			Code:    "VALIDATION_ERROR",
			Message: "validation failed",
			Details: validationDetails(err),
		}
	}

	return nil
}

func validationDetails(err error) map[string]string {
	details := map[string]string{}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, validationError := range validationErrors {
			details[strings.ToLower(validationError.Field())] = "invalid value"
		}
	}

	return details
}

func errorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var appError *usecase.AppError
		if errors.As(err, &appError) {
			status := map[usecase.ErrorKind]int{
				usecase.ErrValidation:   fiber.StatusBadRequest,
				usecase.ErrUnauthorized: fiber.StatusUnauthorized,
				usecase.ErrForbidden:    fiber.StatusForbidden,
				usecase.ErrNotFound:     fiber.StatusNotFound,
				usecase.ErrConflict:     fiber.StatusConflict,
				usecase.ErrInternal:     fiber.StatusInternalServerError,
			}[appError.Kind]

			if status == 0 {
				status = fiber.StatusInternalServerError
			}

			if status >= fiber.StatusInternalServerError {
				logger.Error(
					"request_failed",
					"error", appError.Err,
					"code", appError.Code,
				)
			}

			return c.Status(status).JSON(
				errorResponse{
					Success: false,
					Error: apiError{
						Code:    appError.Code,
						Message: appError.Message,
						Details: appError.Details,
					},
				},
			)
		}

		var fiberError *fiber.Error
		if errors.As(err, &fiberError) {
			return c.Status(fiberError.Code).JSON(
				errorResponse{
					Success: false,
					Error: apiError{
						Code:    "HTTP_ERROR",
						Message: fiberError.Message,
					},
				},
			)
		}

		logger.Error("unhandled_error", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(
			errorResponse{
				Success: false,
				Error: apiError{
					Code:    "INTERNAL_ERROR",
					Message: "internal server error",
				},
			},
		)
	}
}
