package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ahmedazhan/azx-scrap-bot/internal/app"
	"github.com/ahmedazhan/azx-scrap-bot/internal/auth"
	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
)

const (
	accessTTL  = 7 * 24 * time.Hour
	refreshTTL = 30 * 24 * time.Hour
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type setupReq struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	SetupToken string `json:"setup_token"`
}

type refreshReq struct {
	Refresh string `json:"refresh"`
}

type pwChangeReq struct {
	Old string `json:"old"`
	New string `json:"new"`
}

func Health(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return ok(c, fiber.Map{
			"status":  "ok",
			"version": a.Version,
		}, nil)
	}
}

func SetupInfo(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var count int64
		if err := a.DB.Model(&db.User{}).Count(&count).Error; err != nil {
			return serverError(c, err)
		}
		tokenSource := "db"
		if a.Cfg.SetupToken != "" {
			tokenSource = "env"
		} else {
			var meta db.AppMeta
			if err := a.DB.Where("key = ?", auth.MetaSetupToken).First(&meta).Error; err != nil {
				tokenSource = "none"
			}
		}
		return ok(c, fiber.Map{
			"setup_required": count == 0,
			"user_count":     count,
			"token_source":   tokenSource,
			"env_token_set":  a.Cfg.SetupToken != "",
			"env_admin_set":  a.Cfg.AdminUser != "" && a.Cfg.AdminPassword != "",
		}, nil)
	}
}

func Login(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req loginReq
		if err := c.BodyParser(&req); err != nil {
			return badRequest(c, "invalid body")
		}
		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || req.Password == "" {
			return badRequest(c, "username and password required")
		}
		var user db.User
		if err := a.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
			return unauthorized(c, "invalid credentials")
		}
		if !auth.VerifyPassword(user.PasswordHash, req.Password) {
			return unauthorized(c, "invalid credentials")
		}
		return issueTokens(c, a, user.ID)
	}
}

func Refresh(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req refreshReq
		if err := c.BodyParser(&req); err != nil || req.Refresh == "" {
			return badRequest(c, "missing refresh")
		}
		uid, err := a.Auth.Parse(req.Refresh)
		if err != nil {
			return unauthorized(c, "invalid refresh")
		}
		return issueTokens(c, a, uid)
	}
}

func Setup(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req setupReq
		if err := c.BodyParser(&req); err != nil {
			return badRequest(c, "invalid body")
		}
		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || len(req.Password) < 8 || req.SetupToken == "" {
			return badRequest(c, "username, password (>=8 chars), and setup_token required")
		}
		if a.Cfg.SetupToken != "" {
			if req.SetupToken != a.Cfg.SetupToken {
				return unauthorized(c, "invalid setup token")
			}
		} else {
			var meta db.AppMeta
			if err := a.DB.Where("key = ?", auth.MetaSetupToken).First(&meta).Error; err != nil {
				return badRequest(c, "no setup in progress")
			}
			if meta.Value != req.SetupToken {
				return unauthorized(c, "invalid setup token")
			}
		}
		var count int64
		a.DB.Model(&db.User{}).Count(&count)
		if count > 0 {
			return badRequest(c, "user already exists")
		}
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			return badRequest(c, "hash failed")
		}
		user := db.User{Username: req.Username, PasswordHash: hash, Theme: "dark", FilterMode: "sheet"}
		if err := a.DB.Create(&user).Error; err != nil {
			return badRequest(c, "could not create user")
		}
		if a.Cfg.SetupToken == "" {
			if err := auth.ConsumeSetupToken(a.DB); err != nil {
				a.Log.Warn("consume setup token", "err", err)
			}
		} else {
			a.Log.Info("setup complete (env-driven token used; token will remain active for future restarts)")
		}
		return issueTokens(c, a, user.ID)
	}
}

func Logout(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return ok(c, fiber.Map{"ok": true}, nil)
	}
}

func Me(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid := c.Locals("user_id").(uint64)
		var user db.User
		if err := a.DB.First(&user, uid).Error; err != nil {
			return badRequest(c, "user not found")
		}
		return ok(c, fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"theme":       user.Theme,
			"filter_mode": user.FilterMode,
		}, nil)
	}
}

func ChangePassword(a *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid := c.Locals("user_id").(uint64)
		var req pwChangeReq
		if err := c.BodyParser(&req); err != nil || len(req.New) < 8 {
			return badRequest(c, "new password must be >=8 chars")
		}
		var user db.User
		if err := a.DB.First(&user, uid).Error; err != nil {
			return badRequest(c, "user not found")
		}
		if !auth.VerifyPassword(user.PasswordHash, req.Old) {
			return unauthorized(c, "old password mismatch")
		}
		hash, err := auth.HashPassword(req.New)
		if err != nil {
			return badRequest(c, "hash failed")
		}
		user.PasswordHash = hash
		a.DB.Save(&user)
		return ok(c, fiber.Map{"ok": true}, nil)
	}
}

func issueTokens(c *fiber.Ctx, a *app.App, uid uint64) error {
	access, err := a.Auth.Issue(uid, accessTTL)
	if err != nil {
		return badRequest(c, "token issue failed")
	}
	refresh, err := a.Auth.Issue(uid, refreshTTL)
	if err != nil {
		return badRequest(c, "token issue failed")
	}
	return ok(c, fiber.Map{
		"access":  access,
		"refresh": refresh,
		"user": fiber.Map{
			"id": uid,
		},
	}, nil)
}

func ok(c *fiber.Ctx, data any, meta any) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta":    meta,
	})
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success": false,
		"error":   msg,
	})
}

func unauthorized(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"success": false,
		"error":   msg,
	})
}

func notFound(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"success": false,
		"error":   msg,
	})
}

func serverError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}

var _ = gorm.ErrRecordNotFound
