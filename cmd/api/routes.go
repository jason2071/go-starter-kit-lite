package main

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	httpHandler "github.com/jason2071/go-starter-kit-lite/internal/handler"
	"github.com/jason2071/go-starter-kit-lite/internal/middleware"
	"github.com/jason2071/go-starter-kit-lite/internal/platform"
	"github.com/jason2071/go-starter-kit-lite/internal/repository"
	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

type routeDependencies struct {
	DB     *gorm.DB
	Ready  func() error
	Config platform.Config
}

func registerRoutes(
	app *fiber.App,
	dependencies routeDependencies,
) {
	userRepository := repository.NewUserRepository(dependencies.DB)
	refreshTokenRepository := repository.NewRefreshTokenRepository(dependencies.DB)
	security := platform.NewSecurity(
		dependencies.Config.JWTSecret,
		dependencies.Config.JWTIssuer,
		dependencies.Config.JWTAudience,
		dependencies.Config.AccessTTL,
	)
	authService := usecase.NewAuthService(
		userRepository,
		refreshTokenRepository,
		security,
		security,
		dependencies.Config.RefreshTTL,
	)
	userService := usecase.NewUserService(userRepository)
	authHandler := httpHandler.NewAuthHandler(authService)
	userHandler := httpHandler.NewUserHandler(userService)
	systemHandler := httpHandler.NewSystemHandler(dependencies.Ready)

	api := app.Group("/api/v1")
	api.Post("/auth/register", authHandler.RegisterUser)
	api.Post("/auth/login", authHandler.LoginUser)
	api.Post("/auth/refresh", authHandler.RefreshToken)
	api.Post("/auth/logout", authHandler.LogoutUser)

	protected := api.Group("", middleware.Auth(security))
	protected.Get("/auth/me", authHandler.GetCurrentUser)
	protected.Get("/users", middleware.RequireRole("admin"), userHandler.ListUsers)

	app.Get("/healthz", systemHandler.Health)
	app.Get("/readyz", systemHandler.Ready)
	app.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/openapi.yaml")
	})
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger-ui.html")
	})
}
