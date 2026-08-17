package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

const authenticatedUserContextKey = "auth_user"

type AuthenticatedUser struct {
	ID    uuid.UUID
	Roles []string
}

func Auth(tokenManager usecase.TokenManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authorization := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
		if !strings.HasPrefix(authorization, "Bearer ") {
			return usecase.NewError(usecase.ErrUnauthorized, "UNAUTHORIZED", "missing bearer token")
		}

		claims, err := tokenManager.ParseAccessToken(strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")))
		if err != nil {
			return usecase.NewError(usecase.ErrUnauthorized, "UNAUTHORIZED", "invalid access token")
		}
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return usecase.NewError(usecase.ErrUnauthorized, "UNAUTHORIZED", "invalid access token")
		}

		c.Locals(authenticatedUserContextKey, AuthenticatedUser{ID: userID, Roles: claims.Roles})
		return c.Next()
	}
}

func RequireRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		for _, role := range CurrentUser(c).Roles {
			if role == requiredRole {
				return c.Next()
			}
		}
		return usecase.NewError(usecase.ErrForbidden, "FORBIDDEN", "insufficient permissions")
	}
}

func CurrentUser(c *fiber.Ctx) AuthenticatedUser {
	user, _ := c.Locals(authenticatedUserContextKey).(AuthenticatedUser)
	return user
}
