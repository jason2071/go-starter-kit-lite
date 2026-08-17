package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	appConfig "github.com/jason2071/go-starter-kit-lite/internal/config"
	"github.com/jason2071/go-starter-kit-lite/internal/middleware"
)

func NewApp(dependencies Dependencies) *fiber.App {
	app := fiber.New(appConfig.NewFiberConfig(errorHandler(dependencies.Logger)))

	app.Use(recover.New())
	app.Use(requestid.New(appConfig.NewRequestIDConfig()))
	app.Use(appConfig.NewCORSHandler(dependencies.AllowedOrigins))
	app.Use(middleware.RequestLogger(dependencies.Logger))

	return app
}
