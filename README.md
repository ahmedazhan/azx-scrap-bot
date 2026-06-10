# Azx Scrap Bot

© azhan. All rights reserved. UNLICENSED.

A single-binary multi-site scraper with Telegram alerts, in Go + Vue 3.

## Quick start

    make env      # one-time: copy .env.example to .env (Go + UI)
    make dist     # build UI + Go binary, UPX-compress
    ./azx-scrap-bot

Open http://localhost:8080 and follow the first-run setup.

## Development

    make dev          # foreground: Go (:8080) + Vite (:5173) in parallel
    make dev-bg       # detached, survives the spawning shell
    make dev-stop     # kills both

Open http://localhost:5173 — Vite proxies `/api/*` to the Go server.

## Configuration

Config is loaded in this order (later overrides earlier):

1. Built-in defaults
2. `.env` file (Go side) and `ui/.env` (UI side) — created by `make env`
3. Environment variables
4. CLI flags (Go side only)

### Go server flags

    --addr           listen address          (default :8080)
    --db             sqlite path             (default ./store.db)
    --log-dir        log directory           (default ./logs)
    --log-level      debug|info|warn|error   (default debug)
    --log-max-size   MB per file             (default 50)
    --log-max-backups  rotated files to keep (default 5)
    --log-compress   gzip old logs           (default true)
    --assets-dir     SPA dir (dev only; falls back to embedded)
    --jwt-secret     HMAC secret for JWTs (overrides DB-stored one)

Equivalent env vars: `ADDR`, `DB`, `LOG_DIR`, `LOG_LEVEL`, `LOG_MAX_SIZE`,
`LOG_MAX_BACKUPS`, `LOG_COMPRESS`, `ASSETS_DIR`, `JWT_SECRET`.

### UI env (`ui/.env`)

    VITE_API_BASE_URL=/api          # proxied to Go in dev; or absolute URL
    VITE_APP_VERSION=0.1.0          # shown in the Account page

### JWT secret

If you set `JWT_SECRET` in `.env` (or as an env var), the server uses it
instead of the auto-generated one stored in the DB. **Changing it
invalidates all existing browser sessions** — users will have to log in
again. Generate one with:

    openssl rand -hex 32
