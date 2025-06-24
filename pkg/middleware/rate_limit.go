package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiterWithLastSeen struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	ips     map[string]*limiterWithLastSeen
	mu      *sync.RWMutex
	rate    rate.Limit
	burst   int
	ttl     time.Duration
	cleanup time.Duration
}

func NewIPRateLimiter(r rate.Limit, b int, ttl, cleanup time.Duration) *IPRateLimiter {
	i := &IPRateLimiter{
		ips:     make(map[string]*limiterWithLastSeen),
		mu:      &sync.RWMutex{},
		rate:    r,
		burst:   b,
		ttl:     ttl,
		cleanup: cleanup,
	}

	go i.cleanupLoop()

	return i
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	lwls, exists := i.ips[ip]
	if !exists {
		lwls = &limiterWithLastSeen{
			limiter:  rate.NewLimiter(i.rate, i.burst),
			lastSeen: time.Now(),
		}
		i.ips[ip] = lwls
	} else {
		lwls.lastSeen = time.Now()
	}

	return lwls.limiter
}

func (i *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(i.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		now := time.Now()

		for ip, lwls := range i.ips {
			if now.Sub(lwls.lastSeen) > i.ttl {
				delete(i.ips, ip)
			}
		}
		i.mu.Unlock()
	}
}

func RateLimitMiddleware(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if ip == "" {
				ip = "unknown"
			}

			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				ip = realIP
			}

			lim := limiter.GetLimiter(ip)
			if !lim.Allow() {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
