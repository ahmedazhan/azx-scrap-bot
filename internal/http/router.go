package http

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
	"github.com/ahmedazhan/azx-scrap-bot/internal/http/handlers"
)

func NewRouter(a *app.App) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "azx-scrap-bot",
		DisableStartupMessage: true,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
		ErrorHandler:          errorHandler(a.Log),
	})

	app.Use(middlewareRecover())
	app.Use(middlewareRequestID())
	app.Use(middlewareRequestLog(a.Log))
	app.Use(middlewareCORS())

	api := app.Group("/api")

	api.Get("/health", handlers.Health(a))
	api.Get("/setup-info", handlers.SetupInfo(a))
	api.Get("/_echo", handlers.Echo(a))

	authGroup := api.Group("/auth")
	authGroup.Post("/login", handlers.Login(a))
	authGroup.Post("/refresh", handlers.Refresh(a))
	authGroup.Post("/setup", handlers.Setup(a))

	api.Use(middlewareJWT(a))
	api.Get("/auth/me", handlers.Me(a))
	api.Post("/auth/logout", handlers.Logout(a))
	api.Post("/auth/change-password", handlers.ChangePassword(a))

	api.Get("/sources", handlers.ListSources(a))
	api.Get("/sources/:key", handlers.GetSource(a))
	api.Put("/sources/:key", handlers.UpdateSource(a))
	api.Post("/sources/:key/run-now", handlers.RunSourceNow(a))

	api.Get("/items", handlers.ListItems(a))
	api.Get("/items/filters", handlers.ItemFilters(a))
	api.Get("/items/:id", handlers.GetItem(a))

	api.Get("/rules", handlers.ListRules(a))
	api.Post("/rules", handlers.CreateRule(a))
	api.Put("/rules/:id", handlers.UpdateRule(a))
	api.Delete("/rules/:id", handlers.DeleteRule(a))
	api.Post("/rules/test", handlers.TestRule(a))

	api.Get("/telegram/config", handlers.GetTelegramConfig(a))
	api.Put("/telegram/config", handlers.UpdateTelegramConfig(a))
	api.Post("/telegram/test", handlers.TestTelegram(a))
	api.Get("/telegram/subscribers", handlers.ListSubscribers(a))
	api.Post("/telegram/subscribers", handlers.CreateSubscriber(a))
	api.Put("/telegram/subscribers/:id", handlers.UpdateSubscriber(a))
	api.Delete("/telegram/subscribers/:id", handlers.DeleteSubscriber(a))

	api.Get("/dashboard", handlers.Dashboard(a))
	api.Get("/account", handlers.GetAccount(a))
	api.Put("/account", handlers.UpdateAccount(a))

	api.Get("/logs/recent", handlers.LogsRecent(a))
	api.Get("/logs/stream", handlers.LogsStream(a))

	app.Get("/*", spaHandler(a))

	return app
}

func errorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}
		if code >= 500 {
			log.Error("request error", "err", err.Error(), "path", c.Path())
		}
		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
}
