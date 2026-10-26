package routes

import (
	"go-url-shortner/db"
	"go-url-shortner/helpers"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"custom_short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"custom_short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"x_rate_remaining"`
	XRateLimitReset time.Duration `json:"x_rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}
	r2 := db.CreateClient(1)
	defer r2.Close()
	val, err := r2.Get(c.Context(), c.IP()).Result()
	if err == redis.Nil {
		_ = r2.Set(c.Context(), c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second)
		val = os.Getenv("API_QUOTA")
	} else {
		val, _ = r2.Get(c.Context(), c.IP()).Result()
		valInt, err := strconv.Atoi(val)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "cannot parse integer",
			})
		}
		if valInt <= 0 {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":            "rate limit exceeded",
				"rate_limit_reset": time.Hour * time.Duration(valInt),
			})
		}
	}

	//checking if the input is an actul url
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	//check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	//enforce ssl, hhtps
	body.URL = helpers.EnforceHTTPS(body.URL)

	var id string
	if body.CustomShort != "" {
		id = body.CustomShort
	} else {
		id = uuid.New().String()[:6]
	}

	r := db.CreateClient(0)
	defer r.Close()

	val, err = r.Get(c.Context(), id).Result()
	if err == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "custom short already exists",
		})
	}
	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(c.Context(), id, body.URL, body.Expiry*time.Hour).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot set key",
		})
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     id,
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}
	r2.Decr(c.Context(), c.IP())
	val, _ = r2.Get(c.Context(), c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(c.Context(), c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute
	resp.CustomShort = os.Getenv("DOMAIN") + "/" + resp.CustomShort

	return c.Status(fiber.StatusOK).JSON(resp)
}
