# === Environment ===
up:
	docker compose up -d wiretap

up-all:
	docker compose --profile ui --profile laravel up -d

down:
	docker compose --profile ui --profile laravel down

logs:
	docker compose logs -f

logs-wiretap:
	docker compose logs -f wiretap

logs-frontend:
	docker compose logs -f frontend

# === Go server ===
go-test:
	docker compose exec wiretap go test ./...

go-build:
	docker compose exec wiretap go build -o wiretap ./cmd/wiretap/

go-sh:
	docker compose exec wiretap sh

# === Frontend ===
ui-up:
	docker compose --profile ui up -d frontend

ui-build:
	docker compose exec frontend npm run build

ui-sh:
	docker compose exec frontend sh

# === Laravel (Phase 2+) ===
laravel-up:
	docker compose --profile laravel up -d

app-setup:
	docker compose exec laravel composer install
	docker compose exec laravel php artisan migrate
	docker compose exec laravel php artisan db:seed

app-seed-traffic:
	docker compose exec laravel php artisan wiretap:generate-traffic

app-sh:
	docker compose exec laravel bash

php-test:
	docker compose exec laravel composer test --working-dir=/var/www/packages/wiretap

# === Release ===
release:
	docker compose exec frontend npm run build
	docker build -t axvi/wiretap ./go-server

.PHONY: up up-all down logs logs-wiretap logs-frontend go-test go-build go-sh ui-up ui-build ui-sh laravel-up app-setup app-seed-traffic app-sh php-test release
