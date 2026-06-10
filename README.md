# Azx Scrap Bot

© azhan. All rights reserved. UNLICENSED.

A single-binary multi-site scraper with Telegram alerts, in Go + Vue 3.

## Quick start

    make env      # one-time: copy .env.example to .env (Go + UI)
    make dist     # build UI + Go binary, UPX-compress
    ./azx-scrap-bot

The dev `.env` ships with an **env-driven admin** (`azhan` / `changeme123`)
and a **fixed setup token** (`dev-setup-token-12345`). The Go binary will
auto-create the admin user on first boot and print:

```
============================================================
 ADMIN (from env)
   username: azhan
   password: changeme123
   sign in:  http://localhost:8088/login
============================================================
```

Just open `http://localhost:8088/login` (or `http://localhost:5174/login`
in dev) and sign in.

## Development

    make dev          # foreground: Go (:8088) + Vite (:5174) in parallel
    make dev-bg       # detached, survives the spawning shell
    make dev-stop     # kills both

Open http://localhost:5174 — Vite proxies `/api/*` to the Go server.

## Configuration

Config is loaded in this order (later overrides earlier):

1. Built-in defaults
2. `.env` file (Go side) and `ui/.env` (UI side) — created by `make env`
3. Environment variables
4. CLI flags (Go side only)

### Go server flags

    --addr             listen address           (default :8088)
    --db               sqlite path              (default ./store.db)
    --log-dir          log directory            (default ./logs)
    --log-level        debug|info|warn|error    (default debug)
    --log-max-size     MB per file              (default 50)
    --log-max-backups  rotated files to keep    (default 5)
    --log-compress     gzip old logs            (default true)
    --assets-dir       SPA dir (dev only; falls back to embedded)
    --jwt-secret       HMAC secret for JWTs
    --setup-token      fixed setup token (recommended for env-driven dev)
    --admin-user       auto-create this admin on boot (requires --admin-password)
    --admin-password   password for the auto-created admin (>= 8 chars)

Equivalent env vars: `ADDR`, `DB`, `LOG_DIR`, `LOG_LEVEL`, `LOG_MAX_SIZE`,
`LOG_MAX_BACKUPS`, `LOG_COMPRESS`, `ASSETS_DIR`, `JWT_SECRET`,
`AZX_SETUP_TOKEN`, `AZX_ADMIN_USERNAME`, `AZX_ADMIN_PASSWORD`.

### UI env (`ui/.env`)

    VITE_API_BASE_URL=/api
    VITE_APP_VERSION=0.1.0
    VITE_SETUP_TOKEN=                  # mirrors AZX_SETUP_TOKEN; pre-fills the setup form
    VITE_ADMIN_USERNAME=               # mirrors AZX_ADMIN_USERNAME
    VITE_ADMIN_PASSWORD=               # mirrors AZX_ADMIN_PASSWORD

## Env-driven admin (no setup form needed)

If you set `AZX_ADMIN_USERNAME` and `AZX_ADMIN_PASSWORD` (or pass
`--admin-user` / `--admin-password`), the server will create that user
on first boot. The setup form is then unnecessary; just sign in at
`/login`. The dev `.env` ships with this pre-configured.

To rotate the admin password later, use the Account page in the UI or
`POST /api/auth/change-password`.

## Setup token

The setup token protects the `POST /api/auth/setup` endpoint. Two modes:

- **Env-driven (recommended)**: set `AZX_SETUP_TOKEN` in `.env`. The same
  value is used every time, and you can put it in `ui/.env` as
  `VITE_SETUP_TOKEN` to auto-fill the setup form.
- **DB-generated (default)**: a UUID is generated on first boot and
  stored in `app_meta`. It's printed to the console. To get a new one,
  delete the `app_meta` row where `key = 'setup_token'`.

## Ports

Dev uses non-default ports to avoid clashing with common local apps:

- **Go API**: `:8088`
- **Vite UI**: `:5174`

## JWT secret

If you set `JWT_SECRET` in `.env` (or as an env var), the server uses it
instead of the auto-generated one stored in the DB. **Changing it
invalidates all existing browser sessions**. Generate one with:

    openssl rand -hex 32
