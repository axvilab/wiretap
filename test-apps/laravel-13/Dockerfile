FROM php:8.3-cli

RUN apt-get update && apt-get install -y \
    git unzip libzip-dev libcurl4-openssl-dev libpng-dev libonig-dev \
    default-mysql-client \
    && docker-php-ext-install pdo_mysql zip curl mbstring pcntl \
    && pecl install redis && docker-php-ext-enable redis \
    && rm -rf /var/lib/apt/lists/*

COPY --from=composer:2 /usr/bin/composer /usr/bin/composer

WORKDIR /var/www/html

EXPOSE 8000

CMD bash -c "composer install --quiet 2>/dev/null; php artisan key:generate --force 2>/dev/null; php artisan migrate --force 2>/dev/null; php artisan serve --host=0.0.0.0 --port=8000"
