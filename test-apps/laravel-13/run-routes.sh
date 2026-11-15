#!/bin/bash
# Run all test routes against a Laravel test-app
# Usage: ./run-routes.sh http://localhost:8000

BASE="${1:-http://localhost:8000}"

echo "=== Running routes against $BASE ==="

# GET routes
for path in / /view-demo /api/users /api/posts /api/posts-n-plus-one /api/slow /api/cache-demo /api/transaction-demo /api/gate-demo /api/dump-demo /api/kitchen-sink /api/model-demo /api/health; do
    code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE$path" 2>/dev/null)
    echo "GET  $path => $code"
done

# POST routes
for path in /api/orders /api/broadcast /api/mail-demo /api/notification-demo /api/auth-demo; do
    code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE$path" 2>/dev/null)
    echo "POST $path => $code"
done

# Validation errors (expected 422)
code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/validate-demo" -H "Accept: application/json" -H "Content-Type: application/json" -d '{}' 2>/dev/null)
echo "POST /api/validate-demo => $code (expected 422)"

code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/form-request-demo" -H "Accept: application/json" -H "Content-Type: application/json" -d '{}' 2>/dev/null)
echo "POST /api/form-request-demo => $code (expected 422)"

# Exception (expected 500)
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/error" -H "Accept: application/json" 2>/dev/null)
echo "GET  /api/error => $code (expected 500)"

echo "=== Done ==="
