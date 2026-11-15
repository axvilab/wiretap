# Changelog

## 0.1.0

First public release. This is an early version: expect rough edges and breaking changes
before 1.0.

### Server (`ghcr.io/axvilab/wiretap`)
- One Docker image: ingest API, DuckDB storage, Vue UI, SMTP mail catcher, monitoring and notifications.
- Request traces with a tree of queries, jobs, cache, mail, HTTP client calls, logs and exceptions.
- N+1, duplicate and slow query detection.
- Replay of HTTP requests, jobs, broadcasts and commands, and retry of failed jobs.
- Performance pages: slowest endpoints and queries, response time trends, status codes.
- Site monitoring: uptime, SSL certificates, domain expiry, health URLs and scheduler heartbeats.
- Notifications to Telegram, Slack and Microsoft Teams.
- Password-protected UI and an optional ingest key.

### Laravel package (`axvi/wiretap-laravel`)
- Zero-config collector for Laravel 12 and 13 (PHP 8.2+), sent over HTTP or Redis with gzip.
- 18 listeners, each with its own switch, ignore list and data filter.
- `wiretap()` helper to send values to Wiretap without printing anything.
- Sensitive fields are masked before they leave the app.

### Fixed before release
- With `WIRETAP_PASSWORD` set, the UI itself was blocked, so the login form never loaded.
- `vendor:publish --tag=wiretap-config` did nothing until Wiretap was already enabled.
