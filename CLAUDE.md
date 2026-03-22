# Wiretap — Claude Code Project Guide

## What is Wiretap

Self-hosted application inspector. Single Go binary with embedded Vue UI.
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
│   │   ├── analyzer/   # N+1, duplicate, slow query detection
│   │   ├── storage/    # DuckDB operations + settings + sessions
│   │   ├── api/        # REST API + WebSocket + auth
│   │   └── config/     # Env-based configuration
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
- **Go 1.25+**
- **DuckDB** via go-duckdb — embedded OLAP storage
- **go-smtp** — SMTP server
- **chi** — HTTP router
- **gorilla/websocket** — real-time updates to UI
- **go-redis** — Redis receiver for ingest
- **bcrypt** (x/crypto) — password hashing (removed in favor of env-based plain comparison)
- **go:embed** — frontend assets baked into binary (prod only)
- **air** — hot-reload in dev

### Frontend (go-server/frontend/)
- **Vue 3** + **TypeScript**
- **Vite** — build tool + HMR dev server
- **TailwindCSS v4** — styling
- **vue-router** — SPA routing
- **ECharts** (vue-echarts) — dashboard charts
- **lucide-vue-next** — icons
- **Vitest** + **@vue/test-utils** — testing

### PHP Package (laravel-package/)
- **PHP 8.2+**, **Laravel 11/12/13**
- Zero-config via auto-discovery
- 18 listeners: queries, exceptions, jobs, cache, mail, notifications, http client, transactions, auth, commands, schedule, gate, var dumper, validation, broadcasts, logs, models, views
- HTTP + Redis drivers with gzip compression
- Orchestra Testbench for testing

## Configuration

### Wiretap Server (env variables)
```
WIRETAP_PASSWORD=secret           # UI login password (empty = no auth)
WIRETAP_INGEST_KEY=key            # Ingest API key (empty = allow all)
WIRETAP_APP_URL=http://app:8000   # Laravel app URL for replay
WIRETAP_REPLAY_SECRET=wt_xxx     # Replay secret (must match Laravel .env)
WIRETAP_REDIS_ADDR=redis:6379    # Redis address for Redis receiver (optional)
WIRETAP_STORAGE_PATH=/data/wiretap.duckdb
WIRETAP_WEB_PORT=8080
WIRETAP_INGEST_PORT=9090
WIRETAP_SMTP_PORT=1025
WIRETAP_RETENTION_DAYS=7
```

### Laravel .env
```
WIRETAP_ENABLED=true
WIRETAP_DRIVER=http               # or redis
WIRETAP_URL=http://wiretap:9090/ingest
WIRETAP_INGEST_KEY=key
WIRETAP_REPLAY_ENABLED=true
WIRETAP_REPLAY_SECRET=wt_xxx
```

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
| 6379 | redis | Redis for queues/cache/ingest |

### Commands (all via docker compose)

```bash
# Environment
make up                  # Start Go server only
make up-all              # Start all containers (Go + Laravel + Frontend)
make down                # Stop all
make logs                # Follow all logs
make logs-wiretap        # Follow Go server logs

# Go server
make go-test             # Run Go tests
make go-build            # Build binary inside container
make go-sh               # Shell into Go container

# Frontend
make ui-build            # Production build (for go:embed)
make ui-sh               # Shell into frontend container

# PHP / Laravel
make php-test            # Run PHP package tests
make app-sh              # Shell into Laravel container

# Prod
make release             # Build frontend + Docker image
```

## UI Pages (13 pages)

- **Dashboard** — stats cards, charts (7 time periods: 1m, 5m, 15m, 30m, 1h, 24h, 7d)
- **Activity** — root traces only (request, command, schedule) with WebSocket banner
- **Performance** — Top 10 Slow Endpoints + Top 10 Slow Queries
- **Requests** — HTTP requests with method/URI filters, warnings
- **Commands** — artisan commands with exit codes
- **Schedule** — scheduled tasks
- **Jobs** — queue jobs with Replay button
- **Exceptions** — traces with exceptions
- **Mail** — intercepted emails with HTML preview
- **Broadcasts** — broadcast events with channels
- **Notifications** — notification events with data
- **Logs** — log entries with level filter + detail page
- **Settings** — General (thresholds, retention), Storage (DB stats, clear), Info (env examples)

## Key Features

- **18 PHP listeners** — all Laravel events covered
- **Replay/Retry** — replay requests, retry/dispatch jobs, replay broadcasts, replay commands
- **Gzip compression** — PHP drivers compress, Go receiver decompresses
- **HTTP + Redis drivers** — Redis bypasses Laravel prefix automatically
- **Auth** — env-based password, session cookies in DuckDB
- **Ingest API key** — protects ingest endpoint
- **Ignore filters** — wildcard patterns for requests, jobs, commands, broadcasts, models, views
- **Field masking** — sensitive fields masked in request body, job payload
- **Job payload capture** — serialized payload for replay with editable properties
- **HTTP response body** — captured for outgoing HTTP calls (skip binary, 512KB limit)
- **Notification data** — public properties extracted via reflection
- **Model events** — created/updated/deleted with model class and ID
- **Views rendered** — blade view names in trace tree
- **Configurable thresholds** — slow query/request ms via Settings UI
- **Storage management** — DB size, trace/email/job counts, manual clear
- **NewItemsBanner** — non-intrusive WebSocket updates on all table pages
- **TraceTree** — all 18 event types formatted with enums, collapsible details, Open buttons

## Architecture Decisions

- **DuckDB over SQLite** — columnar OLAP, fast aggregations, no separate server
- **Buffered writes** — events buffered in memory, flushed to DuckDB every 500ms or 100 items
- **Request ID (rid)** — UUID propagated through request → jobs → child jobs → mail
- **go:embed** — Vue dist baked into Go binary, single file deployment
- **WebSocket** — real-time dashboard updates, NewItemsBanner for tables
- **Env-based config** — password, keys, URLs from environment (like Postgres/MySQL)
- **Docker-first dev** — all development inside containers, host stays clean
- **Redis driver** — bypasses Laravel prefix via PhpRedis/Predis raw client

## Code Conventions

- Go: standard library style, `internal/` packages, interfaces for testability
- Vue: `<script setup lang="ts">`, Composition API only, EventType enums
- PHP: PSR-12, Laravel conventions
- Tests: table-driven tests in Go, Vitest for frontend, PHPUnit + Orchestra Testbench for PHP
- Error messages: lowercase, no punctuation (Go convention)
- Commits: conventional commits (feat/fix/refactor/test/docs)
- Tables: `overflow-x-auto` on containers, `truncate max-w-*` on long text cells

## Testing

- **Go**: 46+ tests (storage, analyzer, API handlers, auth, sessions, ingest key, replay, dashboard)
- **PHP**: 83+ tests (all 18 listeners, replay controller, validation, commands, ignore filters, masking)
- **Vitest**: 45+ tests (API layer, NewItemsBanner, ReplayModal, Login)
- **Laravel compatibility**: tested on Laravel 11, 12, 13

## Important Notes

- Always run `make go-test` after modifying Go code
- Frontend Vite dev server (`:5173`) proxies API calls to Go server — both containers must be running
- DuckDB file is single-writer — no concurrent write connections
- PHP package supports Laravel 11, 12, and 13 (PHP 8.2+)
- Never install Go/Node/PHP locally — use `make *-sh` to get into containers
- `.env` file in project root for docker-compose env vars (WIRETAP_PASSWORD, WIRETAP_REPLAY_SECRET)
- `.env` is in .gitignore — contains secrets
