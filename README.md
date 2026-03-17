# HillsCheck

Fintech platform with RPG gamification. Automates transaction collection, classifies them with AI, and converts financial behavior into game character progress.

## Stack

- **Backend**: Go 1.24+ (Clean Architecture), Single Binary
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Frontend**: React + Tailwind (embedded via go:embed)
- **Classification**: Gemini / Ollama (chained with fallback)
- **Delivery**: Docker Compose

## Quick Start

```bash
cp .env.example .env
# fill in your GEMINI_API_KEY and other secrets
make docker-up
```

Open http://localhost:8080

## Development

```bash
make dev       # hot-reload with Air
make test      # run tests with -race
make lint      # golangci-lint
```

## Architecture

```
Clean Architecture layers:
  domain       → pure entities, no external imports
  usecase      → business logic + port interfaces
  adapter      → HTTP handlers, DB repos, AI providers
  infrastructure → DB pool, Redis, worker pool, config
```
