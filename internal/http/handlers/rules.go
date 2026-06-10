package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
	"github.com/ahmedazhan/azx-scrap-bot/internal/scraper"
)

type rulePayload struct {
	Name          string `json:"name"`
	Pattern       string `json:"pattern"`
	IsRegex       bool   `json:"is_regex"`
	MatchIn       string `json:"match_in"`
	CaseSensitive bool   `json:"case_sensitive"`
	Enabled       bool   `json:"enabled"`
}

func ListRules(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var rows []db.FlagRule
		if err := a.DB.Order("id ASC").Find(&rows).Error; err != nil {
			return serverError(c, err)
		}
		return ok(c, rows, nil)
	}
}

func CreateRule(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var p rulePayload
		if err := c.BodyParser(&p); err != nil {
			return badRequest(c, "invalid body")
		}
		if p.Pattern == "" || p.Name == "" {
			return badRequest(c, "name and pattern required")
		}
		if p.MatchIn == "" {
			p.MatchIn = "body"
		}
		row := db.FlagRule{
			Name:          p.Name,
			Pattern:       p.Pattern,
			IsRegex:       p.IsRegex,
			MatchIn:       p.MatchIn,
			CaseSensitive: p.CaseSensitive,
			Enabled:       p.Enabled,
		}
		if err := a.DB.Create(&row).Error; err != nil {
			return serverError(c, err)
		}
		return ok(c, row, nil)
	}
}

func UpdateRule(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {
			return badRequest(c, "invalid id")
		}
		var p rulePayload
		if err := c.BodyParser(&p); err != nil {
			return badRequest(c, "invalid body")
		}
		var row db.FlagRule
		if err := a.DB.First(&row, id).Error; err != nil {
			return notFound(c, "not found")
		}
		if p.Name != "" {
			row.Name = p.Name
		}
		if p.Pattern != "" {
			row.Pattern = p.Pattern
		}
		row.IsRegex = p.IsRegex
		if p.MatchIn != "" {
			row.MatchIn = p.MatchIn
		}
		row.CaseSensitive = p.CaseSensitive
		row.Enabled = p.Enabled
		if err := a.DB.Save(&row).Error; err != nil {
			return serverError(c, err)
		}
		return ok(c, row, nil)
	}
}

func DeleteRule(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {
			return badRequest(c, "invalid id")
		}
		if err := a.DB.Delete(&db.FlagRule{}, id).Error; err != nil {
			return serverError(c, err)
		}
		return ok(c, fiber.Map{"ok": true}, nil)
	}
}

type testReq struct {
	Pattern       string   `json:"pattern"`
	IsRegex       bool     `json:"is_regex"`
	MatchIn       string   `json:"match_in"`
	CaseSensitive bool     `json:"case_sensitive"`
	ItemIDs       []uint64 `json:"item_ids"`
}

func TestRule(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req testReq
		if err := c.BodyParser(&req); err != nil {
			return badRequest(c, "invalid body")
		}
		if req.Pattern == "" {
			return badRequest(c, "pattern required")
		}
		matches, samples := a.Flagger.TestRule(req.Pattern, req.IsRegex, req.CaseSensitive, req.MatchIn, req.ItemIDs)
		return ok(c, fiber.Map{
			"matches": matches,
			"sample":  samples,
		}, nil)
	}
}

var _ = scraper.Pool{}
