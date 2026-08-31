package limiter

import (
	"net/http"
	"sync"
	"time"
)

// Option allows configuring the rate Limiter.
type Option func(*Limiter)

// WithOnLimit sets a callback that is executed whenever a client is rate limited.
func WithOnLimit(cb func(ip string, r *http.Request)) Option {
	return func(l *Limiter) {
		l.onLimit = cb
	}
}

// WithCleanupInterval sets a custom pruning interval for inactive client sessions.
func WithCleanupInterval(d time.Duration) Option {
	return func(l *Limiter) {
		l.cleanupInterval = d
	}
}

// clientState tracks rate-limiting tokens for a single client IP.
type clientState struct {
	tokens     float64
	lastRefill time.Time
}

// Limiter is an in-memory token-bucket rate limiter.
type Limiter struct {
	mu              sync.Mutex
	clients         map[string]*clientState
	rate            float64 // tokens refilled per second
	burst           float64 // bucket capacity
	cleanupInterval time.Duration
	onLimit         func(ip string, r *http.Request)
}

// New creates a new Limiter with the specified rate, burst, and options.
func New(rate float64, burst float64, opts ...Option) *Limiter {
	l := &Limiter{
		clients:         make(map[string]*clientState),
		rate:            rate,
		burst:           burst,
		cleanupInterval: 5 * time.Minute,
	}

	for _, opt := range opts {
		opt(l)
	}

	go l.cleanupLoop()
	return l
}

// Allow checks if the client IP is allowed to perform a request.
func (l *Limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	c, exists := l.clients[ip]
	if !exists {
		l.clients[ip] = &clientState{
			tokens:     l.burst - 1,
			lastRefill: now,
		}
		return true
	}

	elapsed := now.Sub(c.lastRefill).Seconds()
	c.lastRefill = now
	c.tokens += elapsed * l.rate
	if c.tokens > l.burst {
		c.tokens = l.burst
	}

	if c.tokens >= 1 {
		c.tokens -= 1
		return true
	}
	return false
}

// cleanupLoop periodically prunes inactive clients to avoid memory leaks.
func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanupInterval)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for ip, c := range l.clients {
			if now.Sub(c.lastRefill) > 15*time.Minute {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}
