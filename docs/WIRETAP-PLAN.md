# WIRETAP — Complete Development Plan for Claude Code CLI

## Project Overview

Wiretap is a self-hosted Laravel application inspector built as a single Go binary with embedded web UI. It replaces 7 separate tools (Telescope, Mailpit, Pulse, Horizon UI, Flare, Ray, Clockwork) with one binary.

Two components:
1. **Go binary** (wiretap) — receiver, SMTP server, DuckDB storage, analyzer, REST API, embedded Vue UI
2. **PHP package** (axvi/wiretap-laravel) — lightweight event collector, zero-config, auto-discovery

Three ports:
- `:8080` — Web UI (dashboard, traces, mail viewer, jobs, exceptions)
- `:9090` — Ingest API (receives events from PHP package)
- `:1025` — SMTP server (intercepts emails like Mailpit)

---

## Development Environment

### Monorepo Structure

```
wiretap/
├── go-server/                    # Go binary (main product)
│   ├── cmd/
│   │   └── wiretap/
│   │       └── main.go
│   ├── internal/
│   │   ├── receiver/
│   │   │   ├── receiver.go       # Interface
│   │   │   ├── http.go           # HTTP POST /ingest
│   │   │   └── redis.go          # Redis BRPOP
│   │   ├── smtp/
│   │   │   ├── server.go         # SMTP listener
│   │   │   ├── session.go        # SMTP session handler
│   │   │   └── parser.go         # MIME parsing
│   │   ├── correlator/
│   │   │   ├── correlator.go     # Groups events by rid, links jobs
│   │   │   ├── tree.go           # Recursive trace tree builder
│   │   │   └── mail_linker.go    # Links SMTP emails to traces
│   │   ├── analyzer/
│   │   │   ├── analyzer.go       # Orchestrates all analyses
│   │   │   ├── n_plus_one.go     # N+1 query detection
│   │   │   ├── duplicate.go      # Duplicate query detection
│   │   │   ├── slow.go           # Slow query/request detection
│   │   │   └── suggestions.go    # Human-readable suggestions
│   │   ├── storage/
│   │   │   ├── duckdb.go         # DuckDB operations
│   │   │   ├── migrations.go     # Auto-migrations on startup
│   │   │   ├── retention.go      # Old data cleanup
│   │   │   └── models.go         # Go struct definitions
│   │   ├── api/
│   │   │   ├── router.go         # All routes
│   │   │   ├── middleware.go     # CORS, auth, logging
│   │   │   ├── dashboard.go      # GET /api/dashboard
│   │   │   ├── requests.go       # GET /api/requests
│   │   │   ├── trace.go          # GET /api/trace/:rid
│   │   │   ├── exceptions.go     # GET/PATCH /api/exceptions
│   │   │   ├── jobs.go           # GET /api/jobs
│   │   │   ├── mail.go           # GET /api/mail, /api/mail/:id
│   │   │   ├── broadcasts.go     # GET /api/broadcasts
│   │   │   ├── replay.go         # POST /api/replay/*
│   │   │   └── ws.go             # WebSocket real-time push
│   │   └── config/
│   │       └── config.go         # YAML config parsing
│   ├── frontend/                  # Vue UI (embedded via go:embed)
│   │   ├── src/
│   │   │   ├── pages/
│   │   │   │   ├── Dashboard.vue
│   │   │   │   ├── Requests.vue
│   │   │   │   ├── Trace.vue
│   │   │   │   ├── Exceptions.vue
│   │   │   │   ├── Jobs.vue
│   │   │   │   ├── Mail.vue
│   │   │   │   ├── Broadcasts.vue
│   │   │   │   └── Settings.vue
│   │   │   ├── components/
│   │   │   │   ├── Waterfall.vue
│   │   │   │   ├── EventDetail.vue
│   │   │   │   ├── MailViewer.vue
│   │   │   │   ├── SqlHighlighter.vue
│   │   │   │   ├── StatsCard.vue
│   │   │   │   ├── TimeChart.vue
│   │   │   │   └── TraceTree.vue
│   │   │   ├── hooks/
│   │   │   │   ├── useWebSocket.ts
│   │   │   │   └── useFilters.ts
│   │   │   └── App.vue
│   │   ├── package.json
│   │   └── vite.config.ts
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   ├── Makefile
│   └── wiretap.yml                 # Default config
│
├── laravel-package/               # PHP composer package
│   ├── src/
│   │   ├── WiretapServiceProvider.php
│   │   ├── WiretapMiddleware.php
│   │   ├── Wiretap.php
│   │   ├── Contracts/
│   │   │   └── DriverInterface.php
│   │   ├── Drivers/
│   │   │   ├── HttpDriver.php
│   │   │   ├── RedisDriver.php
│   │   │   ├── UdpDriver.php
│   │   │   └── ShmDriver.php
│   │   └── Http/
│   │       └── ReplayController.php  # Optional replay routes
│   ├── config/
│   │   └── wiretap.php
│   ├── tests/
│   │   ├── Unit/
│   │   │   ├── WiretapTest.php
│   │   │   ├── BufferTest.php
│   │   │   └── DriverTest.php
│   │   └── Feature/
│   │       └── IntegrationTest.php
│   ├── composer.json
│   └── README.md
│
├── test-app/                      # Laravel test application
│   └── (standard Laravel project)
│   └── .env                       # Configured to use Wiretap
│
├── docker-compose.yml             # Full dev environment
├── Makefile                       # Root-level dev commands
└── README.md
```

### Docker Compose for Development

Вся разработка ведётся внутри контейнеров — никаких Go, Node, PHP на хосте.

