package auth

import (
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
)

const (
	MetaSetupToken  = "setup_token"
	MetaJWTSecret   = "jwt_secret"
	MetaSetupUsedAt = "setup_used_at"
)

func EnsureSetupToken(gdb *gorm.DB, log *slog.Logger) (string, bool, error) {
	var meta db.AppMeta
	err := gdb.Where("key = ?", MetaSetupToken).First(&meta).Error
	if err == nil {
		return meta.Value, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, err
	}
	tok := uuid.NewString()
	meta = db.AppMeta{Key: MetaSetupToken, Value: tok}
	if err := gdb.Create(&meta).Error; err != nil {
		return "", false, err
	}
	log.Info("generated setup token")
	return tok, true, nil
}

func ConsumeSetupToken(gdb *gorm.DB) error {
	return gdb.Where("key = ?", MetaSetupToken).Delete(&db.AppMeta{}).Error
}

func EnsureJWTSecret(gdb *gorm.DB) (string, error) {
	var meta db.AppMeta
	err := gdb.Where("key = ?", MetaJWTSecret).First(&meta).Error
	if err == nil {
		return meta.Value, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	secret := RandomSecret()
	if err := gdb.Create(&db.AppMeta{Key: MetaJWTSecret, Value: secret}).Error; err != nil {
		return "", err
	}
	return secret, nil
}
