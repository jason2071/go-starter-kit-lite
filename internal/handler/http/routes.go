package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"

	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

// NewApp configures the Fiber application and registers all routes.
func NewApp(dep Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "Go Fiber Clean Lite",
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           30 * time.Second,
		BodyLimit:             4 * 1024 * 1024,
		ErrorHandler:          errorHandler(dep.Logger),
	})
	app.Use(recover.New())
	app.Use(requestid.New(requestid.Config{Header: fiber.HeaderXRequestID, Generator: utils.UUIDv4}))
	app.Use(cors.New(cors.Config{AllowOrigins: dep.AllowedOrigins, AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Request-ID", AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS"}))
	app.Use(requestLogger(dep.Logger))

	registerRoutes(app, NewHandler(dep), dep.Tokens)
	return app
}

func registerRoutes(app *fiber.App, handler *Handler, tokens TokenParser) {
	api := app.Group("/api/v1")
	api.Post("/auth/register", handler.Register)
	api.Post("/auth/login", handler.Login)
	api.Post("/auth/refresh", handler.Refresh)
	api.Post("/auth/logout", handler.Logout)

	protected := api.Group("", authMiddleware(tokens))
	protected.Get("/auth/me", handler.Me)
	protected.Get("/users", requireRole("admin"), handler.ListUsers)

	app.Get("/healthz", handler.Health)
	app.Get("/readyz", handler.Ready)
	app.Get("/openapi.yaml", func(c *fiber.Ctx) error { return c.SendFile("./docs/openapi.yaml") })
	app.Get("/docs", func(c *fiber.Ctx) error { return c.Type("html").SendString(swaggerHTML) })
}

// TokenParser keeps routing dependent on the authentication port only.
type TokenParser interface {
	ParseAccessToken(raw string) (*usecase.AccessClaims, error)
}