```yaml
# docker-compose.yml — полное dev-окружение

services:
  # === Wiretap Go server (air hot-reload) ===
  wiretap:
    build:
      context: ./go-server
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"   # Web UI (prod embedded, в dev проксируется через frontend)
      - "9090:9090"   # Ingest API
      - "1025:1025"   # SMTP
    volumes:
      - ./go-server:/app
      - go-mod-cache:/go/pkg/mod
      - go-build-cache:/root/.cache/go-build
      - wiretap-data:/data
    environment:
      - WIRETAP_STORAGE_PATH=/data/wiretap.duckdb
      - WIRETAP_WEB_PORT=8080
      - WIRETAP_INGEST_PORT=9090
      - WIRETAP_SMTP_PORT=1025

  # === Vue frontend (Vite dev server с HMR) ===
  # В проде фронтенд встраивается в Go-бинарь через go:embed.
  # Этот контейнер только для dev — горячая перезагрузка, проксирование API к wiretap.
  frontend:
    image: node:22-alpine
    working_dir: /app
    command: sh -c "npm install && npm run dev -- --host 0.0.0.0"
    ports:
      - "5173:5173"   # Vite HMR — основной URL для разработки UI
    volumes:
      - ./go-server/frontend:/app
      - frontend-node-modules:/app/node_modules
    environment:
      - VITE_API_URL=http://wiretap:9090
      - VITE_WS_URL=ws://wiretap:8080
    depends_on:
      - wiretap

  # === Test Laravel application ===
  laravel:
    build:
      context: ./test-app
      dockerfile: Dockerfile
    ports:
      - "8000:8000"
    volumes:
      - ./test-app:/var/www/html
      - ./laravel-package:/var/www/packages/wiretap
    environment:
      - APP_ENV=local
      - APP_DEBUG=true
      - DB_CONNECTION=mysql
      - DB_HOST=mysql
      - DB_DATABASE=testapp
      - DB_USERNAME=root
      - DB_PASSWORD=secret
      - REDIS_HOST=redis
      - CACHE_DRIVER=redis
      - QUEUE_CONNECTION=redis
      - MAIL_MAILER=smtp
      - MAIL_HOST=wiretap
      - MAIL_PORT=1025
      - MAIL_ENCRYPTION=null
      - WIRETAP_ENABLED=true
      - WIRETAP_DRIVER=http
      - WIRETAP_URL=http://wiretap:9090/ingest
      - BROADCAST_DRIVER=redis
    depends_on:
      - mysql
      - redis
      - wiretap

  # Laravel queue worker (для тестирования job chains)
  worker:
    build:
      context: ./test-app
      dockerfile: Dockerfile
    command: php artisan queue:work redis --tries=3
    volumes:
      - ./test-app:/var/www/html
      - ./laravel-package:/var/www/packages/wiretap
    environment:
      - APP_ENV=local
      - DB_CONNECTION=mysql
      - DB_HOST=mysql
      - DB_DATABASE=testapp
      - DB_USERNAME=root
      - DB_PASSWORD=secret
      - REDIS_HOST=redis
      - QUEUE_CONNECTION=redis
      - WIRETAP_ENABLED=true
      - WIRETAP_DRIVER=http
      - WIRETAP_URL=http://wiretap:9090/ingest
    depends_on:
      - mysql
      - redis
      - wiretap

  # Laravel scheduler (для тестирования scheduled tasks)
  scheduler:
    build:
      context: ./test-app
      dockerfile: Dockerfile
    command: php artisan schedule:work
    volumes:
      - ./test-app:/var/www/html
      - ./laravel-package:/var/www/packages/wiretap
    environment:
      - APP_ENV=local
      - DB_CONNECTION=mysql
      - DB_HOST=mysql
      - DB_DATABASE=testapp
      - DB_USERNAME=root
      - DB_PASSWORD=secret
      - REDIS_HOST=redis
      - WIRETAP_ENABLED=true
      - WIRETAP_DRIVER=http
      - WIRETAP_URL=http://wiretap:9090/ingest
    depends_on:
      - mysql
      - redis
      - wiretap

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: testapp
    ports:
      - "3306:3306"
    volumes:
      - mysql-data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  mysql-data:
  wiretap-data:
  go-mod-cache:
  go-build-cache:
  frontend-node-modules:
```

### Dockerfile.dev для Go-сервера

```dockerfile
# go-server/Dockerfile.dev — dev-образ с air для hot-reload
FROM golang:1.23-bookworm

# DuckDB требует CGO
ENV CGO_ENABLED=1

# Air для hot-reload
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Кэширование зависимостей
COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080 9090 1025

CMD ["air", "-c", ".air.toml"]
```

### Root Makefile

Все команды работают через Docker — на хосте ничего не нужно кроме Docker и make.

```makefile
# Root Makefile — всё через docker compose, никаких локальных зависимостей

# === Полное окружение ===
up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

logs:
	docker compose logs -f

logs-wiretap:
	docker compose logs -f wiretap

logs-frontend:
	docker compose logs -f frontend

# === Go-сервер (внутри контейнера) ===
go-test:
	docker compose exec wiretap go test ./...

go-build:
	docker compose exec wiretap go build -o wiretap ./cmd/wiretap/

go-lint:
	docker compose exec wiretap golangci-lint run ./...

go-sh:
	docker compose exec wiretap sh

# === Frontend (внутри контейнера) ===
ui-build:
	docker compose exec frontend npm run build

ui-install:
	docker compose exec frontend npm install

ui-sh:
	docker compose exec frontend sh

# === PHP пакет (внутри контейнера) ===
php-test:
	docker compose exec laravel composer test --working-dir=/var/www/packages/wiretap

# === Test Laravel App ===
app-setup:
	docker compose exec laravel composer install
	docker compose exec laravel php artisan migrate
	docker compose exec laravel php artisan db:seed

app-seed-traffic:
	docker compose exec laravel php artisan wiretap:generate-traffic

app-sh:
	docker compose exec laravel bash

# === Prod build ===
release:
	docker compose exec frontend npm run build
	docker build -t axvi/wiretap ./go-server
	@echo "Release image ready: axvi/wiretap"

# === Всё вместе ===
setup: up
	@sleep 5
	@make app-setup
	@echo ""
	@echo "Ready!"
	@echo "  UI (dev):  http://localhost:5173"
	@echo "  API:       http://localhost:9090"
	@echo "  Laravel:   http://localhost:8000"
	@echo "  SMTP:      localhost:1025"

# === Release ===
release: ui-build go-build go-docker
	@echo "Release ready!"
```

---

## Test Laravel Application

The test-app is a purpose-built Laravel application that exercises every Laravel feature Wiretap monitors. It provides realistic test data and scenarios.

### Test App Routes and Features

```php
// routes/api.php — тестовые endpoints, каждый тригерит разные Laravel-фичи

// Простой запрос (queries + cache)
Route::get('/users', [UserController::class, 'index']);

// Сложный запрос (queries + events + jobs + mail + notifications + broadcast)
Route::post('/orders', [OrderController::class, 'store']);

// N+1 проблема (специально)
Route::get('/posts-with-n-plus-one', [PostController::class, 'indexBad']);

// Eager loading (правильно)
Route::get('/posts-optimized', [PostController::class, 'indexGood']);

// Ошибка 500
Route::get('/error', fn() => throw new \RuntimeException('Test error'));

// Медленный запрос
Route::get('/slow', [SlowController::class, 'index']);

// Внешний HTTP-вызов
Route::get('/external', [ExternalApiController::class, 'call']);

// Auth
Route::post('/login', [AuthController::class, 'login']);
Route::post('/logout', [AuthController::class, 'logout']);

// Validation error
Route::post('/validate', [ValidationController::class, 'store']);

// Feature flags
Route::get('/feature-test', [FeatureController::class, 'test']);

// Dump
Route::get('/dump-test', fn() => dump(['test' => 'data']));

// Broadcast
Route::post('/chat/message', [ChatController::class, 'send']);

// Gate/Policy
Route::get('/admin/dashboard', [AdminController::class, 'index']);

// Schedule test
// In Console/Kernel.php:
// $schedule->command('test:cleanup')->everyMinute();
// $schedule->job(new TestScheduledJob)->everyFiveMinutes();
```

