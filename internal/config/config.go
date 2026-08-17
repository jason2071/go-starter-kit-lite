package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
)

func NewFiberConfig(errorHandler fiber.ErrorHandler) fiber.Config {
	return fiber.Config{
		AppName:               "Go Fiber Clean Lite",
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           30 * time.Second,
		BodyLimit:             4 * 1024 * 1024,
		ErrorHandler:          errorHandler,
	}
}

func NewRequestIDConfig() requestid.Config {
	return requestid.Config{Header: fiber.HeaderXRequestID, Generator: utils.UUIDv4}
}

func NewCORSHandler(allowedOrigins string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Request-ID",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	})
}
