package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
)

func Dashboard(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var itemsTotal int64
		a.DB.Model(&db.Item{}).Count(&itemsTotal)

		startOfDay := time.Now().UTC().Truncate(24 * time.Hour)
		var itemsToday int64
		a.DB.Model(&db.Item{}).Where("created_at >= ?", startOfDay).Count(&itemsToday)

		var flaggedTotal int64
		a.DB.Model(&db.FlagHit{}).Count(&flaggedTotal)

		var sourcesOn int64
		a.DB.Model(&db.Source{}).Where("enabled = ?", true).Count(&sourcesOn)
		var sourcesOff int64
		a.DB.Model(&db.Source{}).Where("enabled = ?", false).Count(&sourcesOff)

		var recentRuns []db.ScrapeRun
		a.DB.Order("started_at DESC").Limit(20).Find(&recentRuns)

		type recentEntry struct {
			Time    time.Time `json:"time"`
			Source  string    `json:"source"`
			New     int       `json:"new"`
			Updated int       `json:"updated"`
			Errors  int       `json:"errors"`
		}
		recent := make([]recentEntry, 0, len(recentRuns))
		for _, r := range recentRuns {
			var key string
			var s db.Source
			if err := a.DB.First(&s, r.SourceID).Error; err == nil {
				key = s.Key
			}
			recent = append(recent, recentEntry{
				Time:    r.StartedAt,
				Source:  key,
				New:     r.NewItems,
				Updated: r.UpdatedItems,
				Errors:  r.Errors,
			})
		}

		type spark struct {
			Fetched []int `json:"fetched"`
			Flagged []int `json:"flagged"`
		}
		sparkMap := map[string]spark{}
		var sources []db.Source
		a.DB.Find(&sources)
		for _, s := range sources {
			cutoff := time.Now().UTC().Add(-14 * 24 * time.Hour)
			var runs []db.ScrapeRun
			a.DB.Where("source_id = ? AND started_at >= ?", s.ID, cutoff).Order("started_at ASC").Find(&runs)
			sp := spark{}
			for _, r := range runs {
				sp.Fetched = append(sp.Fetched, r.Fetched)
				var hits int64
				a.DB.Model(&db.FlagHit{}).
					Joins("JOIN items ON items.id = flag_hits.item_id").
					Where("items.source_id = ? AND flag_hits.created_at >= ?", s.ID, r.StartedAt).
					Count(&hits)
				sp.Flagged = append(sp.Flagged, int(hits))
			}
			sparkMap[s.Key] = sp
		}

		return ok(c, fiber.Map{
			"totals": fiber.Map{
				"items":       itemsTotal,
				"today":       itemsToday,
				"flagged":     flaggedTotal,
				"sources_on":  sourcesOn,
				"sources_off": sourcesOff,
			},
			"recent_activity": recent,
			"sparklines":      sparkMap,
		}, nil)
	}
}
