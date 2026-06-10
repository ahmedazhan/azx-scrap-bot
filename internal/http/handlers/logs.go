package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
	"github.com/ahmedazhan/azx-scrap-bot/internal/logx"
)

func LogsRecent(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		n, _ := strconv.Atoi(c.Query("limit", "200"))
		if n <= 0 || n > 500 {
			n = 200
		}
		entries := a.Ring.Recent(n)
		out := make([]fiber.Map, 0, len(entries))
		for _, e := range entries {
			out = append(out, fiber.Map{
				"time":  e.Time.UTC().Format(time.RFC3339Nano),
				"level": e.Level.String(),
				"msg":   e.Msg,
				"attrs": e.Attrs,
			})
		}
		return ok(c, out, nil)
	}
}

func LogsStream(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("X-Accel-Buffering", "no")

		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			writeFlush := func() bool {
				return w.Flush() == nil
			}

			last := a.Ring.Snapshot()
			for _, e := range last {
				writeLogxEntry(w, e)
			}
			if !writeFlush() {
				return
			}
			idx := len(last)

			for {
				select {
				case <-c.Context().Done():
					return
				case <-ticker.C:
					if _, err := fmt.Fprintf(w, ":heartbeat\n\n"); err != nil {
						return
					}
					if !writeFlush() {
						return
					}
					entries := a.Ring.Snapshot()
					if len(entries) > idx {
						for i := idx; i < len(entries); i++ {
							writeLogxEntry(w, entries[i])
						}
					} else if idx > 0 && len(entries) < idx {
						for _, e := range entries {
							writeLogxEntry(w, e)
						}
					}
					if !writeFlush() {
						return
					}
					idx = len(entries)
				}
			}
		}))
		return nil
	}
}

func writeLogxEntry(w *bufio.Writer, e logx.Entry) {
	doc := fiber.Map{
		"time":  e.Time.UTC().Format(time.RFC3339Nano),
		"level": e.Level.String(),
		"msg":   e.Msg,
		"attrs": e.Attrs,
	}
	b, _ := json.Marshal(doc)
	fmt.Fprintf(w, "data: %s\n\n", b)
}
