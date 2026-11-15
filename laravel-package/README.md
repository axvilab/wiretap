# Wiretap for Laravel

[![Latest Version on Packagist](https://img.shields.io/packagist/v/axvi/wiretap-laravel.svg)](https://packagist.org/packages/axvi/wiretap-laravel)
[![License](https://img.shields.io/packagist/l/axvi/wiretap-laravel.svg)](https://packagist.org/packages/axvi/wiretap-laravel)

Lightweight event collector for [Wiretap](https://axvi.dev/wiretap) — a self-hosted application inspector and monitoring platform for Laravel.

There are excellent tools in the Laravel ecosystem — Telescope, Pulse, Horizon, Ray, and many others. They are well-designed, battle-tested, and loved by the community. But sometimes you need a **private, self-hosted solution** where all your application data stays on your own infrastructure. Wiretap fills that gap — one Docker container that covers application inspection, email testing, performance analysis, and site monitoring.

For full documentation, screenshots, and feature overview visit **[axvi.dev/wiretap](https://axvi.dev/wiretap)**.

## Requirements

- PHP 8.2+
- Laravel 11, 12, or 13
- Running [Wiretap server](https://github.com/axvi/wiretap) (Docker)

## Installation

```bash
composer require axvi/wiretap-laravel
```

Zero-config — the service provider is auto-discovered.

Add to your `.env`:

```env
WIRETAP_ENABLED=true
WIRETAP_URL=http://wiretap:9090/ingest
```

Open the Wiretap UI at `http://localhost:8080` — your app is now being inspected.

## What Is Collected

Wiretap listens to **18 event types** automatically:

| # | Event | What is captured |
|---|---|---|
| 1 | **Requests** | Method, URI, status, duration, headers, body, IP |
| 2 | **Queries** | SQL, bindings, execution time, connection |
| 3 | **Exceptions** | Class, message, file, line, stack trace (10 frames) |
| 4 | **Jobs** | Class, queue, UUID, status, serialized payload for replay |
| 5 | **Mail** | To, from, subject, HTML/text body, attachments |
| 6 | **Cache** | Key, hit/miss, write (with TTL), forget |
| 7 | **Notifications** | Channel, class, notifiable, extracted public properties |
| 8 | **HTTP Client** | Outgoing method, URL, headers, response body (512KB limit, skips binary) |
| 9 | **Auth** | Login, logout, failed attempts with guard |
| 10 | **Commands** | Artisan command name, input, exit code |
| 11 | **Schedule** | Cron expression, command, success/failure |
| 12 | **Gate** | Ability, result, user ID |
| 13 | **Dumps** | `dump()` / `dd()` interception with file and line |
| 14 | **Validation** | Validation errors and rules |
| 15 | **Transactions** | Begin, commit, rollback with connection name |
| 16 | **Broadcasts** | Event class, channels (public/private/presence), payload |
| 17 | **Models** | Created, updated, deleted — class and key |
| 18 | **Views** | Blade view name, path, variables (configurable per-view) |

Everything is collected non-blocking and compressed with gzip before sending.

## The `wiretap()` Helper

A silent debug helper — like `dump()` but produces **zero output**. Values appear only in the Wiretap UI.

```php
// Simple — appears in trace tree
wiretap($user);

// With label
wiretap($order->total, 'order total');

// Inline — returns the value, works anywhere
$result = wiretap(User::find($id), 'found user');

// Inside chains
return wiretap($query->get(), 'query results')->map(...);
```

## Mail

Wiretap handles emails in two ways:

**As a dev SMTP stub** — set `MAIL_HOST=wiretap` and `MAIL_PORT=1025`. Emails won't reach real recipients — they are caught by Wiretap's built-in SMTP server. You can view full HTML, text body, and download attachments directly in the UI.

**With any other mail driver** — if you use Postmark, SES, Mailgun, or any production SMTP, emails are delivered to real recipients as usual. But Wiretap still **logs the content** of every sent email through Laravel's mail events. You can see what was sent, to whom, with what subject and body, and which request or job triggered it.

Both approaches work simultaneously — you always have a complete picture of your outgoing mail.

## Replay & Retry

Replay HTTP requests, retry failed jobs, dispatch jobs with edited payload, replay broadcasts, and re-run artisan commands — all from the UI.

Enable in `.env`:

```env
WIRETAP_REPLAY_ENABLED=true
WIRETAP_REPLAY_SECRET=your-secret-key
```

Generate a secret:

```bash
php artisan wiretap:generate-replay-secret
```

The replay secret must match between the Laravel app and the Wiretap server.

## Drivers

### HTTP (default)

Sends events via non-blocking HTTP POST with gzip compression.

```env
WIRETAP_DRIVER=http
WIRETAP_URL=http://wiretap:9090/ingest
WIRETAP_TIMEOUT_MS=50
```

### Redis

Sends events via Redis `LPUSH` — Wiretap server reads from the same Redis with `BRPOP`. Automatically bypasses Laravel's Redis prefix.

```env
WIRETAP_DRIVER=redis
WIRETAP_REDIS_CONNECTION=default
WIRETAP_REDIS_QUEUE=wiretap:events
```

Requires `ext-redis` (PhpRedis) or `predis/predis`.

## Configuration

Publish the config file:

```bash
php artisan vendor:publish --tag=wiretap-config
```

### Per-Type Configuration

Every listener has its own section with `enabled`, `ignore`, and data filtering:

```php
// config/wiretap.php

'queries' => [
    'enabled' => true,
    'ignore' => ['select * from sessions*'],  // fnmatch wildcard patterns
],

'jobs' => [
    'enabled' => true,
    'ignore' => ['App\\Jobs\\Heartbeat'],
    'payload' => 'all',  // 'all', 'none', 'keys', ['field1', 'field2']
],

'views' => [
    'enabled' => true,
    'ignore' => ['errors.*'],
    'data' => 'none',
    'overrides' => [
        'dashboard' => 'all',              // capture all vars for this view
        'emails.*' => ['user', 'order'],   // only these vars for email views
    ],
],

'commands' => [
    'enabled' => true,
    'ignore' => [
        'schedule:run', 'queue:work', 'horizon:*',
        'config:*', 'route:*', 'cache:*',
    ],
],
```

### Data Filtering

The `data` / `payload` / `body` options accept:

| Value | Behavior |
|---|---|
| `'all'` | Capture everything |
| `'none'` | Skip data entirely |
| `'keys'` | Capture field names only (no values) |
| `['field1', 'field2']` | Capture only specific fields |

### Ignore Patterns

Wildcard patterns (`fnmatch`) supported for 15 event types. Patterns match against the primary identifier of each event — URI for requests, SQL for queries, class name for jobs, etc.

### Field Masking

Sensitive fields are masked globally with `***`:

```php
'mask_fields' => [
    'password',
    'password_confirmation',
    'token',
    'secret',
    'credit_card',
],
```

Masking is recursive and case-insensitive — works on nested arrays and objects.

## Full `.env` Reference

```env
WIRETAP_ENABLED=true                    # Master switch
WIRETAP_DRIVER=http                     # http or redis
WIRETAP_URL=http://wiretap:9090/ingest  # HTTP driver endpoint
WIRETAP_TIMEOUT_MS=50                   # HTTP driver timeout (ms)
WIRETAP_INGEST_KEY=                     # API key (empty = no auth)
WIRETAP_REPLAY_ENABLED=false            # Enable replay routes
WIRETAP_REPLAY_SECRET=                  # Replay secret (must match server)
WIRETAP_MAX_PAYLOAD_BYTES=52428800      # Max payload before drop (50MB)
WIRETAP_REDIS_CONNECTION=default        # Redis driver connection
WIRETAP_REDIS_QUEUE=wiretap:events      # Redis driver queue name
```

## Wiretap Server

This package sends data to the Wiretap server. Run it with Docker:

```bash
docker run -d \
  --name wiretap \
  -p 8080:8080 \
  -p 9090:9090 \
  -p 1025:1025 \
  -v wiretap_data:/data \
  axvi/wiretap
```

| Port | Purpose |
|---|---|
| 8080 | Web UI (dashboard, traces, mail viewer, monitoring) |
| 9090 | Ingest API (receives events from this package) |
| 1025 | SMTP server (email interception for dev environments) |

For full server documentation, see [Wiretap on GitHub](https://github.com/axvi/wiretap).

## Testing

```bash
composer test
```

119+ tests covering all 18 listeners, `wiretap()` helper, `filterData`, per-type config, ignore patterns, replay, validation, and field masking. Tested on Laravel 11, 12, and 13.

## License

MIT License. See [LICENSE](LICENSE) for details.