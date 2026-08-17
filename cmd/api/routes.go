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
	handler := httpHandler.NewHandler(httpHandler.Dependencies{
		Ready:       dependencies.Ready,
		AuthService: authService,
		UserService: userService,
	})

	api := app.Group("/api/v1")
	api.Post("/auth/register", handler.RegisterUser)
	api.Post("/auth/login", handler.LoginUser)
	api.Post("/auth/refresh", handler.RefreshToken)
	api.Post("/auth/logout", handler.LogoutUser)

	protected := api.Group("", middleware.Auth(security))
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
