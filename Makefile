# veracode-go-cli — Makefile de apoio (build local versionado)

MODULE := github.com/appsecomega/veracode-go-cli
CMD := ./cmd/veracode-go-cli
BIN := bin/veracode-go-cli

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.Date=$(DATE)

.PHONY: build install test vet fmt snapshot clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD)

install:
	go install -ldflags "$(LDFLAGS)" $(CMD)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

# Gera binários locais sem publicar (requer goreleaser instalado).
snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -rf bin dist
