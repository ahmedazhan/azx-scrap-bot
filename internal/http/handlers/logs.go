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
				"seq":   e.Seq,
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

		ctx := c.UserContext()

		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			writeFlush := func() bool {
				return w.Flush() == nil
			}

			history := a.Ring.Snapshot()
			for _, e := range history {
				writeLogxEntry(w, e)
			}
			if !writeFlush() {
				return
			}
			var lastSeq uint64
			if n := len(history); n > 0 {
				lastSeq = history[n-1].Seq
			}

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if _, err := fmt.Fprintf(w, ":heartbeat\n\n"); err != nil {
						return
					}
					if !writeFlush() {
						return
					}
					entries := a.Ring.Snapshot()
					emitted := false
					for _, e := range entries {
						if e.Seq > lastSeq {
							writeLogxEntry(w, e)
							lastSeq = e.Seq
							emitted = true
						}
					}
					if !emitted {
						continue
					}
					if !writeFlush() {
						return
					}
				}
			}
		}))
		return nil
	}
}

func writeLogxEntry(w *bufio.Writer, e logx.Entry) {
	doc := fiber.Map{
		"seq":   e.Seq,
		"time":  e.Time.UTC().Format(time.RFC3339Nano),
		"level": e.Level.String(),
		"msg":   e.Msg,
		"attrs": e.Attrs,
	}
	b, _ := json.Marshal(doc)
	fmt.Fprintf(w, "data: %s\n\n", b)
}
