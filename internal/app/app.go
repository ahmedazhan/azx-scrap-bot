package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ahmedazhan/azx-scrap-bot/internal/auth"
	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
	"github.com/ahmedazhan/azx-scrap-bot/internal/logx"
	"github.com/ahmedazhan/azx-scrap-bot/internal/scraper"
	"github.com/ahmedazhan/azx-scrap-bot/internal/telegram"
)

type App struct {
	Cfg       *Config
	Log       *slog.Logger
	DB        *gorm.DB
	Scheduler *scraper.Scheduler
	Telegram  *telegram.Dispatcher
	Flagger   *scraper.Flagger
	Ring      *logx.RingBuffer
	HTTP      *fiber.App
	Version   string
	Auth      *auth.Service
	Logger    *logx.Logger
	SPAHTML   string
}

func New(version string, cfg *Config) (*App, error) {
	lvl := parseLevel(cfg.LogLevel)
	lx, err := logx.New(logx.Options{
		Dir:        cfg.LogDir,
		Level:      lvl,
		MaxSizeMB:  cfg.LogMaxSize,
		MaxBackups: cfg.LogMaxBackups,
		Compress:   cfg.LogCompress,
		RingSize:   500,
		BufferSize: 64 * 1024,
	})
	if err != nil {
		return nil, fmt.Errorf("logx: %w", err)
	}

	database, err := db.Open(cfg.DB, lx.Log)
	if err != nil {
		lx.Shutdown(context.Background())
		return nil, fmt.Errorf("db: %w", err)
	}

	if err := db.Seed(database, lx.Log); err != nil {
		lx.Shutdown(context.Background())
		return nil, fmt.Errorf("seed: %w", err)
	}

	secret, err := auth.EnsureJWTSecret(database, cfg.JWTSecret)
	if err != nil {
		lx.Shutdown(context.Background())
		return nil, fmt.Errorf("jwt secret: %w", err)
	}
	authSvc := auth.NewService(secret)

	dispatcher := telegram.NewDispatcher(database, lx.Log)
	flagger := scraper.NewFlagger(database, lx.Log, dispatcher.Notify)
	scheduler := scraper.New(database, lx.Log, flagger)

	app := &App{
		Cfg:       cfg,
		Log:       lx.Log,
		DB:        database,
		Scheduler: scheduler,
		Telegram:  dispatcher,
		Flagger:   flagger,
		Ring:      lx.Ring,
		Version:   version,
		Auth:      authSvc,
		Logger:    lx,
	}
	return app, nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}
	return slog.LevelInfo
}

func (a *App) Run(ctx context.Context) error {
	if err := a.Flagger.Start(ctx); err != nil {
		return err
	}
	if err := a.Telegram.Start(ctx); err != nil {
		return err
	}
	if err := a.Scheduler.Start(ctx); err != nil {
		return err
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.Scheduler != nil {
		a.Scheduler.Stop(ctx)
	}
	if a.Flagger != nil {
		a.Flagger.Stop()
	}
	if a.Telegram != nil {
		a.Telegram.Stop()
	}
	if a.HTTP != nil {
		_ = a.HTTP.ShutdownWithContext(ctx)
	}
	if a.DB != nil {
		sqlDB, err := a.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
	if a.Logger != nil {
		return a.Logger.Shutdown(ctx)
	}
	return nil
}
