package app

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr         string
	DB           string
	LogDir       string
	LogLevel     string
	LogMaxSize   int
	LogMaxBackups int
	LogCompress  bool
}

func envOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envIntOr(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envBoolOr(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return def
}

func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("azx-scrap-bot", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "listen address")
	db := fs.String("db", "./store.db", "sqlite path")
	logDir := fs.String("log-dir", "./logs", "log directory")
	logLevel := fs.String("log-level", "debug", "debug|info|warn|error")
	logMax := fs.Int("log-max-size", 50, "MB per file")
	logBackups := fs.Int("log-max-backups", 5, "rotated files to keep")
	logCompress := fs.Bool("log-compress", true, "gzip old logs")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg := &Config{
		Addr:          envOr("ADDR", *addr),
		DB:            envOr("DB", *db),
		LogDir:        envOr("LOG_DIR", *logDir),
		LogLevel:      envOr("LOG_LEVEL", *logLevel),
		LogMaxSize:    envIntOr("LOG_MAX_SIZE", *logMax),
		LogMaxBackups: envIntOr("LOG_MAX_BACKUPS", *logBackups),
		LogCompress:   envBoolOr("LOG_COMPRESS", *logCompress),
	}

	switch strings.ToLower(cfg.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		return nil, fmt.Errorf("invalid log level %q", cfg.LogLevel)
	}
	if cfg.LogMaxSize <= 0 {
		return nil, fmt.Errorf("log-max-size must be > 0")
	}
	return cfg, nil
}
