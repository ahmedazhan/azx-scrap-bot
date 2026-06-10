package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
)

func ListSources(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var rows []db.Source
		if err := a.DB.Find(&rows).Error; err != nil {
			return serverError(c, err)
		}
		return ok(c, rows, nil)
	}
}

func GetSource(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Params("key")
		var row db.Source
		if err := a.DB.Where("key = ?", key).First(&row).Error; err != nil {
			return notFound(c, "source not found")
		}
		return ok(c, row, nil)
	}
}

type sourceUpdate struct {
	Enabled          *bool `json:"enabled"`
	IntervalSec      *int  `json:"interval_sec"`
	Concurrency      *int  `json:"concurrency"`
	MaxPagesPerCycle *int  `json:"max_pages_per_cycle"`
}

func UpdateSource(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Params("key")
		var req sourceUpdate
		if err := c.BodyParser(&req); err != nil {
			return badRequest(c, "invalid body")
		}
		updates := map[string]any{}
		if req.Enabled != nil {
			updates["enabled"] = *req.Enabled
		}
		if req.IntervalSec != nil {
			updates["interval_sec"] = *req.IntervalSec
		}
		if req.Concurrency != nil {
			updates["concurrency"] = *req.Concurrency
		}
		if req.MaxPagesPerCycle != nil {
			updates["max_pages_per_cycle"] = *req.MaxPagesPerCycle
		}
		if len(updates) == 0 {
			return badRequest(c, "no fields")
		}
		res := a.DB.Model(&db.Source{}).Where("key = ?", key).Updates(updates)
		if res.Error != nil {
			return serverError(c, res.Error)
		}
		if res.RowsAffected == 0 {
			return notFound(c, "source not found")
		}
		var row db.Source
		a.DB.Where("key = ?", key).First(&row)
		return ok(c, row, nil)
	}
}

func RunSourceNow(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Params("key")
		if err := a.Scheduler.RunSourceNow(c.Context(), key); err != nil {
			return badRequest(c, err.Error())
		}
		return ok(c, fiber.Map{"queued": true}, nil)
	}
}

var _ = strconv.Atoi
