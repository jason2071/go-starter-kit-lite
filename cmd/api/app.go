package main

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	appConfig "github.com/jason2071/go-starter-kit-lite/internal/config"
	"github.com/jason2071/go-starter-kit-lite/internal/handler"
	"github.com/jason2071/go-starter-kit-lite/internal/middleware"
)

func newApp(logger *slog.Logger, allowedOrigins string) *fiber.App {
	app := fiber.New(appConfig.NewFiberConfig(handler.ErrorHandler(logger)))

	app.Use(recover.New())
	app.Use(requestid.New(appConfig.NewRequestIDConfig()))
	app.Use(appConfig.NewCORSHandler(allowedOrigins))
	app.Use(middleware.RequestLogger(logger))

	return app
}