### Test App Jobs (chain testing)

```php
// app/Jobs/SendOrderEmail.php
class SendOrderEmail implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public function handle()
    {
        // Query
        $order = Order::with('user', 'items')->find($this->orderId);

        // Mail
        Mail::to($order->user)->send(new OrderConfirmation($order));

        // Dispatch child job
        UpdateInventory::dispatch($order);

        // Broadcast
        broadcast(new OrderShipped($order));
    }
}

// app/Jobs/UpdateInventory.php — child job
class UpdateInventory implements ShouldQueue
{
    public function handle()
    {
        // Queries
        foreach ($this->order->items as $item) {
            $item->product->decrement('stock', $item->quantity);
        }

        // Event
        event(new InventoryUpdated($this->order));
    }
}
```

### Test App Artisan Command for Generating Traffic

```php
// app/Console/Commands/GenerateTraffic.php
// Генерирует реалистичный трафик для тестирования UI

class GenerateTraffic extends Command
{
    protected $signature = 'wiretap:generate-traffic {--requests=100}';

    public function handle()
    {
        $client = new \GuzzleHttp\Client(['base_uri' => 'http://localhost:8000']);

        for ($i = 0; $i < $this->option('requests'); $i++) {
            // Mix of different request types
            match(rand(0, 9)) {
                0 => $client->get('/api/users'),
                1 => $client->post('/api/orders', ['json' => $this->fakeOrder()]),
                2 => $client->get('/api/posts-with-n-plus-one'),
                3 => $client->get('/api/posts-optimized'),
                4 => $client->get('/api/slow'),
                5 => $client->get('/api/external'),
                6 => $client->post('/api/login', ['json' => $this->fakeLogin()]),
                7 => $client->post('/api/validate', ['json' => []]),  // validation error
                8 => $client->post('/api/chat/message', ['json' => $this->fakeMessage()]),
                9 => $client->get('/api/error'),  // 500 error
            };

            usleep(rand(10000, 100000)); // 10-100ms between requests
        }

        $this->info("Generated {$this->option('requests')} requests");
    }
}
```

---

## Detailed Development Phases

### PHASE 0: Project Setup (Day 1-2)

**Goal: Working dev environment, all repos initialized, test app running.**

Tasks:
```
[x] 0.1  Create monorepo directory structure as described above
[x] 0.2  Initialize Go module:
         cd go-server && go mod init github.com/axvi/wiretap
[x] 0.3  Initialize PHP package:
         cd laravel-package && composer init --name=axvi/wiretap-laravel
[x] 0.4  Create test Laravel app:
         composer create-project laravel/laravel test-app
[x] 0.5  Configure test-app composer.json to use local package:
         Add to repositories: {"type": "path", "url": "../laravel-package"}
         Then: composer require axvi/wiretap-laravel:@dev
[x] 0.6  Create docker-compose.yml as described above
[x] 0.7  Create Dockerfiles:
         - test-app/Dockerfile (PHP 8.3 + extensions + composer)
         - go-server/Dockerfile.dev (Go 1.22 + build-essential + air for hot reload)
         - go-server/Dockerfile (production multi-stage)
[x] 0.8  Create root Makefile with all dev commands
[x] 0.9  Run `make setup` — verify all containers start
[x] 0.10 Verify test-app is accessible at http://localhost:8000
[x] 0.11 Create go-server/cmd/wiretap/main.go with minimal "Hello Wiretap" HTTP server
[x] 0.12 Verify Go server starts and responds at http://localhost:8080
```

Testing:
```
- curl http://localhost:8000/api/users → Laravel responds
- curl http://localhost:8080 → Go responds "Wiretap running"
- docker-compose logs → no errors in any service
```

---

### PHASE 1: Minimal Data Pipeline (Week 1)

**Goal: PHP sends events → Go receives and stores → can query via API.**

#### 1A: Go Ingest Server

Tasks:
```
[x] 1A.1  Define Go data models in internal/storage/models.go:
          - TracePayload (what PHP sends)
          - Trace (what we store)
          - Event (individual event within trace)
          - JobLink (dispatch→execution mapping)
          
[x] 1A.2  Create internal/receiver/http.go:
          - POST /ingest endpoint
          - Parse JSON body into TracePayload struct
          - Log received payload to stdout (for debugging)
          - Return 202 Accepted
          
[x] 1A.3  Create internal/storage/duckdb.go:
          - Init DuckDB connection
          - Create file at configured path
          
[x] 1A.4  Create internal/storage/migrations.go:
          - Auto-create tables on startup
          - Schema: traces, job_links tables (emails table later)
          
[x] 1A.5  Implement storage.SaveTrace():
          - Parse events from JSON
          - Calculate counters (query_count, event_count, etc.)
          - INSERT into traces table
          - If events contain job.queued — INSERT into job_links
          
[x] 1A.6  Implement batch writer:
          - Buffer incoming traces in Go channel
          - Flush to DuckDB every 500ms or when buffer reaches 100 items
          - Single transaction per batch
          
[x] 1A.7  Create internal/api/router.go:
          - Mount ingest on :9090
          - Mount API + UI on :8080
          
[x] 1A.8  Create internal/api/requests.go:
          - GET /api/requests — list traces from DuckDB
          - Support query params: page, per_page, type, status, method
          - Return JSON with pagination
          
[x] 1A.9  Create internal/api/trace.go:
          - GET /api/trace/:rid — single trace with events_json parsed
          - Return full trace data as JSON

[x] 1A.10 Wire everything in main.go:
          - Load config
          - Init DuckDB
          - Run migrations
          - Start ingest server on :9090
          - Start API server on :8080
          - Start batch writer goroutine
          - Graceful shutdown on SIGTERM
```

Testing:
```
# Send fake event manually
curl -X POST http://localhost:9090/ingest \
  -H "Content-Type: application/json" \
  -d '{"v":1,"pid":1,"rid":"test-001","ts_start":1710000000.0,"ts_end":1710000000.1,"type":"request","meta":{"method":"GET","uri":"/test","status":200,"duration_ms":100},"events":[{"type":"query","ts":0.01,"data":{"sql":"SELECT 1","time_ms":0.5}}]}'

# Verify it's stored
curl http://localhost:8080/api/requests
# Should return JSON with our test trace

curl http://localhost:8080/api/trace/test-001
# Should return full trace with events
```

#### 1B: PHP Package — Core

