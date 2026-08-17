package main

import (
	"github.com/gofiber/fiber/v2"

	httpHandler "github.com/jason2071/go-starter-kit-lite/internal/handler"
	"github.com/jason2071/go-starter-kit-lite/internal/middleware"
	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

func registerRoutes(
	app *fiber.App,
	handler *httpHandler.Handler,
	tokenManager usecase.TokenManager,
) {
	api := app.Group("/api/v1")
	api.Post("/auth/register", handler.RegisterUser)
	api.Post("/auth/login", handler.LoginUser)
	api.Post("/auth/refresh", handler.RefreshToken)
	api.Post("/auth/logout", handler.LogoutUser)

	protected := api.Group("", middleware.Auth(tokenManager))
	protected.Get("/auth/me", handler.GetCurrentUser)
	protected.Get("/users", middleware.RequireRole("admin"), handler.ListUsers)

	app.Get("/healthz", handler.Health)
	app.Get("/readyz", handler.Ready)
	app.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/openapi.yaml")
	})
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger-ui.html")
	})
}
