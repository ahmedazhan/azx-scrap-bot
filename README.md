# Azx Scrap Bot

© azhan. All rights reserved. UNLICENSED.

A single-binary multi-site scraper with Telegram alerts, in Go + Vue 3.

## Quick start

    make dist
    ./azx-scrap-bot

Open http://localhost:8080 and follow the first-run setup.

## Flags

    --addr           listen address (default :8080)
    --db             sqlite path    (default ./store.db)
    --log-dir        log directory  (default ./logs)
    --log-level      debug|info|warn|error (default debug)
    --log-max-size   MB per file    (default 50)
    --log-max-backups  rotations    (default 5)
    --log-compress   gzip old logs  (default true)

Equivalent env vars: ADDR, DB, LOG_DIR, LOG_LEVEL, LOG_MAX_SIZE, LOG_MAX_BACKUPS, LOG_COMPRESS.