Tasks:
```
[x] 1B.1  Create src/Wiretap.php:
          - Static $rid property
          - Static $buffer array
          - Static send(string $type, array $data) — appends to buffer
          - Static flush() — encodes buffer to JSON, sends via driver
          - Static reset() — clears buffer (for queue workers between jobs)
          
[x] 1B.2  Create src/Contracts/DriverInterface.php:
          - send(string $payload): void
          
[x] 1B.3  Create src/Drivers/HttpDriver.php:
          - Uses curl to POST JSON to configured endpoint
          - Timeout 50ms
          - Silently fails if server unreachable (@ error suppression)
          
[x] 1B.4  Create src/WiretapMiddleware.php:
          - handle(): Generate rid = uniqid('r_', true), set Wiretap::$rid
          - handle(): Wiretap::send('request.start', [method, uri, ip])
          - terminate(): Wiretap::send('request.end', [status, duration, memory])
          - terminate(): fastcgi_finish_request() if available
          - terminate(): Wiretap::flush()
          
[x] 1B.5  Create src/WiretapServiceProvider.php:
          - Register: bind DriverInterface to configured driver
          - Boot: push WiretapMiddleware globally
          - Boot: DB::listen → Wiretap::send('query', ...)
          - Boot: register exception handler via reportable()
          
[x] 1B.6  Create config/wiretap.php:
          - enabled, driver, drivers.http.endpoint, collect.* flags
          
[x] 1B.7  Configure composer.json:
          - autoload, extra.laravel.providers for auto-discovery
          
[x] 1B.8  Install in test-app: composer require axvi/wiretap-laravel:@dev
[x] 1B.9  Create test routes in test-app that trigger queries and exceptions
```

Testing:
```
# Make a request to test-app
curl http://localhost:8000/api/users

# Check Go server logs — should show received payload
docker-compose logs wiretap

# Check API
curl http://localhost:8080/api/requests
# Should show the /api/users request with query events

curl http://localhost:8080/api/trace/{rid-from-above}
# Should show full trace with SQL queries
```

---

### PHASE 2: Full Event Collection + Correlation (Week 2-3)

**Goal: All Laravel events captured, trace trees built with job chains.**

#### 2A: Complete PHP Listeners

Tasks:
```
[x] 2A.1  Add Event::listen('*') for all application events
          - Detect ShouldBroadcast for broadcast flag
          - Extract broadcast channels and event name
          - Filter out Illuminate\* events (configurable)
          
[x] 2A.2  Add Queue event listeners:
          - Queue::before → send('job.start', {job_uuid, job_class, queue})
          - Queue::after → send('job.end', {job_uuid, status:'ok', duration})
          - Queue::failing → send('job.end', {job_uuid, status:'failed', error})
          - JobQueued event → send('job.queued', {job_uuid, job_class, queue})
          - Reset Wiretap buffer between jobs (Wiretap::$rid = new uniqid for each job)
          
[x] 2A.3  Add Cache event listeners:
          - CacheHit → send('cache', {key, hit:true})
          - CacheMissed → send('cache', {key, hit:false})
          - KeyWritten → send('cache.write', {key, ttl})
          - KeyForgotten → send('cache.forget', {key})
          
[x] 2A.4  Add Mail tracking:
          - MessageSending → send('mail.sending', {to, subject, mailer})
          - MessageSent → send('mail.sent', {to, subject, mailer})
          
[x] 2A.5  Add Notification tracking:
          - NotificationSending/Sent → send('notification', {channel, class, notifiable})
          
[x] 2A.6  Add HTTP Client tracking:
          - Http::globalRequestMiddleware() → send('http.out.start', {url, method})
          - Http::globalResponseMiddleware() → send('http.out.end', {url, status, duration})
          
[x] 2A.7  Add Auth event listeners:
          - Login, Logout, Failed, Registered events
          
[x] 2A.8  Add Artisan command tracking:
          - CommandStarting, CommandFinished events
          
[x] 2A.9  Add Schedule tracking:
          - ScheduledTaskStarting, ScheduledTaskFinished, ScheduledTaskFailed
          
[x] 2A.10 Add Gate/Policy tracking:
          - GateEvaluated event (Laravel 10+)
          
[x] 2A.11 Add VarDumper intercept:
          - VarDumper::setHandler → send('dump', {value, file, line})
          - Call original handler too so dump() still works
          
[x] 2A.12 Add Validation tracking:
          - Catch ValidationException → send('validation', {errors, rules})
          
[x] 2A.13 Add Transaction tracking:
          - DB events for begin/commit/rollback
          
[x] 2A.14 Add optional request body/headers capture:
          - Only if config('wiretap.collect.request_body') is true
          - Apply mask_fields to sanitize PII

[x] 2A.15 Add optional replay routes (ReplayController):
          - POST /_wiretap/replay-request (re-sends HTTP request)
          - POST /_wiretap/retry-job (retries failed job by uuid)
          - POST /_wiretap/replay-broadcast (re-broadcasts event)
          - Protected by wiretap.replay.secret config
          - Only registered if wiretap.replay.enabled = true
```

#### 2B: Go Correlator

Tasks:
```
[x] 2B.1  Create internal/correlator/correlator.go:
          - ProcessTrace(payload) — called after storage.SaveTrace
          - Check events for 'job.queued' → insert job_links
          - Check if this trace is a job → look up job_links by job_uuid
          - If found → set parent_rid on this trace

[x] 2B.2  Create internal/correlator/tree.go:
          - BuildTree(rid) → returns recursive tree structure
          - Load root trace
          - Find all children (traces where parent_rid = root.rid)
          - For each child, recursively find their children
          - Return nested structure

[x] 2B.3  Update GET /api/trace/:rid:
          - Use tree.BuildTree() instead of flat trace
          - Return: {trace: {..., events}, children: [{trace, events, children: [...]}]}
          
[x] 2B.4  Implement job status tracking:
          - When job.start received → UPDATE job_links SET child_rid, started_at, status='processing'
          - When job.end received → UPDATE job_links SET finished_at, status='completed'|'failed'
```

#### 2C: Go Analyzer

Tasks:
```
[x] 2C.1  Create internal/analyzer/analyzer.go:
          - Analyze(trace) → returns []Warning
          - Runs all sub-analyzers on the trace's events

[x] 2C.2  Create internal/analyzer/n_plus_one.go:
          - Normalize SQL (replace values with ?)
          - Group by normalized SQL
          - If same SELECT appears 3+ times → N+1 warning
          - Include suggestion: "Consider eager loading with ::with()"

[x] 2C.3  Create internal/analyzer/duplicate.go:
          - Find exact same SQL + bindings appearing 2+ times
          - Warning: "Duplicate query executed N times, consider caching"

[x] 2C.4  Create internal/analyzer/slow.go:
          - Flag queries > threshold_ms
          - Flag requests > threshold_ms
          - Include the specific slow SQL in warning

[x] 2C.5  Save warnings in traces.warnings_json
[x] 2C.6  Set boolean flags: has_n_plus_one, has_slow_query, has_duplicate_query
```

