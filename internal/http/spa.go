package http

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
)

func spaHandler(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasPrefix(path, "/api/") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   "not found",
			})
		}
		if a.SPAHTML != "" {
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.Status(http.StatusOK).SendString(a.SPAHTML)
		}
		return c.Status(http.StatusOK).SendString(`<!doctype html><html><body><h1>azx-scrap-bot</h1></body></html>`)
	}
}
