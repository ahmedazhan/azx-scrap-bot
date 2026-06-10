package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
	"github.com/ahmedazhan/azx-scrap-bot/internal/auth"
	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
)

type accountUpdate struct {
	Theme      *string `json:"theme"`
	FilterMode *string `json:"filter_mode"`
}

func GetAccount(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid := c.Locals("user_id").(uint64)
		var u db.User
		if err := a.DB.First(&u, uid).Error; err != nil {
			return badRequest(c, "user not found")
		}
		return ok(c, fiber.Map{
			"username":    u.Username,
			"theme":       u.Theme,
			"filter_mode": u.FilterMode,
		}, nil)
	}
}

func UpdateAccount(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid := c.Locals("user_id").(uint64)
		var req accountUpdate
		if err := c.BodyParser(&req); err != nil {
			return badRequest(c, "invalid body")
		}
		var u db.User
		if err := a.DB.First(&u, uid).Error; err != nil {
			return badRequest(c, "user not found")
		}
		if req.Theme != nil {
			u.Theme = *req.Theme
		}
		if req.FilterMode != nil {
			u.FilterMode = *req.FilterMode
		}
		a.DB.Save(&u)
		return ok(c, fiber.Map{
			"username":    u.Username,
			"theme":       u.Theme,
			"filter_mode": u.FilterMode,
		}, nil)
	}
}

var _ = auth.HashPassword
