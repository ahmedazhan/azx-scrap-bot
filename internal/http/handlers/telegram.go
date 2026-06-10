package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
	"github.com/ahmedazhan/azx-scrap-bot/internal/telegram"
)

type tgConfig struct {
	BotToken         string `json:"bot_token"`
	Enabled          bool   `json:"enabled"`
	NotifyOnFlagOnly bool   `json:"notify_on_flag_only"`
	ThrottleMs       int    `json:"throttle_ms"`
}

func GetTelegramConfig(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var cfg db.TelegramConfig
		if err := a.DB.First(&cfg, 1).Error; err != nil {
			cfg = db.TelegramConfig{ID: 1}
		}
		masked := ""
		if cfg.BotToken != "" {
			masked = "***set***"
		}
		return ok(c, fiber.Map{
			"bot_token_masked":   masked,
			"enabled":            cfg.Enabled,
			"notify_on_flag_only": cfg.NotifyOnFlagOnly,
			"throttle_ms":        cfg.ThrottleMs,
		}, nil)
	}
}

func UpdateTelegramConfig(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req tgConfig
		if err := c.BodyParser(&req); err != nil {
			return badRequest(c, "invalid body")
		}
		var cfg db.TelegramConfig
		if err := a.DB.First(&cfg, 1).Error; err != nil {
			cfg = db.TelegramConfig{ID: 1, ThrottleMs: 250, NotifyOnFlagOnly: true}
		}
		if req.BotToken != "" && req.BotToken != "***set***" {
			cfg.BotToken = req.BotToken
		}
		cfg.Enabled = req.Enabled
		cfg.NotifyOnFlagOnly = req.NotifyOnFlagOnly
		if req.ThrottleMs > 0 {
			cfg.ThrottleMs = req.ThrottleMs
		}
		a.DB.Save(&cfg)
		masked := ""
		if cfg.BotToken != "" {
			masked = "***set***"
		}
		return ok(c, fiber.Map{
			"bot_token_masked":    masked,
			"enabled":             cfg.Enabled,
			"notify_on_flag_only": cfg.NotifyOnFlagOnly,
			"throttle_ms":         cfg.ThrottleMs,
		}, nil)
	}
}

type tgTestReq struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func TestTelegram(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req tgTestReq
		if err := c.BodyParser(&req); err != nil {
			return badRequest(c, "invalid body")
		}
		if req.ChatID == 0 || req.Text == "" {
			return badRequest(c, "chat_id and text required")
		}
		var cfg db.TelegramConfig
		if err := a.DB.First(&cfg, 1).Error; err != nil || cfg.BotToken == "" {
			return badRequest(c, "bot not configured")
		}
		bot := telegram.NewBot(cfg.BotToken, time.Duration(cfg.ThrottleMs)*time.Millisecond)
		ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
		defer cancel()
		if err := bot.Send(ctx, req.ChatID, req.Text); err != nil {
			return badRequest(c, err.Error())
		}
		return ok(c, fiber.Map{"ok": true}, nil)
	}
}

func ListSubscribers(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var rows []db.TelegramSubscriber
		if err := a.DB.Order("id ASC").Find(&rows).Error; err != nil {
			return serverError(c, err)
		}
		return ok(c, rows, nil)
	}
}

type subPayload struct {
	ChatID  int64  `json:"chat_id"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

func CreateSubscriber(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var p subPayload
		if err := c.BodyParser(&p); err != nil {
			return badRequest(c, "invalid body")
		}
		if p.ChatID == 0 {
			return badRequest(c, "chat_id required")
		}
		row := db.TelegramSubscriber{ChatID: p.ChatID, Label: p.Label, Enabled: p.Enabled}
		if err := a.DB.Create(&row).Error; err != nil {
			return serverError(c, err)
		}
		return ok(c, row, nil)
	}
}

func UpdateSubscriber(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {
			return badRequest(c, "invalid id")
		}
		var p subPayload
		if err := c.BodyParser(&p); err != nil {
			return badRequest(c, "invalid body")
		}
		var row db.TelegramSubscriber
		if err := a.DB.First(&row, id).Error; err != nil {
			return notFound(c, "not found")
		}
		if p.Label != "" {
			row.Label = p.Label
		}
		row.Enabled = p.Enabled
		if err := a.DB.Save(&row).Error; err != nil {
			return serverError(c, err)
		}
		return ok(c, row, nil)
	}
}

func DeleteSubscriber(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {
			return badRequest(c, "invalid id")
		}
		if err := a.DB.Delete(&db.TelegramSubscriber{}, id).Error; err != nil {
			return serverError(c, err)
		}
		return ok(c, fiber.Map{"ok": true}, nil)
	}
}
