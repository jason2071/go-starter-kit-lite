package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handler) Ready(c *fiber.Ctx) error {
	if h.ready != nil && h.ready() != nil {
		return c.Status(fiber.StatusServiceUnavailable).
			JSON(fiber.Map{"status": "not_ready"})
	}

	return c.JSON(fiber.Map{"status": "ready"})
}
