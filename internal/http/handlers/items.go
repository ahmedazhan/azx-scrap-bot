package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
)

func ListItems(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		q := a.DB.Model(&db.Item{})

		if s := c.Query("source"); s != "" {
			q = q.Where("source_id = (SELECT id FROM sources WHERE key = ?)", s)
		}
		if t := c.Query("type"); t != "" {
			q = q.Where("type_slug = ?", t)
		}
		if flagged := c.Query("flagged"); flagged == "1" {
			q = q.Where("id IN (SELECT item_id FROM flag_hits)")
		}
		if search := strings.TrimSpace(c.Query("q")); search != "" {
			like := "%" + search + "%"
			q = q.Where("title LIKE ? OR body_text LIKE ? OR office LIKE ? OR reference_no LIKE ?", like, like, like, like)
		}
		if from := c.Query("from"); from != "" {
			if t, err := parseDate(from); err == nil {
				q = q.Where("published_at >= ?", t)
			}
		}
		if to := c.Query("to"); to != "" {
			if t, err := parseDate(to); err == nil {
				q = q.Where("published_at <= ?", t)
			}
		}

		page, _ := strconv.Atoi(c.Query("page", "1"))
		if page < 1 {
			page = 1
		}
		size, _ := strconv.Atoi(c.Query("page_size", "20"))
		if size < 1 || size > 200 {
			size = 20
		}
		var total int64
		q.Count(&total)

		var rows []db.Item
		if err := q.Order("published_at DESC NULLS LAST, fetched_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
			return serverError(c, err)
		}
		return ok(c, rows, fiber.Map{
			"page":      page,
			"page_size": size,
			"total":     total,
		})
	}
}

func GetItem(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {
			return badRequest(c, "invalid id")
		}
		var item db.Item
		if err := a.DB.First(&item, id).Error; err != nil {
			return notFound(c, "not found")
		}
		return ok(c, item, nil)
	}
}

func ItemFilters(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var types []string
		a.DB.Model(&db.Item{}).Distinct("type_slug").Pluck("type_slug", &types)

		var sources []string
		a.DB.Model(&db.Source{}).Where("enabled = ?", true).Pluck("key", &sources)

		return ok(c, fiber.Map{
			"types":   types,
			"sources": sources,
		}, nil)
	}
}

func parseDate(s string) (time.Time, error) {
	formats := []string{"2006-01-02", "2006-01-02T15:04:05Z"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, nil
}