Testing:
```
# Test full chain: request → event → job → child job → mail

# 1. Trigger an order in test-app
curl -X POST http://localhost:8000/api/orders \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1, "quantity": 2}'

# 2. Wait 2-3 seconds for queue worker to process jobs

# 3. Check trace — should show full tree
curl http://localhost:8080/api/requests
# Find the POST /api/orders request

curl http://localhost:8080/api/trace/{rid}
# Should show:
# - Request with queries, events, job.queued
# - Child: Job:SendOrderEmail with queries, mail, job.queued
# - Grandchild: Job:UpdateInventory with queries

# 4. Test N+1 detection
curl http://localhost:8000/api/posts-with-n-plus-one
curl http://localhost:8080/api/trace/{rid}
# Should have warnings_json with N+1 warning

# 5. Test exception tracking
curl http://localhost:8000/api/error
curl http://localhost:8080/api/trace/{rid}
# Should have exception event in trace

# 6. Generate bulk traffic
docker exec -it test-app php artisan wiretap:generate-traffic --requests=500
curl http://localhost:8080/api/requests?page=1
# Should show paginated list of diverse requests
```

---

### PHASE 3: SMTP Server (Week 3-4)

**Goal: Intercept emails like Mailpit, link to traces.**

Tasks:
```
[x] 3.1  Add go-smtp and enmime to go.mod:
         go get github.com/emersion/go-smtp
         go get github.com/jhillyerd/enmime

[ ] 3.2  Create internal/smtp/server.go:
         - StartSMTPServer(addr, storage) — listens on :1025
         - AllowInsecureAuth = true (local dev)
         - MaxMessageBytes = 25MB

[ ] 3.3  Create internal/smtp/session.go:
         - Implement smtp.Session interface: Mail(), Rcpt(), Data(), Reset(), Logout()
         - In Data(): read raw email bytes, parse with enmime

[ ] 3.4  Create internal/smtp/parser.go:
         - ExtractEmail(raw []byte) → Email struct
         - Parse: from, to, cc, subject, html, text, headers
         - Extract attachments: name, size, content_type, bytes

[x] 3.5  Add emails table to DuckDB migrations

[x] 3.6  Implement storage.SaveEmail():
         - Insert into emails table
         - Store raw bytes for full fidelity

[x] 3.7  Create internal/correlator/mail_linker.go:
         - LinkEmailToTrace(email) — runs async after save
         - Match by: to_address + subject + time_window (±5 seconds)
         - If matched: UPDATE emails SET trace_rid = ?

[x] 3.8  Create internal/api/mail.go:
         - GET /api/mail — paginated list (from, to, subject, time, has_trace)
         - GET /api/mail/:id — full email metadata
         - GET /api/mail/:id/html — raw HTML body (for iframe rendering)
         - GET /api/mail/:id/source — raw email source
         - GET /api/mail/:id/attachments/:filename — download attachment
         - DELETE /api/mail/:id — delete single email
         - DELETE /api/mail — delete all emails

[x] 3.9  Wire SMTP server in main.go — start on :1025

[x] 3.10 Configure test-app .env: MAIL_HOST=wiretap, MAIL_PORT=1025
```

Testing:
```
# 1. Send email from test-app
curl -X POST http://localhost:8000/api/orders \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1, "quantity": 2}'

# 2. Check mail API
curl http://localhost:8080/api/mail
# Should show the order confirmation email

curl http://localhost:8080/api/mail/{id}
# Should show full email with HTML body

curl http://localhost:8080/api/mail/{id}/html
# Should return raw HTML for preview

# 3. Verify trace linkage
curl http://localhost:8080/api/trace/{order-rid}
# The trace should reference the email
# The email should have trace_rid set

# 4. Test with plain SMTP (without Wiretap PHP package)
# Use swaks or any SMTP client:
swaks --to test@example.com --from sender@example.com \
      --server localhost:1025 --body "Test email"
curl http://localhost:8080/api/mail
# Should show email without trace linkage

# 5. Test attachments
# Send email with attachment from test-app
curl http://localhost:8080/api/mail/{id}
# Should list attachments with download URLs
```

---

### PHASE 4: Web UI (Week 4-6)

**Goal: Full Vue UI embedded in Go binary.**

#### 4A: Frontend Setup

Tasks:
```
[x] 4A.1  Initialize Vue project:
          cd go-server/frontend
          npm create vite@latest . -- --template vue-ts
          
[x] 4A.2  Install dependencies:
          npm install tailwindcss @tailwindcss/vite
          npm install echarts vue-echarts
          npm install vue-router
          npm install lucide-vue-next
          npm install @vueuse/core
          npx shadcn-vue@latest init

[x] 4A.3  Configure Tailwind

[x] 4A.4  Configure Vite proxy for dev:
          // vite.config.ts
          proxy: {
            '/api': 'http://localhost:8080',
            '/ws': { target: 'ws://localhost:8080', ws: true }
          }
          
[x] 4A.5  Create layout: sidebar navigation + main content area
          - Navigation items: Dashboard, Requests, Jobs, Exceptions, Mail, Broadcasts, Settings
          - Connection status indicator
          - App name selector (for multi-app, later)
```

#### 4B: Dashboard Page

Tasks:
```
[x] 4B.1  Create StatsCard component (number + label + trend arrow)
[x] 4B.2  Create TimeChart component (vue-echarts area/bar chart)
[x] 4B.3  Create Dashboard page:
          - 5 stat cards: Requests, Avg Duration, Errors, Jobs, Emails
          - Area chart: Requests per minute
          - Bar chart: Errors per minute
          - Table: Slowest Endpoints (top 10)
          - Table: Recent Exceptions (top 5)
          - Period selector: 1h | 6h | 24h | 7d
[x] 4B.4  Create GET /api/dashboard endpoint in Go with DuckDB aggregations
[x] 4B.5  Auto-refresh via WebSocket push
```

#### 4C: Requests Page

Tasks:
```
[x] 4C.1  Create Requests page:
          - Filterable table with columns: Time, Method, URI, Status, Duration, Queries, Jobs, Warnings
          - Filter bar: Status, Method, URI search, Slow threshold, Has exceptions
          - Pagination
          - Color coding: green (2xx), yellow (4xx), red (5xx)
          - Warning icons: ⚠ slow, 🔄 N+1, ❌ exception
          - Click row → navigate to Trace page
[x] 4C.2  Create useFilters hook for URL-based filter state
```

#### 4D: Trace Page (most important)

