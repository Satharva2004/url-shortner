package routes

import (
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url")

	// TODO: look up the short URL in the database and redirect
	_ = url

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "not implemented yet",
	})
}
