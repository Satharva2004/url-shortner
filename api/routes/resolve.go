package routes

import (
	"go-url-shortner/db"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url")
	_ = url

	r := db.CreateClient(0)

	value, err := r.Get(c.Context(), url).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "url not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}
	rInr := db.CreateClient(1)
	defer rInr.Close()

	rInr.Incr(c.Context(), "clicks_"+url)

	return c.Redirect(value, fiber.StatusSeeOther)
}
