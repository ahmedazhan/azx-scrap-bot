package http

import (
	"net/http"
	"os"
	"path/filepath"
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

		if a.Cfg.AssetsDir != "" {
			if ok, body := tryServeFromDisk(a.Cfg.AssetsDir, path); ok {
				c.Set("Content-Type", "text/html; charset=utf-8")
				return c.Status(http.StatusOK).Send(body)
			}
		}

		if a.SPAHTML != "" {
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.Status(http.StatusOK).SendString(a.SPAHTML)
		}

		return c.Status(http.StatusOK).SendString(`<!doctype html><html><body><h1>azx-scrap-bot</h1></body></html>`)
	}
}

func tryServeFromDisk(root, reqPath string) (bool, []byte) {
	clean := strings.TrimPrefix(reqPath, "/")
	if clean == "" {
		clean = "index.html"
	}
	full := filepath.Join(root, clean)
	rel, err := filepath.Rel(root, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return false, nil
	}
	if info, err := os.Stat(full); err == nil && !info.IsDir() {
		b, err := os.ReadFile(full)
		if err == nil {
			return true, b
		}
	}
	idx, err := os.ReadFile(filepath.Join(root, "index.html"))
	if err != nil {
		return false, nil
	}
	return true, idx
}
