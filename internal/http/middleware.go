package http

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
)

const RequestIDKey = "X-Request-Id"

func middlewareRecover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "internal error",
				})
			}
		}()
		return c.Next()
	}
}

func middlewareRequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(RequestIDKey)
		if id == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Set(RequestIDKey, id)
		c.Locals("request_id", id)
		return c.Next()
	}
}

func middlewareRequestLog(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		dur := time.Since(start)
		log.Info("http",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"dur_ms", dur.Milliseconds(),
			"req_id", c.Locals("request_id"),
			"ip", c.IP(),
		)
		return err
	}
}

func middlewareCORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Request-Id")
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Next()
	}
}

func middlewareJWT(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasPrefix(path, "/api/auth/login") ||
			strings.HasPrefix(path, "/api/auth/refresh") ||
			strings.HasPrefix(path, "/api/auth/setup") ||
			strings.HasPrefix(path, "/api/health") ||
			strings.HasPrefix(path, "/api/logs/") ||
			strings.HasPrefix(path, "/api/setup-info") ||
			strings.HasPrefix(path, "/api/_echo") {
			a.Log.Debug("jwt bypass", "path", path)
			return c.Next()
		}
		authz := c.Get("Authorization")
		a.Log.Debug("jwt check", "path", path, "has_authz", authz != "", "token_prefix", tokenPrefix(authz))
		const prefix = "Bearer "
		if !strings.HasPrefix(authz, prefix) {
			return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
		}
		token := strings.TrimPrefix(authz, prefix)
		uid, err := a.Auth.Parse(token)
		if err != nil {
			if err == jwt.ErrTokenExpired || strings.Contains(err.Error(), "token expired") {
				a.Log.Warn("jwt expired", "path", path, "token_prefix", tokenPrefix(token), "err", err.Error())
				return fiber.NewError(fiber.StatusUnauthorized, "token expired")
			}
			a.Log.Warn("jwt invalid", "path", path, "token_prefix", tokenPrefix(token), "err", err.Error())
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}
		c.Locals("user_id", uid)
		return c.Next()
	}
}

func tokenPrefix(s string) string {
	s = strings.TrimPrefix(s, "Bearer ")
	if len(s) > 30 {
		return s[:30] + "..."
	}
	return s
}
