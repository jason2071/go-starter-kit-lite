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

func NewApp(dependencies Dependencies) *fiber.App {
	app := fiber.New(
		fiber.Config{
			AppName:               "Go Fiber Clean Lite",
			DisableStartupMessage: true,
			ReadTimeout:           10 * time.Second,
			WriteTimeout:          10 * time.Second,
			IdleTimeout:           30 * time.Second,
			BodyLimit:             4 * 1024 * 1024,
			ErrorHandler:          errorHandler(dependencies.Logger),
		},
	)

	app.Use(recover.New())
	app.Use(
		requestid.New(
			requestid.Config{
				Header:    fiber.HeaderXRequestID,
				Generator: utils.UUIDv4,
			},
		),
	)
	app.Use(
		cors.New(
			cors.Config{
				AllowOrigins: dependencies.AllowedOrigins,
				AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Request-ID",
				AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
			},
		),
	)
	app.Use(requestLogger(dependencies.Logger))

	registerRoutes(
		app,
		NewHandler(dependencies),
		dependencies.TokenManager,
	)

	return app
}

func registerRoutes(
	app *fiber.App,
	handler *Handler,
	tokenParser TokenParser,
) {
	api := app.Group("/api/v1")

	api.Post("/auth/register", handler.RegisterUser)
	api.Post("/auth/login", handler.LoginUser)
	api.Post("/auth/refresh", handler.RefreshToken)
	api.Post("/auth/logout", handler.LogoutUser)

	protected := api.Group("", authMiddleware(tokenParser))
	protected.Get("/auth/me", handler.GetCurrentUser)
	protected.Get("/users", requireRole("admin"), handler.ListUsers)

	app.Get("/healthz", handler.Health)
	app.Get("/readyz", handler.Ready)
	app.Get(
		"/openapi.yaml",
		func(c *fiber.Ctx) error {
			return c.SendFile("./docs/openapi.yaml")
		},
	)
	app.Get(
		"/docs",
		func(c *fiber.Ctx) error {
			return c.Type("html").SendString(swaggerHTML)
		},
	)
}

type TokenParser interface {
	ParseAccessToken(raw string) (*usecase.AccessClaims, error)
}

const swaggerHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>API Docs</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
SwaggerUIBundle({url:'/openapi.yaml',dom_id:'#swagger-ui'});
</script>
</body>
</html>`
