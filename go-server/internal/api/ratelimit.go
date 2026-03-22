package api

import (
	"net/http"
	"sync"
	"time"
)

// maxBodySize limits the request body size for all API routes.
func maxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// rateLimiter limits write operations per IP.
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int           // max requests
	window   time.Duration // per time window
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old entries
	times := rl.requests[ip]
	clean := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			clean = append(clean, t)
		}
	}

	if len(clean) >= rl.limit {
		rl.requests[ip] = clean
		return false
	}

	rl.requests[ip] = append(clean, now)
	return true
}

// writeRateLimit wraps handlers that create/update/delete with rate limiting.
func writeRateLimit(rl *rateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = fwd
		}
		if !rl.allow(ip) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
