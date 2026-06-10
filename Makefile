APP       := azx-scrap-bot
PKG       := ./cmd/azx-scrap-bot
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -s -w -X main.version=$(VERSION)
GO        ?= CGO_ENABLED=0 GOOS=linux go

.PHONY: all ui build dist run test lint clean tidy

all: ui build

ui:
	cd ui && npm ci && npm run build
	mkdir -p internal/assets/ui
	rm -rf internal/assets/ui/*
	cp -R ui/dist/* internal/assets/ui/

build:
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(APP) $(PKG)

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