Tasks:
```
[x] 4D.1  Create Waterfall component:
          - Vertical timeline of events
          - Each event = row with colored duration bar
          - Colors: blue (query), purple (event), orange (job), green (mail),
                    red (exception), gray (cache), cyan (http.out), pink (broadcast)
          - Bar width proportional to duration
          - Dashed separator line for async boundary (request end → job start)
          - Indent for child jobs (recursive tree rendering)
          - Click on event → shows detail panel

[x] 4D.2  Create EventDetail panel (right side):
          - For query: SQL with syntax highlighting, bindings, duration, connection
          - For job: class, queue, uuid, status, duration, error
          - For exception: class, message, file:line, full stack trace
          - For mail: to, subject, mailer + "View email →" link
          - For broadcast: channel, event, payload JSON
          - For cache: key, hit/miss, TTL
          - For http.out: URL, method, status, duration
          - For dump: value, file:line
          - For event: name, listeners count
          - For auth: user, guard, success/fail

[x] 4D.3  Create Warnings panel (bottom):
          - List of analyzer warnings
          - N+1 with affected SQL and suggestion
          - Duplicate queries
          - Slow queries

[x] 4D.4  Create TraceTree component:
          - Renders trace header (method + uri + status + duration)
          - Renders events via Waterfall
          - Recursively renders children (child jobs)
          - Each child has its own header and waterfall

[x] 4D.5  Add Replay button:
          - "▶ Replay Request" button in header
          - Sends POST to /api/replay/request with original request data
          - Shows success/error notification

[x] 4D.6  Add linked email display:
          - If trace has linked emails, show "📨 1 email" badge
          - Click → navigate to Mail page with that email
```

#### 4E: Exceptions Page

Tasks:
```
[x] 4E.1  Create Exceptions page:
          - Grouped list of unique exceptions
          - For each: class, message, count, first/last seen, status badge
          - Filters: Status (open/resolved/ignored)
          - Sort: count, last_seen
          - Actions: Resolve, Ignore, View latest trace
          - Click → navigates to latest trace with that exception
          
[x] 4E.2  Implement PATCH /api/exceptions/:hash in Go
```

#### 4F: Jobs Page

Tasks:
```
[x] 4F.1  Create Jobs page:
          - Stats bar: throughput/min, success rate, failed count
          - Table: Class, Queue, Status, Duration, Dispatched At
          - Filters: Status (completed/failed), Queue name
          - Failed jobs: show error message + "View trace" link
          - "🔄 Retry" button for failed jobs
          - Chart: jobs per minute + failure rate over time
```

#### 4G: Mail Page

Tasks:
```
[x] 4G.1  Create MailViewer component:
          - Split view: list (left) + preview (right)
          - List: from, to, subject, time, trace badge
          - Preview tabs: HTML | Text | Source | Headers
          - HTML preview in sandboxed iframe
          - Attachment list with download buttons
          - "🔗 View trace" link if trace_rid exists
          - "Delete" and "Delete all" buttons
          - Search by to/from/subject
```

#### 4H: Broadcasts Page

Tasks:
```
[x] 4H.1  Create Broadcasts page:
          - Table: Channel, Event, Time, Trace link
          - Filter by channel pattern
          - Click → payload detail
          - "📡 Replay" button
```

#### 4I: Settings Page

Tasks:
```
[x] 4I.1  Create Settings page:
          - Connection status (is PHP sending data? last event time)
          - Storage stats (DuckDB file size, trace count, email count)
          - Retention settings display
          - Driver info (which receiver is active)
          - Version info
          - Link to docs
```

#### 4J: Embed and WebSocket

Tasks:
```
[x] 4J.1  Build frontend: cd frontend && npm run build
[x] 4J.2  Embed in Go:
          //go:embed frontend/dist/*
          var frontendFS embed.FS
          // Serve from :8080 root
          
[x] 4J.3  Implement WebSocket server:
          - internal/api/ws.go
          - On new trace saved → broadcast to all WS clients
          - On new exception → broadcast
          - On new email → broadcast
          - On job failed → broadcast
          
[x] 4J.4  Create useWebSocket hook in Vue:
          - Connect to ws://host/api/ws
          - On message → update relevant page data
          - Auto-reconnect on disconnect
```

Testing:
```
# 1. Build and start
make ui-build
make go-build
./go-server/wiretap

# 2. Open http://localhost:8080 — should show Dashboard

# 3. Generate traffic
docker exec -it test-app php artisan wiretap:generate-traffic --requests=200

# 4. Verify all pages:
- Dashboard: cards update, charts show data
- Requests: list populated, filters work, click → trace
- Trace: waterfall renders correctly, detail panel works
- Exceptions: grouped correctly, resolve/ignore works
- Jobs: shows completed and failed jobs
- Mail: emails appear after order creation, HTML preview works, trace linked
- Broadcasts: events listed

# 5. Test real-time: 
# Open Dashboard in browser
# Send request → UI updates without refresh

# 6. Test replay:
# Open a trace → click Replay → new trace appears
```

---

### PHASE 5: Redis Driver + Polish (Week 7)

**Goal: Production-ready, Redis driver, Docker release.**

Tasks:
```
[ ] 5.1  Create Redis driver in PHP package:
         - src/Drivers/RedisDriver.php
         - Uses Redis::lpush('wiretap:events', $payload)
         
[ ] 5.2  Create Redis receiver in Go:
         - internal/receiver/redis.go
         - BRPOP loop reading from wiretap:events
         - Parse JSON, feed to same pipeline as HTTP receiver
         
[ ] 5.3  Config: driver selection (http | redis) in both PHP and Go

[x] 5.4  Implement retention cleanup:
         - Goroutine running every hour
         - DELETE FROM traces WHERE created_at < now() - retention_days
         - DELETE FROM emails WHERE received_at < now() - email_retention_days
         - VACUUM (or DuckDB equivalent for space reclaim)

[x] 5.5  Add YAML config file support (wiretap.yml)
[x] 5.6  Add CLI flags: --config, --web-port, --ingest-port, --smtp-port
[x] 5.7  Production Dockerfile (multi-stage, minimal image)
[x] 5.8  Add health check endpoint: GET /api/health
[x] 5.9  Graceful shutdown: finish batch writes, close DuckDB, close SMTP

[ ] 5.10 Error handling: what happens when DuckDB is full, disk full, etc.
[ ] 5.11 Logging: structured JSON logs from Go server

[ ] 5.12 Write README.md for go-server:
         - Quick start (Docker and binary)
         - Configuration reference
         - Screenshots/GIFs

[ ] 5.13 Write README.md for laravel-package:
         - Installation
         - Configuration
         - What is collected
         - Drivers

[ ] 5.14 Load testing:
         - php artisan wiretap:generate-traffic --requests=10000
         - Monitor DuckDB file size
         - Monitor Go memory usage
         - Monitor write latency

[ ] 5.15 Test on real Laravel project (your own work project or open-source app)
```

---

### PHASE 6: Launch (Week 8)

