# Makefile — MJAYNRI
# Usage : make <target>
# Toutes les commandes supposent Go 1.22+ installé.

BINARY   := mjaynri
CMD_PATH := ./cmd/mjaynri
PORT     ?= 8080

.PHONY: all build test lint run clean coverage help

## all : Build par défaut
all: build

## build : Compile le binaire dans ./bin/
build:
	@mkdir -p bin
	go build -o bin/$(BINARY) $(CMD_PATH)
	@echo "✓ Binaire : bin/$(BINARY)"

## test : Lance tous les tests avec le race detector
test:
	go test -race -count=1 -timeout=60s ./...

## coverage : Génère un rapport de couverture HTML
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Rapport : coverage.html"

## lint : Vérifie le formatage et lance go vet
lint:
	@echo "→ gofmt"
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then echo "Non formatés :\n$$UNFORMATTED"; exit 1; fi
	@echo "→ go vet"
	go vet ./...
	@echo "✓ Lint OK"

## run : Lance l'application en mode développement
run: build
	MJAYNRI_PORT=$(PORT) ./bin/$(BINARY)

## clean : Supprime les artefacts de build
clean:
	rm -rf bin/ coverage.out coverage.html

## help : Affiche cette aide
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
