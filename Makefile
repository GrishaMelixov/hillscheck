.PHONY: dev build test lint migrate-up migrate-down docker-up docker-down frontend

# Local development with hot reload
dev:
	docker compose -f docker-compose.dev.yml up --build

# Build production binary (requires web/dist to exist)
build:
	cd web && npm run build
	go build -o bin/hillscheck ./cmd/server

# Run tests with race detector
test:
	go test ./... -race -count=1 -timeout=60s

# Run linter
lint:
	golangci-lint run ./...

# Build frontend only
frontend:
	cd web && npm install && npm run build

# Run migrations
migrate-up:
	go run ./cmd/server -migrate=up

migrate-down:
	go run ./cmd/server -migrate=down

# Docker
docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app