Tasks:
```
[ ] 6.1  Create GitHub repositories (or Codeberg):
         - axvi/wiretap (Go server)
         - axvi/wiretap-laravel (PHP package)
         
[ ] 6.2  Push code, tag v0.1.0
[ ] 6.3  Build Docker image, push to Docker Hub or ghcr.io
[ ] 6.4  Publish PHP package to Packagist
[ ] 6.5  Create landing page (axvi.dev/wiretap)
[ ] 6.6  Write launch posts:
         - DEV.to article
         - r/laravel post
         - r/selfhosted post
         - r/golang post
         - Hacker News "Show HN"
         - Laravel News submission
         - Twitter/X thread
[ ] 6.7  Create 2-minute demo video (screen recording with narration)
[ ] 6.8  Monitor feedback, fix critical bugs
```

---

### PHASE 7: Advanced Features (Week 9-14)

Project is fully open-source and free. No monetization, no license keys, no paid tiers.

Tasks:
```
[ ] 7.1  Multi-app support:
         - PHP sends app_name in payload
         - Go stores per-app
         - UI: app selector dropdown

[x] 7.2  Auth for dashboard:
         - Basic auth middleware in Go
         - Configured via wiretap.yml

[ ] 7.3  Alerts — Slack:
         - internal/notifications/slack.go
         - Webhook URL in config
         - Alert rules: exception_spike, slow_endpoint, job_failure_rate

[ ] 7.4  Alerts — Telegram:
         - internal/notifications/telegram.go
         - Bot token + chat_id in config

[ ] 7.5  User tracing:
         - GET /api/users/:id/traces
         - UI page: enter user_id → see all their requests

[ ] 7.6  Deployment comparison:
         - Group metrics by git_hash
         - Show: "After deploy abc123, errors on /api/payments increased 300%"

[ ] 7.7  UDP driver (PHP + Go sides)
[ ] 7.8  SHM driver (PHP + Go sides)
```

---

## Key Technical Decisions

### DuckDB Configuration
```go
// On startup:
db.Exec("SET memory_limit='256MB'")
db.Exec("SET threads TO 2")
db.Exec("SET enable_progress_bar=false")

// For writes:
// Use batch INSERT in transactions (see Phase 1)
// Flush every 500ms or 100 items

// For reads (dashboard):
// DuckDB handles aggregations natively — no pre-aggregation needed
// time_bucket() for time-series charts
// PERCENTILE_CONT() for p95/p99
```

### Event Format from PHP
```json
{
  "v": 1,
  "app": "my-app",
  "env": "local",
  "pid": 42,
  "rid": "r_6789abc",
  "ts_start": 1710000000.123,
  "ts_end": 1710000000.468,
  "type": "request",
  "meta": {
    "method": "POST",
    "uri": "/api/orders",
    "route": "api.orders.store",
    "controller": "OrderController@store",
    "status": 201,
    "duration_ms": 345,
    "memory_peak_mb": 24.5,
    "ip": "192.168.1.1",
    "user_id": 42,
    "git_hash": "a1b2c3d"
  },
  "events": [
    {"type":"query","ts":0.002,"data":{"sql":"SELECT...","time_ms":0.8,"connection":"mysql"}},
    {"type":"event","ts":0.008,"data":{"name":"OrderCreated","broadcast":true,"channels":["private-orders.42"]}},
    {"type":"job.queued","ts":0.010,"data":{"job":"SendOrderEmail","job_uuid":"550e8400-...","queue":"emails"}},
    {"type":"cache","ts":0.012,"data":{"action":"put","key":"order:123","ttl":3600}},
    {"type":"mail.sent","ts":0.015,"data":{"to":["john@test.com"],"subject":"Order #123"}},
    {"type":"dump","ts":0.018,"data":{"value":"Order{id:123}","file":"OrderController.php","line":45}}
  ]
}
```

### Trace Tree API Response
```json
{
  "trace": {
    "rid": "r_aaa",
    "type": "request",
    "method": "POST",
    "uri": "/api/orders",
    "status": 201,
    "duration_ms": 345,
    "user_id": 42,
    "query_count": 5,
    "has_n_plus_one": false,
    "events": [...],
    "warnings": [
      {"type": "slow_request", "message": "Request took 345ms (threshold: 200ms)"}
    ]
  },
  "children": [
    {
      "trace": {
        "rid": "r_bbb",
        "type": "job",
        "job_class": "SendOrderEmail",
        "job_uuid": "550e8400-...",
        "status": "completed",
        "duration_ms": 25,
        "started_at_offset_ms": 2300,
        "events": [...]
      },
      "children": [
        {
          "trace": {
            "rid": "r_ccc",
            "type": "job",
            "job_class": "UpdateInventory",
            "duration_ms": 8,
            "events": [...]
          },
          "children": []
        }
      ]
    }
  ],
  "linked_emails": [
    {"id": "email-123", "to": "john@test.com", "subject": "Order #123"}
  ]
}
```

---

## How to Use This Plan with Claude Code CLI

Load this file as context when starting a new coding session:

```bash
# Start working on Phase 1A (Go ingest server):
claude --context wiretap-plan.md "Let's start Phase 1A. Create the Go project structure and implement the ingest HTTP server with DuckDB storage."

# Start working on Phase 1B (PHP package):
claude --context wiretap-plan.md "Let's start Phase 1B. Create the Laravel package with ServiceProvider, Middleware, and HttpDriver."

# Continue a phase:
claude --context wiretap-plan.md "Continue Phase 2A. We finished 2A.1-2A.5 last session. Now implement cache, mail, and notification listeners."

# Debug an issue:
claude --context wiretap-plan.md "The job correlation isn't working. When a job runs, Go doesn't link it to the parent request. The job_uuid is being sent but BuildTree returns no children."
```

