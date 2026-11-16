# Wiretap

Self-hosted application inspector and monitoring platform for Laravel.

There are many excellent tools in the Laravel ecosystem — Telescope, Pulse, Horizon, Ray, Clockwork, Flare, Mailpit, Oh Dear. They are well-designed, battle-tested, and loved by the community. But sometimes you need everything in one place, running on your own infrastructure, with full control over your data. Wiretap is built for that — a single Docker container that covers application inspection, email testing, performance analysis, and site monitoring.

For full documentation, screenshots, and feature overview visit **[axvi.dev/projects/wiretap](https://axvi.dev/projects/wiretap)**.

## How It Works

```
┌──────────────┐     HTTP/Redis      ┌─────────────────────────────────────┐
│  Laravel App │ ──────────────────> │           Wiretap                   │
│              │     (gzip)          │                                     │
│  composer    │                     │  :9090  Ingest API                  │
│  require     │                     │  :8080  Web UI (Vue 3, embedded)    │
│  axvi/       │ ──── SMTP ────────> │  :1025  SMTP Server                 │
│  wiretap-    │     (emails)        │                                     │
│  laravel     │                     │  DuckDB (embedded, columnar OLAP)   │
└──────────────┘                     └─────────────────────────────────────┘
```

Two components:
- **Docker image** (`ghcr.io/axvilab/wiretap`) — receiver, DuckDB storage, analyzer, REST API, embedded Vue UI, SMTP server, monitoring, notifications
- **PHP package** (`axvi/wiretap-laravel`) — lightweight event collector, zero-config, auto-discovery

## Quick Start

### 1. Run Wiretap

```bash
docker run -d \
  --name wiretap \
  -p 8080:8080 \
  -p 9090:9090 \
  -p 1025:1025 \
  -v wiretap_data:/data \
  ghcr.io/axvilab/wiretap
```

### 2. Install the Laravel package

```bash
composer require axvi/wiretap-laravel
```

### 3. Configure `.env`

```env
WIRETAP_ENABLED=true
WIRETAP_URL=http://wiretap:9090/ingest
```

Open http://localhost:8080 — done.

## Application Inspection

### 18 Event Types

Wiretap automatically captures requests, database queries, exceptions, queue jobs, mail, cache operations, notifications, outgoing HTTP calls, auth events, artisan commands, scheduled tasks, gate/policy checks, dump() calls, validation errors, transactions, broadcast events, model events, and rendered views.

Each listener has its own config section — enable/disable, ignore patterns (fnmatch wildcards), and data filtering (`'all'`, `'none'`, `'keys'`, or specific fields).

### Trace Tree

Every request, job, command, and scheduled task produces a trace — a tree of all events that happened during execution. You can see the full picture: which queries ran, what was cached, which jobs were dispatched, what emails were sent, all in one view.

### `wiretap()` Helper

A silent debug function — values appear only in the Wiretap UI, with zero effect on your application output.

```php
wiretap($variable);                    // inspect in UI
wiretap($order->total, 'order total'); // with label
$user = wiretap(User::find(1));        // inline — returns the value
```

### N+1 & Slow Query Detection

Automatic analysis on every request — detects N+1 query patterns, duplicate queries, and slow queries. Warning icons appear in the request list and trace tree. Thresholds are configurable via the Settings page.

### Replay & Retry

Replay HTTP requests, retry failed jobs, dispatch new jobs with edited payload, replay broadcasts, and re-run artisan commands — all from the UI.

## Mail

Wiretap handles emails in two ways:

**As a dev SMTP stub** — set `MAIL_HOST=wiretap` and `MAIL_PORT=1025`. Emails won't reach real recipients — they are caught by Wiretap's built-in SMTP server. You can view full HTML, text body, and download attachments directly in the UI.

**With any other mail driver** — if you use Postmark, SES, Mailgun, or any production SMTP, emails are delivered to real recipients as usual. But Wiretap still **logs the content** of every sent email through Laravel's mail events. You can see what was sent, to whom, with what subject and body, and which request or job triggered it.

Both approaches work simultaneously — you always have a complete picture of your outgoing mail.

## Performance Analysis

Dedicated Performance page with:
- **Top 10 Slow Endpoints** — hits count, average, P95, and max response time
- **Top 10 Slow Queries** — SQL, execution time, link to the trace

Dashboard with period selector (1 minute to 7 days):
- Request/error timeline charts
- Response time trends (average + P95)
- Status code distribution
- Stats cards with totals

## Site Monitoring

Monitor your sites directly from Wiretap — uptime, SSL certificates, domain expiration, and health endpoints. Useful when you have multiple services for one project — frontend, backend, API, admin panel — and want to keep an eye on all of them from one place.

### Uptime Checks
HTTP checks at configurable intervals (from 5 seconds). Supports Bearer token, Basic auth, custom headers, and query parameter authentication. Tracks status codes, response times, and uptime percentage.

### SSL Certificate Monitoring
Hourly TLS checks — issuer, subject, expiration date, days remaining. Configurable warning threshold (default 14 days before expiry).

### Domain WHOIS Monitoring
Daily WHOIS lookups — registrar, registry expiry date, days remaining. Warns when domain expires within 30 days. Automatic deduplication for sites sharing the same domain.

### Health Endpoint Monitoring
Polls health check URLs every 30 seconds. Parses Spatie Health format and generic JSON. Displays individual check results and overall status.

### Schedule Heartbeat Monitoring
Track scheduled tasks — expects a heartbeat within a configured interval or at a specific time of day. Alerts when a task misses its expected run.

## Notifications

Alert channels: **Telegram**, **Slack**, **Microsoft Teams** (via webhooks).

Alert rules with configurable cooldown:
- **Exception** — application exceptions
- **Site Down** — uptime check failures
- **SSL Expiry** — certificate approaching expiration
- **Domain Expiry** — domain registration expiring
- **Schedule Missed** — scheduled task didn't run

Full alert log with history of all sent notifications.

## UI Pages

| Page | Description |
|---|---|
| Dashboard | Stats cards, timeline charts, period selector (1m–7d) |
| Activity | Root traces (request, command, schedule, job) with live WebSocket updates |
| Performance | Top 10 slow endpoints + Top 10 slow queries |
| Requests | HTTP requests with method/URI filters, warning icons |
| Commands | Artisan commands with exit codes |
| Schedule | Scheduled tasks execution history |
| Jobs | Queue jobs with status, payload, replay |
| Exceptions | Traces with exceptions |
| Mail | Intercepted emails with HTML preview and attachments |
| Broadcasts | Broadcast events with channels and payload |
| Notifications | Notification events with extracted data |
| Logs | Log entries with level filter |
| Monitoring | Sites, notification channels, schedule heartbeats |
| Settings | Thresholds, retention, storage stats |

## Server Configuration

All configuration via environment variables:

```env
# Authentication
WIRETAP_PASSWORD=secret           # Web UI password (empty = no auth)
WIRETAP_INGEST_KEY=key            # Ingest API key (empty = allow all)

# Replay
WIRETAP_APP_URL=http://app:8000   # Laravel app URL
WIRETAP_REPLAY_SECRET=wt_xxx      # Must match Laravel .env

# Ports
WIRETAP_WEB_PORT=8080
WIRETAP_INGEST_PORT=9090
WIRETAP_SMTP_PORT=1025

# Storage
WIRETAP_STORAGE_PATH=/data/wiretap.duckdb
WIRETAP_RETENTION_DAYS=7

# Redis receiver (optional, for Redis driver)
WIRETAP_REDIS_ADDR=redis:6379
```

## Docker Compose

```yaml
services:
  wiretap:
    image: ghcr.io/axvilab/wiretap
    ports:
      - "8080:8080"
      - "9090:9090"
      - "1025:1025"
    volumes:
      - wiretap_data:/data
    environment:
      WIRETAP_PASSWORD: secret
      WIRETAP_INGEST_KEY: my-key
      WIRETAP_REPLAY_SECRET: wt_xxx
      WIRETAP_APP_URL: http://app:8000
      WIRETAP_RETENTION_DAYS: 7

volumes:
  wiretap_data:
```

## Security

- Password-protected web UI with HttpOnly session cookies (24h TTL)
- Ingest API key authentication
- CORS origin-specific (not wildcard)
- Rate limiting — 30 write requests/minute per IP
- Body size limits — 1MB for API, 50MB for ingest
- Sensitive field masking (recursive, case-insensitive)

## Tech Stack

- **Go 1.25+** — server, DuckDB storage, SMTP, WebSocket, monitoring
- **Vue 3 + TypeScript** — embedded SPA with Vite, TailwindCSS v4, ECharts
- **DuckDB** — embedded columnar OLAP database
- **PHP 8.2+ / Laravel 12–13** — event collector package

## From the author

Curious about Rust? I wrote two books for Laravel developers:
[Rust for Laravel Developers](https://leanpub.com/rust-for-laravel-developers) and
[Loco for Laravel Developers](https://leanpub.com/loco-for-laravel-developers).
More tools and posts on [axvi.dev](https://axvi.dev).

## License

MIT License. See [LICENSE](LICENSE) for details.