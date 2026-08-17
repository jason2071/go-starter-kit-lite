package handler

import "github.com/gofiber/fiber/v2"

type SystemHandler struct {
	ready func() error
}

func NewSystemHandler(ready func() error) *SystemHandler {
	return &SystemHandler{ready: ready}
}

func (h *SystemHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *SystemHandler) Ready(c *fiber.Ctx) error {
	if h.ready != nil && h.ready() != nil {
		return c.Status(fiber.StatusServiceUnavailable).
			JSON(fiber.Map{"status": "not_ready"})
	}

	return c.JSON(fiber.Map{"status": "ready"})
}