Each phase has clear:
- Tasks with checkboxes (track progress)
- Testing steps (verify before moving on)
- File paths (know exactly what to create)
- Expected behavior (know when it's working)

Start with Phase 0, verify everything works, then proceed to Phase 1A.

---

## TODO

### Next
- [x] Replay/retry механизм (2A.15: ReplayController в PHP, кнопки в UI)
- [x] Тесты (Go, PHP, Vitest), CI/CD, README, GitHub release

### Backlog
- [ ] Протестировать PHP пакет на Laravel 10 и 11 (сейчас тестируем только на 12)
  - Swift Mailer vs Symfony Mailer в mail listeners
  - JobQueued event format differences
  - Добавить compatibility layer или дропнуть Laravel 10

### Done (2026-03-14)
- [x] Phase 0: Docker-first dev environment (docker-compose, air, Makefile)
- [x] Phase 1A: Go ingest server + DuckDB storage + batch writer + API
- [x] Phase 1B: PHP package (middleware, listeners, HttpDriver) + test-app
- [x] Phase 2A: All PHP listeners (queries, exceptions, jobs, cache, mail, notifications, HTTP client, transactions, auth, artisan, schedule, gate, vardumper, validation, broadcasts)
- [x] Phase 2B: Go correlator (job linking, trace tree builder)
- [x] Phase 2C: Go analyzer (N+1, duplicate query, slow query detection)
- [x] Phase 3: SMTP sink (dumb accept+discard, email capture via PHP events)
- [x] Phase 4: Vue UI (Dashboard, Requests, Trace detail, Exceptions, Jobs, Mail viewer, Broadcasts, Settings)
- [x] Phase 4B: Dashboard with ECharts (timeline, response time, status donut, slow endpoints, period selector 1h/24h/7d)
- [x] Phase 4H: Broadcasts page + detail page with channel types (public/private/presence)
- [x] Phase 4I: Settings page (server status, WebSocket, config hints)
- [x] Phase 4J: WebSocket real-time updates
- [x] Phase 5: Config, retention, prod Dockerfile with go:embed, graceful shutdown
- [x] Phase 6: PHP package polish (Redis driver, all collectors, config flags)
- [x] Mail viewer: HTML preview in iframe, HTML/Text toggle, attachment download via PHP events
- [x] Job correlation: request → ProcessOrder child trace verified
- [x] N+1 detection: warning flags + messages working

### Done (2026-03-15)
- [x] Refactor: single email pipeline — PHP captures via MessageSent event, SMTP is dumb sink, removed enmime
- [x] Broadcast listener: ShouldBroadcast events via Event::listen('*'), channel types (public/private/presence)
- [x] Dashboard: ECharts (requests timeline, response time avg/p95, status codes donut), period selector, stats cards (requests, jobs, exceptions, emails, avg, median, p95, req/min), slow endpoints, recent activity
- [x] Trace tree: recursive accordion (children collapsed by default, click to expand), human-readable event display, Open buttons for mail/broadcast, status badges on job.end
- [x] UI consistency: tabular layout for all list pages (Requests, Jobs, Mail, Exceptions, Broadcasts)
- [x] Pagination component: reusable across all pages, 25/50/100 per page selector
- [x] Constants: TraceType, EventType, JobStatus, ROOT_TRACE_TYPES — no magic strings
- [x] MethodBadge: colored backgrounds (GET green, POST amber, PUT sky, PATCH orange, DELETE red)
- [x] Dashboard stats cards clickable → navigate to corresponding pages
- [x] Slow endpoints clickable → Requests filtered by URI
- [x] URI search filter in Requests page
- [x] Requests page: only HTTP requests (no jobs), type filter removed
- [x] Dashboard: Recent Activity shows root types only (request, command, schedule), not jobs
- [x] Commands page: list with command name, exit code badge, duration, queries
- [x] Schedule page: list with cron expression badge, command, duration
- [x] Command → Job → Child Job tree verified (test:reports → ProcessOrder → UpdateInventory)
- [x] TraceTree: human-readable display for command.start/end, schedule.start/end/failed events
- [x] Constants: EVENT_TYPE_COLORS, all event types in EventType enum

---

## Universal Architecture

The Go server is framework-agnostic by design. It accepts a standard JSON event format.
Only the client SDK (PHP package) is framework-specific.

**Future SDKs roadmap:**

```
v0.1  axvi/wiretap-laravel       — Laravel hooks (first release)
v0.2  axvi/wiretap-php            — Generic PHP (manual instrumentation)
v0.3  axvi/wiretap-symfony        — Symfony EventDispatcher + Doctrine
v0.x  axvi/wiretap-python         — Django / FastAPI
v0.x  axvi/wiretap-node           — Express / NestJS
```

Each SDK is a thin adapter: listens to framework events → sends same JSON format to Go.
Go server doesn't know or care which framework sent the data.

This means when designing Go server internals, never hardcode Laravel-specific logic.
All framework-specific display labels (e.g. "Eloquent", "Artisan") come from the SDK
metadata, not from Go server assumptions.

---

## Branding

```
Name:        Wiretap
Tagline:     "See everything. One binary."
Vendor:      axvi

Website:     axvi.dev/wiretap
Docs:        axvi.dev/wiretap/docs
GitHub:      github.com/axvi/wiretap
Docker Hub:  axvi/wiretap
Packagist:   axvi/wiretap-laravel

Ports:
  :8080  — Web UI
  :9090  — Ingest API
  :1025  — SMTP Server

Install:
  docker run -p 8080:8080 -p 9090:9090 -p 1025:1025 axvi/wiretap
  composer require axvi/wiretap-laravel
```

---

## Completed Beyond Original Plan

### Authentication & Security
- [x] Web UI password authentication (bcrypt + persistent sessions in DuckDB)
- [x] Setup flow (first-launch password creation)
- [x] Change password in Settings
- [x] Ingest API key protection (X-Wiretap-Key header)
- [x] Session persistence across server restarts

### Replay/Retry System
- [x] Replay Request (Go direct HTTP, no PHP self-call deadlock)
- [x] Retry Job (queue:retry for failed jobs)
- [x] Dispatch Job (replay any job from serialized payload with editable properties)
- [x] Replay Broadcast (via Laravel broadcast driver)
- [x] Replay Command (Artisan::call)
- [x] ReplayModal with editable forms for all types
- [x] php artisan wiretap:replay-secret command
- [x] Masked field warnings, large payload warnings

### Data Collection Enhancements
- [x] Gzip compression (PHP drivers → Go receiver)
- [x] Job payload capture with masking and serialized replay
- [x] Request body capture with field masking (enabled by default)
- [x] Log listener (MessageLogged event)
- [x] Ignore filters for requests, jobs, commands, broadcasts (wildcard patterns)
- [x] Validation listener fix (via renderable exception handler)
- [x] FormRequest validation support
- [x] Laravel connection warning (Log::warning on ingest failure)

### New UI Pages
- [x] Activity page (root traces only — request, command, schedule)
- [x] Performance page (Top 10 Slow Endpoints + Top 10 Slow Queries from all trace types)
- [x] Logs page with level filter + detail page
- [x] Notifications page with detail page
- [x] Settings tabs (General, Security, Replay, Info)
- [x] Configurable analysis thresholds (slow query/request ms)
- [x] Data retention settings

### UI Improvements
- [x] NewItemsBanner on all table pages (non-intrusive WebSocket updates)
- [x] TraceTree: dump collapsible with preview, log truncated with Open, validation collapsible
- [x] Open buttons on log, exception, notification, broadcast, mail events in trace tree
- [x] Favicon (SVG)
- [x] 13 sidebar pages total

### Testing
- [x] Go tests: 37 (storage, analyzer, receiver gzip, API handlers, auth, sessions, ingest key)
- [x] PHP tests: 24 (Wiretap unit, ignore filters, ReplayController, GenerateReplaySecret)
- [x] Vitest: 28 (API layer, NewItemsBanner, ReplayModal)

### Test Application
- [x] Comprehensive test routes for all 16 listener types
- [x] TestNotification class
- [x] CreateOrderRequest FormRequest
- [x] OrderShipped broadcast event
- [x] ProcessOrder + UpdateInventory job chain
- [x] broadcasting.php config with Redis
