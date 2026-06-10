package db

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type filterWriter struct{ w logger.Writer }

func (f filterWriter) Printf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if strings.Contains(msg, "record not found") {
		return
	}
	f.w.Printf("%s", msg)
}

func Open(path string, log *slog.Logger) (*gorm.DB, error) {
	gormLog := logger.New(
		filterWriter{w: slogWriter{log: log}},
		logger.Config{
			LogLevel:  logger.Warn,
			Colorful:  false,
			IgnoreRecordNotFoundError: true,
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

func (s slogWriter) Printf(format string, args ...any) {
	s.log.Info(strings.TrimSpace(fmt.Sprintf(format, args...)))
}
