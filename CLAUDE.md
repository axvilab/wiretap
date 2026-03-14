# Wiretap — Claude Code Project Guide

## What is Wiretap

Self-hosted Laravel application inspector. Single Go binary with embedded Vue UI.
Replaces Telescope, Mailpit, Pulse, Horizon UI, Flare, Ray, Clockwork — one binary, zero deps.

## Monorepo Structure

```
wiretap/
├── go-server/          # Go binary (main product)
│   ├── cmd/wiretap/    # Entrypoint
│   ├── internal/       # All business logic
│   │   ├── receiver/   # HTTP + Redis ingest
│   │   ├── smtp/       # SMTP server (email interception)
│   │   ├── correlator/ # Trace tree builder, job linker
│   │   ├── analyzer/   # N+1, duplicate query detection
│   │   ├── storage/    # DuckDB operations
│   │   └── api/        # REST API + WebSocket
│   ├── frontend/       # Vue 3 UI (embedded via go:embed in prod)
│   ├── Dockerfile.dev  # Dev image with air hot-reload
│   ├── Dockerfile      # Prod multi-stage build
│   └── .air.toml       # Air config for hot-reload
├── laravel-package/    # axvi/wiretap-laravel (Composer package)
├── test-app/           # Laravel test application
├── docker-compose.yml  # Full dev environment
├── Makefile            # All commands via docker compose
└── docs/               # WIRETAP-PLAN.md and other docs
```

## Tech Stack

### Go Server
- **Go 1.23+**
- **DuckDB** via go-duckdb — embedded OLAP storage
- **go-smtp** + **enmime** — SMTP server and MIME parsing
- **chi** or **gorilla/mux** — HTTP router
- **gorilla/websocket** — real-time updates to UI
- **go:embed** — frontend assets baked into binary (prod only)
- **air** — hot-reload in dev

### Frontend (go-server/frontend/)
- **Vue 3** + **TypeScript**
- **Vite** — build tool + HMR dev server
- **TailwindCSS** — styling
- **Pinia** — state management
- **vue-router** — SPA routing
- **ECharts** (vue-echarts) — dashboard charts
- **lucide-vue-next** — icons

### PHP Package (laravel-package/)
- **PHP 8.1+**, **Laravel 10+**
- Zero-config via auto-discovery
- Collectors for: requests, queries, jobs, mail, cache, logs, exceptions

## Development Environment

Everything runs in Docker. No Go, Node, or PHP on the host.

### Ports

| Port | Service | Purpose |
|------|---------|---------|
| 5173 | frontend | Vite dev server (main dev URL for UI) |
| 8080 | wiretap | Web UI (prod embedded assets) |
| 9090 | wiretap | Ingest API (receives events from PHP) |
| 1025 | wiretap | SMTP server (intercepts emails) |
| 8000 | laravel | Test Laravel app |
| 3306 | mysql | MySQL for test-app |
| 6379 | redis | Redis for queues/cache |

### Commands (all via docker compose)

```bash
# Environment
make up                  # Start all containers
make down                # Stop all
make logs                # Follow all logs
make logs-wiretap        # Follow Go server logs
make logs-frontend       # Follow Vite logs
make setup               # First run: up + migrate + seed

# Go server
make go-test             # docker compose exec wiretap go test ./...
make go-build            # Build binary inside container
make go-sh               # Shell into Go container

# Frontend
make ui-build            # Production build (for go:embed)
make ui-sh               # Shell into frontend container

# PHP / Laravel
make php-test            # Run PHP tests
make app-setup           # composer install + migrate + seed
make app-seed-traffic    # Generate test traffic
make app-sh              # Shell into Laravel container

# Prod
make release             # Build frontend + Docker image axvi/wiretap
```

## Development Plan

Full plan in `docs/WIRETAP-PLAN.md`. Phases:

1. **Phase 1** — Go server skeleton: ingest HTTP, DuckDB storage, correlator, basic API
2. **Phase 2** — Job correlation, trace tree builder, job linking
3. **Phase 3** — SMTP server, email capture and linking to traces
4. **Phase 4** — Vue UI: dashboard, request detail, jobs, mail viewer, exceptions, embed in Go
5. **Phase 5** — Polish: retention, graceful shutdown, Docker, config, performance
6. **Phase 6** — PHP package: collectors, auto-discovery, middleware, config

## Key Architecture Decisions

- **DuckDB over SQLite** — columnar OLAP, fast aggregations, no separate server
- **Buffered writes** — events buffered in memory, flushed to DuckDB every 500ms or 100 items
- **Request ID (rid)** — UUID propagated through request → jobs → child jobs → mail for full trace correlation
- **go:embed** — Vue dist baked into Go binary, single file deployment
- **SSE/WebSocket** — real-time UI updates, no polling
- **No pre-aggregation** — DuckDB handles dashboard aggregations natively
- **Docker-first dev** — all development inside containers, host stays clean

## Code Conventions

- Go: standard library style, `internal/` packages, interfaces for testability
- Vue: `<script setup lang="ts">`, Composition API only
- PHP: PSR-12, Laravel conventions
- Tests: table-driven tests in Go, Vitest for frontend, PHPUnit for PHP package
- Error messages: lowercase, no punctuation (Go convention)
- Commits: conventional commits (feat/fix/refactor/test/docs)

## Important Notes

- Always run `make go-test` after modifying Go code
- Frontend Vite dev server (`:5173`) proxies API calls to Go server — both containers must be running
- DuckDB file is single-writer — no concurrent write connections
- SMTP session handler must be goroutine-safe
- PHP package must support Laravel 10, 11, and 12
- Never install Go/Node/PHP locally — use `make *-sh` to get into containers