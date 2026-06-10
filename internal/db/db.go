package db

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(path string, log *slog.Logger) (*gorm.DB, error) {
	gormLog := logger.New(
		slogWriter{log: log},
		logger.Config{
			LogLevel: logger.Warn,
			Colorful: false,
		},
	)
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)"), &gorm.Config{
		Logger:                                   gormLog,
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&User{}, &Source{}, &Item{}, &FlagRule{}, &FlagHit{},
		&TelegramConfig{}, &TelegramSubscriber{}, &ScrapeRun{},
		&AppMeta{}, &UIPref{},
	); err != nil {
		return nil, err
	}
	return db, nil
}

type slogWriter struct{ log *slog.Logger }

func (s slogWriter) Write(p []byte) (int, error) {
	s.log.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}

func (s slogWriter) Printf(format string, args ...any) {
	s.log.Info(strings.TrimSpace(fmt.Sprintf(format, args...)))
}
