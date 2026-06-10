APP       := azx-scrap-bot
PKG       := ./cmd/azx-scrap-bot
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -s -w -X main.version=$(VERSION)
GO        ?= CGO_ENABLED=0 GOOS=linux go
GODEV     ?= go
PKG_MGR   ?= pnpm

# Add local bin to PATH so pnpm-installed CLIs (e.g. vite) are found.
export PATH := $(CURDIR)/ui/node_modules/.bin:$(PATH)

.PHONY: all ui build dist run test lint clean tidy dev dev-ui dev-go dev-stop

all: ui build

# --- UI -----------------------------------------------------------------------

deps:
	cd ui && $(PKG_MGR) install --frozen-lockfile

ui: deps
	cd ui && $(PKG_MGR) run build
	mkdir -p internal/assets/ui
	rm -rf internal/assets/ui/*
	cp -R ui/dist/* internal/assets/ui/

# Vite dev server only. Proxies /api -> http://localhost:8080.
dev-ui:
	cd ui && $(PKG_MGR) run dev

# --- Go -----------------------------------------------------------------------

build:
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(APP) $(PKG)

# Run the Go binary with the SPA served from disk via ASSETS_DIR (no rebuild on UI change).
# Uses the host Go toolchain (no GOOS=linux) so `go run` works on macOS/Linux dev boxes.
dev-go:
	ASSETS_DIR=$(CURDIR)/ui/dist $(GODEV) run ./cmd/azx-scrap-bot

# Both at once: Go server on :8080, Vite on :5173. Vite proxies /api -> :8080.
# Open http://localhost:5173 for the UI; the Go API is hit directly.
dev:
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) -j2 dev-go dev-ui

# --- Ship ---------------------------------------------------------------------

dist: ui build
	@command -v upx >/dev/null 2>&1 || { echo "upx not installed; skipping compression"; ls -lh $(APP); exit 0; }
	upx --best --lzma $(APP) || true
	@ls -lh $(APP)

run: build
	./$(APP)

test:
	$(GO) test ./...

tidy:
	$(GO) mod tidy

clean:
	rm -f $(APP)
	rm -rf internal/assets/ui/*

lint:
	golangci-lint run ./...
