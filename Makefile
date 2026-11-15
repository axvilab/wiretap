# === Environment ===
up:
	docker compose up -d

down:
	docker compose down

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
ui-build:
	docker compose exec frontend npm run build

ui-sh:
	docker compose exec frontend sh

# === Laravel apps ===
app-setup-11:
	docker compose exec laravel-11 composer install
	docker compose exec laravel-11 php artisan key:generate --force
	docker compose exec laravel-11 php artisan migrate --force

app-setup-12:
	docker compose exec laravel-12 composer install
	docker compose exec laravel-12 php artisan key:generate --force
	docker compose exec laravel-12 php artisan migrate --force

app-setup-13:
	docker compose exec laravel-13 composer install
	docker compose exec laravel-13 php artisan key:generate --force
	docker compose exec laravel-13 php artisan migrate --force

app-setup:
	$(MAKE) app-setup-11
	$(MAKE) app-setup-12
	$(MAKE) app-setup-13

app-sh-11:
	docker compose exec laravel-11 bash

app-sh-12:
	docker compose exec laravel-12 bash

app-sh-13:
	docker compose exec laravel-13 bash

# === Traffic (run all test routes) ===
traffic-11:
	docker compose exec laravel-11 bash /var/www/html/run-routes.sh http://localhost:8000

traffic-12:
	docker compose exec laravel-12 bash /var/www/html/run-routes.sh http://localhost:8000

traffic-13:
	docker compose exec laravel-13 bash /var/www/html/run-routes.sh http://localhost:8000

traffic-all:
	@echo "=== Laravel 11 ==="
	$(MAKE) traffic-11
	@echo ""
	@echo "=== Laravel 12 ==="
	$(MAKE) traffic-12
	@echo ""
	@echo "=== Laravel 13 ==="
	$(MAKE) traffic-13

# === Tests ===
php-test:
	docker compose exec laravel-12 composer test --working-dir=/var/www/packages/wiretap

test-all:
	$(MAKE) go-test
	$(MAKE) php-test

# === Release ===
release:
	docker compose exec frontend npm run build
	docker build -t axvi/wiretap ./go-server

.PHONY: up down logs logs-wiretap logs-frontend go-test go-build go-sh ui-build ui-sh app-setup-11 app-setup-12 app-setup-13 app-setup app-sh-11 app-sh-12 app-sh-13 traffic-11 traffic-12 traffic-13 traffic-all php-test test-all release
