package http

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

func authMiddleware(tokenParser TokenParser) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authorization := strings.TrimSpace(
			c.Get(fiber.HeaderAuthorization),
		)
		if !strings.HasPrefix(authorization, "Bearer ") {
			return usecase.NewError(
				usecase.ErrUnauthorized,
				"UNAUTHORIZED",
				"missing bearer token",
			)
		}

		rawToken := strings.TrimSpace(
			strings.TrimPrefix(authorization, "Bearer "),
		)
		claims, err := tokenParser.ParseAccessToken(rawToken)
		if err != nil {
			return usecase.NewError(
				usecase.ErrUnauthorized,
				"UNAUTHORIZED",
				"invalid access token",
			)
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return usecase.NewError(
				usecase.ErrUnauthorized,
				"UNAUTHORIZED",
				"invalid access token",
			)
		}

		c.Locals(
			authenticatedUserContextKey,
			AuthenticatedUser{
				ID:    userID,
				Roles: claims.Roles,
			},
		)

		return c.Next()
	}
}

func requireRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		for _, role := range currentUser(c).Roles {
			if role == requiredRole {
				return c.Next()
			}
		}

		return usecase.NewError(
			usecase.ErrForbidden,
			"FORBIDDEN",
			"insufficient permissions",
		)
	}
}

func currentUser(c *fiber.Ctx) AuthenticatedUser {
	user, _ := c.Locals(authenticatedUserContextKey).(AuthenticatedUser)
	return user
}

func requestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		logger.Info(
			"http_request",
			"request_id", c.GetRespHeader(fiber.HeaderXRequestID),
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", time.Since(start).Milliseconds(),
		)

		return err
	}
}
