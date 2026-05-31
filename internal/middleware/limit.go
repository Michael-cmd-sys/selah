package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/selah/internal/response"
)

// MaxBodyBytes wraps the request body with a size limit on every request.
func MaxBodyBytes(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter is a per-IP sliding-window rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rlEntry
	max     int
	window  time.Duration
}

type rlEntry struct {
	count    int
	reset    time.Time
	lastSeen time.Time
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rlEntry),
		max:     max,
		window:  window,
	}
	go rl.gc()
	return rl
}

func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		ip := r.RemoteAddr

		rl.mu.Lock()
		e, ok := rl.entries[ip]
		if !ok || now.After(e.reset) {
			e = &rlEntry{reset: now.Add(rl.window)}
			rl.entries[ip] = e
		}
		e.count++
		e.lastSeen = now
		allowed := e.count <= rl.max
		rl.mu.Unlock()

		if !allowed {
			response.Error(w, http.StatusTooManyRequests, "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) gc() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, e := range rl.entries {
			if time.Since(e.lastSeen) > 10*time.Minute {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}
